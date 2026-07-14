package tests

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

func TestAuditLogAPICapturesBackofficeMutationsAndExcludesBookings(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/audit-logs", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)

	userPassword := "MemberPassword123!"
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "fullName": "Member One", "password": userPassword, "timezone": "UTC",
	}), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{
		"fullName": "Member Two", "password": "ChangedPassword123!",
	}), http.StatusOK)

	eventType := eventTypeBody("Audit Call", []string{adminEmail}, 2)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventType), http.StatusCreated)
	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	createdBooking := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": "audit-call", "time": bookingTime, "attendeeName": "Guest", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC",
	}), http.StatusCreated)
	var bookingResponse struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}
	if err := json.Unmarshal(createdBooking, &bookingResponse); err != nil {
		t.Fatal(err)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/bookings/"+bookingResponse.Booking.ID, nil), http.StatusNoContent)

	updatedEventType := eventTypeBody("Renamed Audit Call", []string{adminEmail}, 2)
	expectStatus(t, f.request(http.MethodPut, "/api/event-types/audit-call", updatedEventType), http.StatusOK)
	expectStatus(t, f.request(http.MethodPut, "/api/branding", map[string]string{"name": "Audited Calls"}), http.StatusOK)

	createdToken := expectStatus(t, f.request(http.MethodPost, "/api/integration/tokens", map[string]string{"name": "Audit client"}), http.StatusCreated)
	var tokenResponse struct {
		APIToken struct {
			ID string `json:"id"`
		} `json:"apiToken"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(createdToken, &tokenResponse); err != nil {
		t.Fatal(err)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/integration/tokens/"+tokenResponse.APIToken.ID, nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/event-types/audit-call", nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)

	body := expectStatus(t, f.request(http.MethodGet, "/api/audit-logs", nil), http.StatusOK)
	if strings.Contains(string(body), userPassword) || strings.Contains(string(body), "ChangedPassword123!") || strings.Contains(string(body), tokenResponse.Token) || strings.Contains(string(body), "passwordHash") {
		t.Fatalf("audit logs contain a secret: %s", body)
	}
	var response struct {
		AuditLogs []model.AuditLog `json:"auditLogs"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.AuditLogs) != 9 {
		t.Fatalf("audit log count = %d, want 9: %s", len(response.AuditLogs), body)
	}
	actions := make(map[string]int)
	for _, entry := range response.AuditLogs {
		actions[entry.Action]++
		if entry.Actor.Email != adminEmail || entry.CreatedAt.Location() != time.UTC || entry.CreatedAt.Nanosecond() != 0 {
			t.Fatalf("invalid audit actor or timestamp: %#v", entry)
		}
	}
	for action, expected := range map[string]int{
		"created": 2, "edited": 3, "generated_token": 1, "revoked_token": 1, "deleted": 2,
	} {
		if actions[action] != expected {
			t.Fatalf("action %q count = %d, want %d", action, actions[action], expected)
		}
	}
	assertAuditPayloadContains(t, response.AuditLogs, "edited", "user", `"fullName":{"before":"Member One","after":"Member Two"}`)
	assertAuditPayloadContains(t, response.AuditLogs, "edited", "event_type", `"name":{"before":"Audit Call","after":"Renamed Audit Call"}`)
}

func TestAuditLogRetentionAndTableRecreation(t *testing.T) {
	dataPath := t.TempDir()
	database, err := store.Open(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	startedAt := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	for index, id := range []string{"first", "second", "third"} {
		if err := database.AppendAuditLog(model.AuditLog{
			ID: id, Actor: model.AuditLogActor{Email: adminEmail}, Action: "edited", Resource: "branding",
			ResourceID: "current", CreatedAt: startedAt.Add(time.Duration(index) * time.Second), Payload: json.RawMessage(`{}`),
		}, 2); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := database.ListAuditLogs()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].ID != "third" || entries[1].ID != "second" {
		t.Fatalf("unexpected retained audit logs: %#v", entries)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	tablePath := filepath.Join(dataPath, store.AuditLogsTableName+".leveldb")
	if _, err := os.Stat(tablePath); err != nil {
		t.Fatalf("predictable audit LevelDB table was not created: %v", err)
	}
	if err := os.RemoveAll(tablePath); err != nil {
		t.Fatal(err)
	}
	database, err = store.Open(dataPath)
	if err != nil {
		t.Fatalf("reopen store after audit table removal: %v", err)
	}
	defer database.Close()
	entries, err = database.ListAuditLogs()
	if err != nil || len(entries) != 0 {
		t.Fatalf("recreated audit table should be empty: entries=%#v err=%v", entries, err)
	}
}

func assertAuditPayloadContains(t *testing.T, entries []model.AuditLog, action, resource, expected string) {
	t.Helper()
	for _, entry := range entries {
		if entry.Action == action && entry.Resource == resource && strings.Contains(string(entry.Payload), expected) {
			return
		}
	}
	t.Fatalf("no %s %s audit payload contains %s", action, resource, expected)
}
