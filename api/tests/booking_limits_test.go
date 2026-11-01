package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/security"
)

func TestBookingRejectsMoreThanMaxGuestEmails(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Group Call", []string{adminEmail}, 20)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	booking := func(guestCount int) map[string]any {
		guests := make([]string, guestCount)
		for i := range guests {
			guests[i] = fmt.Sprintf("guest%d@example.com", i)
		}
		return map[string]any{
			"eventSlug": "group-call", "time": bookingTime, "attendeeName": "Guest Person", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC", "guestEmails": guests,
		}
	}
	rejected := expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking(11)), http.StatusBadRequest)
	if !strings.Contains(string(rejected), "guestEmails must not exceed 10 addresses") {
		t.Fatalf("unexpected guest cap error: %s", rejected)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking(10)), http.StatusCreated)
}

func TestBookingMutationsAreRateLimitedPerClientIP(t *testing.T) {
	f := newFixture(t, false)
	missing := map[string]string{"eventSlug": "missing"}
	for range 20 {
		expectStatus(t, f.request(http.MethodPost, "/api/bookings", missing), http.StatusNotFound)
	}
	limited := expectStatus(t, f.request(http.MethodPost, "/api/bookings", missing), http.StatusTooManyRequests)
	if !strings.Contains(string(limited), "too many booking requests") {
		t.Fatalf("unexpected rate limit error: %s", limited)
	}
	expectStatus(t, f.request(http.MethodPatch, "/api/events/missing", map[string]string{"notes": ""}), http.StatusTooManyRequests)
	expectStatus(t, f.request(http.MethodPost, "/api/events/missing/cancel", map[string]string{"reason": ""}), http.StatusTooManyRequests)
	expectStatus(t, f.request(http.MethodGet, "/api/events/missing", nil), http.StatusNotFound)
	expectStatus(t, bookingAttemptFromIP(f, "203.0.113.7"), http.StatusNotFound)
}

func TestRateLimiterSlidingWindow(t *testing.T) {
	limiter := security.NewRateLimiter(2, 10*time.Minute)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	steps := []struct {
		at      time.Duration
		allowed bool
	}{
		{0, true},
		{time.Minute, true},
		{2 * time.Minute, false},
		{10 * time.Minute, true},
		{10*time.Minute + time.Second, false},
	}
	for i, step := range steps {
		if got := limiter.Allow("ip", start.Add(step.at)); got != step.allowed {
			t.Fatalf("step %d: allowed %v, want %v", i, got, step.allowed)
		}
	}
	if !limiter.Allow("other", start.Add(2*time.Minute)) {
		t.Fatal("other key was limited")
	}
}

func bookingAttemptFromIP(f *fixture, ip string) *http.Response {
	f.t.Helper()
	request, err := http.NewRequest(http.MethodPost, f.server.URL+"/api/bookings", strings.NewReader(`{"eventSlug":"missing"}`))
	if err != nil {
		f.t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("CF-Connecting-IP", ip)
	return mustDo(f.t, f.client, request)
}
