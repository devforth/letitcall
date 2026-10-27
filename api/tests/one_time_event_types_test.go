package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOneTimeEventTypeAcceptsOneActiveBooking(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	regular := expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Regular Call", []string{adminEmail}, 1)), http.StatusCreated)
	if !strings.Contains(string(regular), `"oneTime":false`) {
		t.Fatalf("event types without oneTime must be regular: %s", regular)
	}
	oneTime := eventTypeBody("Single Call", []string{adminEmail}, 1)
	oneTime["oneTime"] = true
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", oneTime), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	day := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, time.UTC)
	book := func(hour int, email string, status int) []byte {
		return expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
			"eventSlug": "single-call", "time": day.Add(time.Duration(hour) * time.Hour).Format(time.RFC3339),
			"attendeeName": "Guest", "attendeeEmail": email, "attendeeTimezone": "UTC",
		}), status)
	}
	expectBooked := func(booked bool) {
		t.Helper()
		publicEventType := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/single-call", nil), http.StatusOK)
		if !strings.Contains(string(publicEventType), `"booked":`+map[bool]string{true: "true", false: "false"}[booked]) {
			t.Fatalf("public event type booked state is not %v: %s", booked, publicEventType)
		}
	}

	expectBooked(false)
	created := book(10, "first@example.com", http.StatusCreated)
	expectBooked(true)
	if rejected := book(12, "second@example.com", http.StatusConflict); !strings.Contains(string(rejected), "this one-time meeting is already booked") {
		t.Fatalf("unexpected second booking error: %s", rejected)
	}

	var creation struct {
		ManageURL string `json:"manageURL"`
	}
	if err := json.Unmarshal(created, &creation); err != nil {
		t.Fatal(err)
	}
	manageURL, err := url.Parse(creation.ManageURL)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/events/"+strings.TrimPrefix(manageURL.Path, "/event/")+"/cancel", map[string]string{"reason": ""}), http.StatusOK)
	expectBooked(false)
	book(12, "second@example.com", http.StatusCreated)
	expectBooked(true)
}
