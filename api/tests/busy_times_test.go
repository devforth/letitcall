package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
)

func TestBusyTimeAPIsBlockRequiredHostBookings(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/busy-times", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Busy Call", []string{adminEmail}, 1)), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	start := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC)
	busyTime := map[string]any{"userEmail": adminEmail, "start": start.Format(time.RFC3339), "durationMinutes": 60, "comment": " Dentist "}
	for field, value := range map[string]any{"userEmail": "missing@example.com", "start": "2026-10-28T12:00:00+02:00", "durationMinutes": 0} {
		invalid := map[string]any{}
		for key, original := range busyTime {
			invalid[key] = original
		}
		invalid[field] = value
		expectStatus(t, f.request(http.MethodPost, "/api/busy-times", invalid), http.StatusBadRequest)
	}
	created := expectStatus(t, f.request(http.MethodPost, "/api/busy-times", busyTime), http.StatusCreated)
	var creation struct {
		BusyTime model.BusyTime `json:"busyTime"`
	}
	if err := json.Unmarshal(created, &creation); err != nil {
		t.Fatal(err)
	}
	if creation.BusyTime.Comment != "Dentist" || creation.BusyTime.DurationMinutes != 60 || !creation.BusyTime.Start.Equal(start) {
		t.Fatalf("unexpected busy time: %s", created)
	}
	listed := expectStatus(t, f.request(http.MethodGet, "/api/busy-times", nil), http.StatusOK)
	if !strings.Contains(string(listed), creation.BusyTime.ID) {
		t.Fatalf("busy time missing from list: %s", listed)
	}
	publicEventType := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/busy-call", nil), http.StatusOK)
	busyRange := `{"start":"` + start.Format(time.RFC3339) + `","end":"` + start.Add(time.Hour).Format(time.RFC3339) + `"}`
	if !strings.Contains(string(publicEventType), busyRange) {
		t.Fatalf("busy time was not exposed as unavailable: %s", publicEventType)
	}

	booking := func(at time.Time, email string) map[string]any {
		return map[string]any{
			"eventSlug": "busy-call", "time": at.Format(time.RFC3339), "attendeeName": "Guest", "attendeeEmail": email, "attendeeTimezone": "UTC",
		}
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking(start.Add(30*time.Minute), "first@example.com")), http.StatusConflict)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking(start.Add(time.Hour), "second@example.com")), http.StatusCreated)

	expectStatus(t, f.request(http.MethodDelete, "/api/busy-times/"+creation.BusyTime.ID, nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/busy-times/"+creation.BusyTime.ID, nil), http.StatusNotFound)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking(start.Add(30*time.Minute), "first@example.com")), http.StatusCreated)
	auditLogs := expectStatus(t, f.request(http.MethodGet, "/api/audit-logs", nil), http.StatusOK)
	if strings.Count(string(auditLogs), `"resource":"busy_time"`) != 2 {
		t.Fatalf("busy time changes were not audited: %s", auditLogs)
	}
}

func TestBusyTimeIgnoresOptionalHostsAndIsDeletedWithUser(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{"email": "member@example.com", "timezone": "UTC"}), http.StatusCreated)
	eventType := eventTypeBody("Optional Call", []string{adminEmail}, 1)
	eventType["optionalHostEmails"] = []string{"member@example.com"}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventType), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	start := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	expectStatus(t, f.request(http.MethodPost, "/api/busy-times", map[string]any{
		"userEmail": "MEMBER@example.com", "start": start, "durationMinutes": 30, "comment": "",
	}), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": "optional-call", "time": start, "attendeeName": "Guest", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC",
	}), http.StatusCreated)

	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)
	listed := expectStatus(t, f.request(http.MethodGet, "/api/busy-times", nil), http.StatusOK)
	if string(listed) != "{\"busyTimes\":[]}\n" {
		t.Fatalf("busy times of the deleted user were kept: %s", listed)
	}
}
