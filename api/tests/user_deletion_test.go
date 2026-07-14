package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
)

func TestUserDeletionImpactAndBookingReassignment(t *testing.T) {
	f := newFixture(t, true)
	const memberEmail = "member@example.com"
	const memberPassword = "MemberPassword123!"

	expectStatus(t, f.request(http.MethodGet, "/api/users/"+memberEmail+"/deletion-impact", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": memberEmail, "password": memberPassword, "timezone": "UTC",
	}), http.StatusCreated)

	expectStatus(t, f.login(memberEmail, memberPassword), http.StatusOK)
	mockGoogleProvider(t, memberEmail, "Member Host", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	mockGoogleProvider(t, adminEmail, "Admin Host", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)

	calendarCreates, calendarDeletes := mockHostReassignmentCalendar(t)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Member meetings", []string{memberEmail}, 1)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Member without bookings", []string{memberEmail}, 1)), http.StatusCreated)

	firstStart := futureBookingTime(2, 12)
	lastStart := futureBookingTime(4, 15)
	firstID := createDeletionTestBooking(t, f, "member-meetings", firstStart, "first@example.com")
	lastID := createDeletionTestBooking(t, f, "member-meetings", lastStart, "last@example.com")

	now := time.Now().UTC().Truncate(time.Second)
	canceledAt := now
	limit := 1
	for _, booking := range []model.Booking{
		{
			ID: "past", EventSlug: "member-meetings", Time: now.Add(-24 * time.Hour), EndTime: now.Add(-23*time.Hour - 30*time.Minute),
			AttendeeName: "Past", AttendeeEmail: "past@example.com", RecipientEmails: []string{memberEmail}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "canceled", EventSlug: "member-meetings", Time: now.Add(72 * time.Hour), EndTime: now.Add(72*time.Hour + 30*time.Minute),
			AttendeeName: "Canceled", AttendeeEmail: "canceled@example.com", RecipientEmails: []string{memberEmail}, CanceledAt: &canceledAt, CreatedAt: now, UpdatedAt: now,
		},
	} {
		key := booking.EventSlug + booking.Time.Format(time.RFC3339) + "-" + booking.EndTime.Format(time.RFC3339)
		if err := f.store.CreateBooking(key, booking, []string{memberEmail}, &limit); err != nil {
			t.Fatal(err)
		}
	}

	var impact struct {
		RequiresReassignment bool       `json:"requiresReassignment"`
		FutureBookingCount   int        `json:"futureBookingCount"`
		EarliestBookingAt    *time.Time `json:"earliestBookingAt"`
		LatestBookingAt      *time.Time `json:"latestBookingAt"`
	}
	body := expectStatus(t, f.request(http.MethodGet, "/api/users/"+memberEmail+"/deletion-impact", nil), http.StatusOK)
	if err := json.Unmarshal(body, &impact); err != nil {
		t.Fatal(err)
	}
	if !impact.RequiresReassignment || impact.FutureBookingCount != 2 || impact.EarliestBookingAt == nil || !impact.EarliestBookingAt.Equal(firstStart) || impact.LatestBookingAt == nil || !impact.LatestBookingAt.Equal(lastStart) {
		t.Fatalf("unexpected deletion impact: %s", body)
	}

	expectStatus(t, f.request(http.MethodPost, "/api/users/"+memberEmail+"/reassign-bookings", map[string]string{
		"newHostEmail": memberEmail,
	}), http.StatusBadRequest)
	reassigned := expectStatus(t, f.request(http.MethodPost, "/api/users/"+memberEmail+"/reassign-bookings", map[string]string{
		"newHostEmail": adminEmail,
	}), http.StatusOK)
	if !strings.Contains(string(reassigned), `"reassignedBookingCount":2`) {
		t.Fatalf("unexpected reassignment response: %s", reassigned)
	}
	if *calendarCreates != 4 || *calendarDeletes != 2 {
		t.Fatalf("unexpected Google Calendar changes: creates=%d deletes=%d", *calendarCreates, *calendarDeletes)
	}
	for _, id := range []string{firstID, lastID} {
		booking, err := f.store.GetBooking(id)
		if err != nil || len(booking.RecipientEmails) != 1 || booking.RecipientEmails[0] != adminEmail || booking.GoogleEventIDs[adminEmail] == "" || booking.GoogleEventIDs[memberEmail] != "" {
			t.Fatalf("booking was not reassigned: booking=%#v err=%v", booking, err)
		}
	}
	for _, slug := range []string{"member-meetings", "member-without-bookings"} {
		eventType, err := f.store.GetEventType(slug)
		if err != nil || len(eventType.RequiredHostEmails) != 1 || eventType.RequiredHostEmails[0] != adminEmail {
			t.Fatalf("event type host was not reassigned: eventType=%#v err=%v", eventType, err)
		}
	}

	noImpact := expectStatus(t, f.request(http.MethodGet, "/api/users/"+memberEmail+"/deletion-impact", nil), http.StatusOK)
	if !strings.Contains(string(noImpact), `"requiresReassignment":false`) || !strings.Contains(string(noImpact), `"futureBookingCount":0`) {
		t.Fatalf("reassigned user still has deletion impact: %s", noImpact)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/users/"+memberEmail, nil), http.StatusNoContent)
}

func futureBookingTime(days, hour int) time.Time {
	candidate := time.Now().UTC().AddDate(0, 0, days)
	return time.Date(candidate.Year(), candidate.Month(), candidate.Day(), hour, 0, 0, 0, time.UTC)
}

func createDeletionTestBooking(t *testing.T, f *fixture, slug string, start time.Time, attendeeEmail string) string {
	t.Helper()
	body := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": slug, "time": start.Format(time.RFC3339), "attendeeName": "Guest", "attendeeEmail": attendeeEmail, "attendeeTimezone": "UTC",
	}), http.StatusCreated)
	var response struct {
		Booking model.Booking `json:"booking"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	return response.Booking.ID
}

func mockHostReassignmentCalendar(t *testing.T) (*int, *int) {
	t.Helper()
	originalTransport := http.DefaultTransport
	creates := 0
	deletes := 0
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.Method {
		case http.MethodGet:
			return jsonHTTPResponse(request, `{"timeZone":"UTC","items":[]}`), nil
		case http.MethodPost:
			creates++
			body, _ := io.ReadAll(request.Body)
			if request.URL.Query().Get("sendUpdates") != "all" || !bytes.Contains(body, []byte("Cancel or update event")) {
				t.Errorf("Google Calendar event was incomplete: url=%s body=%s", request.URL, body)
			}
			return jsonHTTPResponse(request, fmt.Sprintf(`{"id":"calendar-%d"}`, creates)), nil
		case http.MethodDelete:
			deletes++
			if request.URL.Query().Get("sendUpdates") != "all" || !strings.HasPrefix(request.URL.Path, "/calendar/v3/calendars/primary/events/calendar-") {
				t.Errorf("Google Calendar deletion was incomplete: %s", request.URL)
			}
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Status:     http.StatusText(http.StatusNoContent),
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    request,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected Google Calendar request: %s %s", request.Method, request.URL)
		}
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })
	return &creates, &deletes
}

func jsonHTTPResponse(request *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}
