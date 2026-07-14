package tests

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/httpapi"
	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	adminEmail    = "admin@example.com"
	adminPassword = "AdminPassword123!"
)

type fixture struct {
	t        *testing.T
	api      *httpapi.Server
	server   *httptest.Server
	client   *http.Client
	store    *store.Store
	basePath string
	dataPath string
}

func newFixture(t *testing.T, googleEnabled bool) *fixture {
	return newFixtureAtBasePath(t, googleEnabled, "")
}

func newFixtureAtBasePath(t *testing.T, googleEnabled bool, basePath string) *fixture {
	return newConfiguredFixture(t, googleEnabled, basePath, nil)
}

func newConfiguredFixture(t *testing.T, googleEnabled bool, basePath string, configure func(*config.Config)) *fixture {
	t.Helper()
	dataPath := t.TempDir()
	database, err := store.Open(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	user, err := security.NewUser(adminEmail, adminPassword, "UTC", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig(dataPath)
	cfg.HTTP.BaseURL = "http://example.test" + basePath
	if googleEnabled {
		cfg.Login.Google = config.GoogleOAuth{
			ClientID:     "google-client-id",
			ClientSecret: "google-client-secret",
		}
	}
	if configure != nil {
		configure(&cfg)
	}
	handler, err := httpapi.New(cfg, database)
	if err != nil {
		database.Close()
		t.Fatal(err)
	}
	testServer := httptest.NewServer(handler.Handler())
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := testServer.Client()
	client.Jar = jar
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	result := &fixture{t: t, api: handler, server: testServer, client: client, store: database, basePath: basePath, dataPath: dataPath}
	t.Cleanup(func() {
		testServer.Close()
		if err := database.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	return result
}

func testConfig(dataPath string) config.Config {
	return config.Config{
		HTTP:    config.HTTP{Port: 80, BaseURL: config.DefaultBaseURL},
		Storage: config.Storage{LevelDBPath: dataPath},
		Login: config.Login{
			SessionTTL:          time.Hour,
			PasswordMaxAttempts: 20,
			PasswordLockout:     time.Minute,
		},
		AuditLog: config.AuditLog{MaxItems: config.DefaultAuditLogMaxItems},
	}
}

func (f *fixture) request(method, path string, body any) *http.Response {
	f.t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			f.t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, f.server.URL+f.basePath+path, reader)
	if err != nil {
		f.t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := f.client.Do(request)
	if err != nil {
		f.t.Fatal(err)
	}
	return response
}

func (f *fixture) login(email, password string) *http.Response {
	f.t.Helper()
	return f.request(http.MethodPost, "/api/auth/login", map[string]string{
		"email": email, "password": password,
	})
}

func expectStatus(t *testing.T, response *http.Response, expected int) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, response.StatusCode, body)
	}
	return body
}

func TestPublicAndUnknownEndpoints(t *testing.T) {
	f := newFixture(t, false)
	health := expectStatus(t, f.request(http.MethodGet, "/api/health", nil), http.StatusOK)
	if !strings.Contains(string(health), `"status":"ok"`) {
		t.Fatalf("unexpected health body: %s", health)
	}
	publicConfig := expectStatus(t, f.request(http.MethodGet, "/api/config/public", nil), http.StatusOK)
	if !strings.Contains(string(publicConfig), `"googleLoginEnabled":false`) || !strings.Contains(string(publicConfig), `"brandName":"Let It Call"`) {
		t.Fatalf("unexpected public config: %s", publicConfig)
	}
	expectStatus(t, f.request(http.MethodGet, "/api/does-not-exist", nil), http.StatusNotFound)
	portal := expectStatus(t, f.request(http.MethodGet, "/users", nil), http.StatusOK)
	if !strings.Contains(string(portal), "Let It Call") {
		t.Fatalf("SPA fallback did not serve the portal: %s", portal)
	}
}

func TestPublicConfigAndPortalUseStoredBrandName(t *testing.T) {
	f := newFixture(t, false)
	if err := f.store.PutBranding(model.Branding{Name: "DevForth"}); err != nil {
		t.Fatal(err)
	}
	publicConfig := expectStatus(t, f.request(http.MethodGet, "/api/config/public", nil), http.StatusOK)
	if !strings.Contains(string(publicConfig), `"brandName":"DevForth"`) {
		t.Fatalf("public config did not include the brand: %s", publicConfig)
	}
	portal := expectStatus(t, f.request(http.MethodGet, "/", nil), http.StatusOK)
	if !strings.Contains(string(portal), "DevForth") || strings.Contains(string(portal), "Let It Call") {
		t.Fatalf("portal fallback did not use the brand: %s", portal)
	}
}

func TestBrandingAPIsStoreAndServeLogo(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/branding", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)

	initial := expectStatus(t, f.request(http.MethodGet, "/api/branding", nil), http.StatusOK)
	if !strings.Contains(string(initial), `"name":"Let It Call"`) {
		t.Fatalf("unexpected initial branding: %s", initial)
	}
	updated := expectStatus(t, f.request(http.MethodPut, "/api/branding", map[string]string{
		"name": "DevForth", "logo": jpegDataURL(t, 512, 512),
	}), http.StatusOK)
	logoFilename := logoFilenameFromResponse(t, updated)
	branding, err := f.store.GetBranding()
	if err != nil || branding.Name != "DevForth" || branding.LogoPath != logoFilename {
		t.Fatalf("branding was not stored: branding=%#v err=%v", branding, err)
	}
	if _, err := os.Stat(filepath.Join(f.dataPath, "branding.leveldb")); err != nil {
		t.Fatalf("branding LevelDB was not created: %v", err)
	}
	stored, err := os.ReadFile(filepath.Join(f.dataPath, "content", "logos", logoFilename))
	if err != nil || !bytes.Equal(stored, jpegBytes(t, 512, 512)) {
		t.Fatalf("logo JPEG was not stored: %v", err)
	}
	served := f.request(http.MethodGet, "/content/logos/"+logoFilename, nil)
	servedBody := expectStatus(t, served, http.StatusOK)
	if served.Header.Get("Content-Type") != "image/jpeg" || !bytes.Equal(servedBody, stored) {
		t.Fatal("stored logo was not served as a JPEG")
	}
	publicConfig := expectStatus(t, f.request(http.MethodGet, "/api/config/public", nil), http.StatusOK)
	if !strings.Contains(string(publicConfig), `"brandName":"DevForth"`) || !strings.Contains(string(publicConfig), `"logoPath":"`+logoFilename+`"`) {
		t.Fatalf("public config did not include stored branding: %s", publicConfig)
	}

	reuploaded := expectStatus(t, f.request(http.MethodPut, "/api/branding", map[string]string{
		"name": "DevForth", "logo": jpegDataURL(t, 512, 512),
	}), http.StatusOK)
	secondLogoFilename := logoFilenameFromResponse(t, reuploaded)
	if secondLogoFilename == logoFilename {
		t.Fatal("re-uploaded logo reused the previous filename")
	}
	if _, err := os.Stat(filepath.Join(f.dataPath, "content", "logos", logoFilename)); !os.IsNotExist(err) {
		t.Fatalf("previous logo was not removed: %v", err)
	}
	expectStatus(t, f.request(http.MethodGet, "/content/logos/not-a-logo.txt", nil), http.StatusNotFound)
}

func TestBrandingAPIValidatesNameAndLogo(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPut, "/api/branding", map[string]string{"name": " "}), http.StatusBadRequest)
	expectStatus(t, f.request(http.MethodPut, "/api/branding", map[string]string{
		"name": "DevForth", "logo": jpegDataURL(t, 64, 64),
	}), http.StatusBadRequest)
	branding, err := f.store.GetBranding()
	if err != nil || branding.Name != model.DefaultBrandName || branding.LogoPath != "" {
		t.Fatalf("invalid branding update was stored: branding=%#v err=%v", branding, err)
	}
}

func TestBasePathMountsAPIAndPortal(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/letitcall")
	expectStatus(t, f.request(http.MethodGet, "/api/health", nil), http.StatusOK)
	expectStatus(t, f.request(http.MethodGet, "/users", nil), http.StatusOK)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodGet, "/api/auth/session", nil), http.StatusOK)
	request, err := http.NewRequest(http.MethodGet, f.server.URL+"/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, mustDo(t, f.client, request), http.StatusNotFound)
}

func TestAuthenticationAPIs(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/auth/session", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, "incorrect-password"), http.StatusUnauthorized)
	loginBody := expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	if strings.Contains(string(loginBody), "passwordHash") {
		t.Fatal("login response exposed a password hash")
	}
	expectStatus(t, f.request(http.MethodGet, "/api/auth/session", nil), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodGet, "/api/auth/session", nil), http.StatusUnauthorized)
}

func TestFirstUserRequiresEmail(t *testing.T) {
	if _, err := security.NewFirstUser("admin", "admin", time.Now()); err == nil || !strings.Contains(err.Error(), "email must be a valid address") {
		t.Fatalf("expected invalid first-user email error, got %v", err)
	}
}

func TestAuthenticationRejectsInvalidMediaTypeAndOrigin(t *testing.T) {
	f := newFixture(t, false)
	request, err := http.NewRequest(http.MethodPost, f.server.URL+"/api/auth/login", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, mustDo(t, f.client, request), http.StatusUnsupportedMediaType)

	request, err = http.NewRequest(http.MethodPost, f.server.URL+"/api/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"AdminPassword123!"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://attacker.example")
	expectStatus(t, mustDo(t, f.client, request), http.StatusForbidden)
}

func TestUserManagementAPIs(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/users", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	users := expectStatus(t, f.request(http.MethodGet, "/api/users", nil), http.StatusOK)
	if !strings.Contains(string(users), adminEmail) || strings.Contains(string(users), "passwordHash") {
		t.Fatalf("unexpected users response: %s", users)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member", "password": "MemberPassword123!", "timezone": "UTC",
	}), http.StatusBadRequest)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "short", "timezone": "UTC",
	}), http.StatusBadRequest)

	createBody := map[string]string{
		"email": "member@example.com", "fullName": "Ada Lovelace", "password": "MemberPassword123!", "timezone": "Europe/London",
	}
	created := expectStatus(t, f.request(http.MethodPost, "/api/users", createBody), http.StatusCreated)
	if !strings.Contains(string(created), `"fullName":"Ada Lovelace"`) || !strings.Contains(string(created), `"googleConnected":false`) {
		t.Fatalf("unexpected create response: %s", created)
	}
	storedUser, err := f.store.GetUser("member@example.com")
	if err != nil || storedUser.FullName != "Ada Lovelace" {
		t.Fatalf("full name was not stored in LevelDB: user=%#v err=%v", storedUser, err)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/users", createBody), http.StatusConflict)
	emailOnly := expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{"email": "google-only@example.com"}), http.StatusCreated)
	if !strings.Contains(string(emailOnly), `"fullName":""`) || !strings.Contains(string(emailOnly), `"timezone":"UTC"`) {
		t.Fatalf("unexpected email-only user response: %s", emailOnly)
	}
	storedEmailOnly, err := f.store.GetUser("google-only@example.com")
	if err != nil || storedEmailOnly.PasswordHash != "" || storedEmailOnly.Timezone != "UTC" {
		t.Fatalf("email-only user was not stored correctly: user=%#v err=%v", storedEmailOnly, err)
	}
	expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"timezone": "America/New_York"}), http.StatusOK)
	profileUpdated := expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"fullName": "Grace Hopper"}), http.StatusOK)
	if !strings.Contains(string(profileUpdated), `"fullName":"Grace Hopper"`) {
		t.Fatalf("unexpected profile update response: %s", profileUpdated)
	}
	updated := expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"avatar": jpegDataURL(t, 512, 512)}), http.StatusOK)
	firstAvatarFilename := avatarFilenameFromResponse(t, updated, "member__example.com")
	if _, err := os.Stat(filepath.Join(f.dataPath, "content", "avatars", firstAvatarFilename)); err != nil {
		t.Fatalf("updated avatar file was not stored: %v", err)
	}
	reuploaded := expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"avatar": jpegDataURL(t, 512, 512)}), http.StatusOK)
	secondAvatarFilename := avatarFilenameFromResponse(t, reuploaded, "member__example.com")
	if secondAvatarFilename == firstAvatarFilename {
		t.Fatal("re-uploaded avatar reused the previous filename")
	}
	if _, err := os.Stat(filepath.Join(f.dataPath, "content", "avatars", firstAvatarFilename)); !os.IsNotExist(err) {
		t.Fatalf("previous avatar file was not removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.dataPath, "content", "avatars", secondAvatarFilename)); err != nil {
		t.Fatalf("re-uploaded avatar file was not stored: %v", err)
	}
	expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"email": "renamed@example.com"}), http.StatusBadRequest)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/admin@example.com", nil), http.StatusConflict)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNotFound)
}

func TestUserAvatarStorageAndServing(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/team")
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	avatar := jpegDataURL(t, 512, 512)
	created := expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member+calls@example.com", "password": "MemberPassword123!", "timezone": "UTC", "avatar": avatar,
	}), http.StatusCreated)
	avatarFilename := avatarFilenameFromResponse(t, created, "member+calls__example.com")
	user, err := f.store.GetUser("member+calls@example.com")
	if err != nil || user.AvatarPath != avatarFilename {
		t.Fatalf("avatar filename was not stored on the user: user=%#v err=%v", user, err)
	}
	stored, err := os.ReadFile(filepath.Join(f.dataPath, "content", "avatars", avatarFilename))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(stored)); err != nil {
		t.Fatalf("stored avatar is not a JPEG: %v", err)
	}
	served := f.request(http.MethodGet, "/content/avatars/"+avatarFilename, nil)
	servedBody := expectStatus(t, served, http.StatusOK)
	if served.Header.Get("Content-Type") != "image/jpeg" || !bytes.Equal(servedBody, stored) {
		t.Fatal("avatar response did not serve the stored JPEG")
	}

	expectStatus(t, f.request(http.MethodGet, "/content/avatars/not-an-avatar.txt", nil), http.StatusNotFound)
	expectStatus(t, f.request(http.MethodGet, "/content/users.leveldb/CURRENT", nil), http.StatusNotFound)
	traversal := f.request(http.MethodGet, "/content/avatars/%2e%2e%2fusers.leveldb", nil)
	defer traversal.Body.Close()
	if traversal.StatusCode >= 200 && traversal.StatusCode < 300 {
		t.Fatalf("path traversal request succeeded with status %d", traversal.StatusCode)
	}
}

func TestUserAvatarValidation(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC", "avatar": jpegDataURL(t, 64, 64),
	}), http.StatusBadRequest)
	if _, err := f.store.GetUser("member@example.com"); err == nil {
		t.Fatal("user was created with an invalid avatar")
	}
}

func TestBookingAPIs(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/bookings", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	createdEventType := expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Planning Call", []string{adminEmail}, 1)), http.StatusCreated)
	if !strings.Contains(string(createdEventType), `"eventSlug":"planning-call"`) {
		t.Fatalf("unexpected event type response: %s", createdEventType)
	}
	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	fractional := map[string]string{
		"eventSlug": "planning-call", "time": strings.TrimSuffix(bookingTime, "Z") + ".000Z", "attendeeEmail": "guest@example.com",
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", fractional), http.StatusBadRequest)
	offset := map[string]string{
		"eventSlug": "planning-call", "time": strings.TrimSuffix(bookingTime, "Z") + "+00:00", "attendeeEmail": "guest@example.com",
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", offset), http.StatusBadRequest)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	booking := map[string]string{
		"eventSlug": "planning-call", "time": bookingTime, "attendeeName": "Guest Person", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC", "notes": "Discuss launch",
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": "planning-call", "time": bookingTime, "attendeeName": "Guest Person", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC", "guestEmails": []string{"friend@example.com"}, "notes": "Discuss launch",
	}), http.StatusConflict)
	created := expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking), http.StatusCreated)
	var createdResponse struct {
		Booking struct {
			ID string `json:"id"`
		} `json:"booking"`
	}
	if err := json.Unmarshal(created, &createdResponse); err != nil || createdResponse.Booking.ID == "" {
		t.Fatalf("booking response did not include an ID: %s", created)
	}
	publicEventType := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/planning-call", nil), http.StatusOK)
	if !strings.Contains(string(publicEventType), `"busyRanges":[{"start":"`+bookingTime+`"`) {
		t.Fatalf("booked slot was not exposed as unavailable: %s", publicEventType)
	}
	if !strings.Contains(string(publicEventType), `"remainingInvitees":{"`+bookingTime+`":0}`) {
		t.Fatalf("remaining invitee capacity was not exposed: %s", publicEventType)
	}
	limitReached := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": "planning-call", "time": bookingTime, "attendeeName": "Second Guest", "attendeeEmail": "second@example.com", "attendeeTimezone": "UTC", "notes": "",
	}), http.StatusConflict)
	if !strings.Contains(string(limitReached), "invitee limit has been reached") {
		t.Fatalf("unexpected capacity error: %s", limitReached)
	}
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	listed := expectStatus(t, f.request(http.MethodGet, "/api/bookings", nil), http.StatusOK)
	if !strings.Contains(string(listed), "guest@example.com") {
		t.Fatalf("booking missing from list: %s", listed)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/bookings/"+createdResponse.Booking.ID, nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/bookings/"+createdResponse.Booking.ID, nil), http.StatusNotFound)
}

func TestBookingsUseUTCSlotKey(t *testing.T) {
	dataPath := t.TempDir()
	database, err := store.Open(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.July, 16, 8, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	slotKey := "planning-call" + start.Format(time.RFC3339) + "-" + end.Format(time.RFC3339)
	limit := 2
	for index, email := range []string{"first@example.com", "second@example.com"} {
		booking := model.Booking{
			ID:            fmt.Sprintf("booking-%d", index),
			EventSlug:     "planning-call",
			Time:          start,
			EndTime:       end,
			AttendeeName:  "Guest",
			AttendeeEmail: email,
		}
		if err := database.CreateBooking(slotKey, booking, nil, &limit); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	bookings, err := leveldb.OpenFile(filepath.Join(dataPath, "bookings.leveldb"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer bookings.Close()
	iterator := bookings.NewIterator(nil, nil)
	defer iterator.Release()
	if !iterator.Next() || string(iterator.Key()) != slotKey {
		t.Fatalf("booking slot key = %q, want %q", iterator.Key(), slotKey)
	}
	if iterator.Next() {
		t.Fatalf("multiple LevelDB records were created for one booking slot")
	}
}

func TestPublicEventTypeMarksStoredSlotUnavailableWithoutInviteeLimit(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	eventType := eventTypeBody("Exclusive Call", []string{adminEmail}, 1)
	eventType["inviteeLimit"] = nil
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventType), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": "exclusive-call", "time": bookingTime, "attendeeName": "Guest", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC",
		"guestEmails": []string{"friend@example.com"},
	}), http.StatusCreated)

	publicEventType := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/exclusive-call", nil), http.StatusOK)
	if !strings.Contains(string(publicEventType), `"busyRanges":[{"start":"`+bookingTime+`"`) {
		t.Fatalf("stored slot was not exposed as unavailable without an invitee limit: %s", publicEventType)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": "exclusive-call", "time": bookingTime, "attendeeName": "Second Guest", "attendeeEmail": "second@example.com", "attendeeTimezone": "UTC",
	}), http.StatusConflict)
}

func TestBookingManagementAPIsWithSecretAndAuthenticatedActors(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Managed Call", []string{adminEmail}, 3)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)

	candidate := time.Now().UTC().AddDate(0, 0, 3)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 13, 0, 0, 0, time.UTC).Format(time.RFC3339)
	created := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": "managed-call", "time": bookingTime, "attendeeName": "Public Guest", "attendeeEmail": "guest@example.com", "attendeeTimezone": "Europe/London", "notes": "Initial notes",
	}), http.StatusCreated)
	var creation struct {
		Booking   model.Booking `json:"booking"`
		ManageURL string        `json:"manageURL"`
	}
	if err := json.Unmarshal(created, &creation); err != nil {
		t.Fatal(err)
	}
	manageURL, err := url.Parse(creation.ManageURL)
	if err != nil {
		t.Fatal(err)
	}
	secretToken := strings.TrimPrefix(manageURL.Path, "/event/")
	if secretToken == "" || strings.ContainsAny(secretToken, "/+=") {
		t.Fatalf("booking secret is not URL-safe: %q", secretToken)
	}
	if _, err := os.Stat(filepath.Join(f.dataPath, "secret_link_map.leveldb")); err != nil {
		t.Fatalf("secret link map was not created: %v", err)
	}
	managed := expectStatus(t, f.request(http.MethodGet, "/api/events/"+secretToken, nil), http.StatusOK)
	if !strings.Contains(string(managed), `"authenticated":false`) || !strings.Contains(string(managed), `"guestLimit":2`) {
		t.Fatalf("public management response was not anonymous: %s", managed)
	}
	updated := expectStatus(t, f.request(http.MethodPatch, "/api/events/"+secretToken, map[string]any{
		"notes": "Updated description", "guestEmails": []string{"one@example.com", "two@example.com"},
	}), http.StatusOK)
	if !strings.Contains(string(updated), "Updated description") || !strings.Contains(string(updated), "two@example.com") {
		t.Fatalf("booking details were not updated: %s", updated)
	}
	expectStatus(t, f.request(http.MethodPatch, "/api/events/"+secretToken, map[string]any{
		"notes": "Too many", "guestEmails": []string{"one@example.com", "two@example.com", "three@example.com"},
	}), http.StatusConflict)
	canceled := expectStatus(t, f.request(http.MethodPost, "/api/events/"+secretToken+"/cancel", map[string]string{
		"reason": "Plans changed",
	}), http.StatusOK)
	if !strings.Contains(string(canceled), `"email":"guest@example.com"`) || !strings.Contains(string(canceled), `"cancellationReason":"Plans changed"`) {
		t.Fatalf("public cancellation actor or reason was not recorded: %s", canceled)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/events/"+secretToken+"/cancel", map[string]string{"reason": "Again"}), http.StatusConflict)

	secondTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 14, 0, 0, 0, time.UTC).Format(time.RFC3339)
	secondCreated := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]string{
		"eventSlug": "managed-call", "time": secondTime, "attendeeName": "Second Guest", "attendeeEmail": "second@example.com", "attendeeTimezone": "UTC", "notes": "",
	}), http.StatusCreated)
	if err := json.Unmarshal(secondCreated, &creation); err != nil {
		t.Fatal(err)
	}
	secondURL, _ := url.Parse(creation.ManageURL)
	secondSecret := strings.TrimPrefix(secondURL.Path, "/event/")
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	authenticated := expectStatus(t, f.request(http.MethodGet, "/api/events/"+secondSecret, nil), http.StatusOK)
	if !strings.Contains(string(authenticated), `"authenticated":true`) {
		t.Fatalf("signed-in management response was not authenticated: %s", authenticated)
	}
	adminCanceled := expectStatus(t, f.request(http.MethodPost, "/api/events/"+secondSecret+"/cancel", map[string]string{"reason": "Host unavailable"}), http.StatusOK)
	if !strings.Contains(string(adminCanceled), `"email":"admin@example.com"`) {
		t.Fatalf("authenticated cancellation actor was not recorded: %s", adminCanceled)
	}
}

func TestBookingDeliveryUsesMailgunAndConnectedGoogleCalendar(t *testing.T) {
	f := newConfiguredFixture(t, true, "", func(cfg *config.Config) {
		cfg.Mailing.Mailgun = config.Mailgun{
			APIKey: "mailgun-key", BaseURL: "https://api.eu.mailgun.net", Domain: "mail.example.com", From: "Bookings <bookings@example.com>",
		}
	})
	mockGoogleProvider(t, adminEmail, "", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	calendarRequests, calendarUpdates, mailgunRequests := mockBookingDelivery(t)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Delivery test", []string{adminEmail}, 3)), http.StatusCreated)
	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	created := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": "delivery-test", "time": bookingTime, "attendeeName": "Guest Person", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC", "guestEmails": []string{"friend@example.com"}, "notes": "",
	}), http.StatusCreated)
	if *calendarRequests != 1 || *calendarUpdates != 0 || *mailgunRequests != 1 {
		t.Fatalf("unexpected delivery calls: calendar=%d updates=%d mailgun=%d", *calendarRequests, *calendarUpdates, *mailgunRequests)
	}
	var response struct {
		ManageURL string `json:"manageURL"`
	}
	if err := json.Unmarshal(created, &response); err != nil {
		t.Fatal(err)
	}
	manageURL, _ := url.Parse(response.ManageURL)
	secretToken := strings.TrimPrefix(manageURL.Path, "/event/")
	expectStatus(t, f.request(http.MethodPatch, "/api/events/"+secretToken, map[string]any{
		"notes": "Bring the project plan", "guestEmails": []string{"friend@example.com", "second-friend@example.com"},
	}), http.StatusOK)
	if *calendarUpdates != 1 {
		t.Fatalf("guest update did not update Google Calendar: %d", *calendarUpdates)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/events/"+secretToken+"/cancel", map[string]string{"reason": "Host unavailable"}), http.StatusOK)
	if *calendarRequests != 1 || *calendarUpdates != 2 || *mailgunRequests != 2 {
		t.Fatalf("unexpected cancellation delivery calls: calendar=%d updates=%d mailgun=%d", *calendarRequests, *calendarUpdates, *mailgunRequests)
	}
}

func TestEventTypeAPIs(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/event-types", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "fullName": "Ada Lovelace", "password": "MemberPassword123!", "timezone": "Europe/Kyiv",
	}), http.StatusCreated)
	created := expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Team Consultation!", []string{adminEmail}, 3)), http.StatusCreated)
	if !strings.Contains(string(created), `"eventSlug":"team-consultation"`) {
		t.Fatalf("event slug was not generated: %s", created)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Team Consultation", []string{adminEmail}, 3)), http.StatusConflict)
	updatedBody := eventTypeBody("Renamed Consultation", []string{adminEmail, "member@example.com"}, 4)
	updated := expectStatus(t, f.request(http.MethodPut, "/api/event-types/team-consultation", updatedBody), http.StatusOK)
	if !strings.Contains(string(updated), `"eventSlug":"team-consultation"`) || !strings.Contains(string(updated), "member@example.com") {
		t.Fatalf("event type update changed the slug or missed its recipient: %s", updated)
	}
	listed := expectStatus(t, f.request(http.MethodGet, "/api/event-types", nil), http.StatusOK)
	if !strings.Contains(string(listed), "Renamed Consultation") {
		t.Fatalf("event type missing from list: %s", listed)
	}
	public := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/team-consultation", nil), http.StatusOK)
	if !strings.Contains(string(public), `"requiredHosts"`) || !strings.Contains(string(public), `"fullName":"Ada Lovelace"`) {
		t.Fatalf("public event type did not include hosts: %s", public)
	}
	invalid := eventTypeBody("No recipients", nil, 1)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", invalid), http.StatusBadRequest)
	duplicate := eventTypeBody("Duplicate host", []string{adminEmail}, 1)
	duplicate["optionalHostEmails"] = []string{adminEmail}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", duplicate), http.StatusBadRequest)
	expectStatus(t, f.request(http.MethodDelete, "/api/event-types/team-consultation", nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodGet, "/api/event-types/team-consultation", nil), http.StatusNotFound)
	expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/team-consultation", nil), http.StatusNotFound)
	expectStatus(t, f.request(http.MethodDelete, "/api/event-types/team-consultation", nil), http.StatusNotFound)
}

func TestDeletingUserUpdatesEventTypeRecipients(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	createUser := map[string]string{"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC"}
	expectStatus(t, f.request(http.MethodPost, "/api/users", createUser), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Shared", []string{adminEmail, "member@example.com"}, 1)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)
	eventType, err := f.store.GetEventType("shared")
	if err != nil || len(eventType.RequiredHostEmails) != 1 || eventType.RequiredHostEmails[0] != adminEmail {
		t.Fatalf("deleted user was not removed from event type: event=%#v err=%v", eventType, err)
	}
}

func TestDeletingFinalEventTypeRecipientIsRejected(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	createUser := map[string]string{"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC"}
	expectStatus(t, f.request(http.MethodPost, "/api/users", createUser), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Member only", []string{"member@example.com"}, 1)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusConflict)
	if _, err := f.store.GetUser("member@example.com"); err != nil {
		t.Fatal("final event type recipient was deleted")
	}
}

func TestRequiredAndOptionalHostAvailability(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC",
	}), http.StatusCreated)
	eventType := eventTypeBody("Host roles", []string{adminEmail}, 1)
	eventType["optionalHostEmails"] = []string{"member@example.com"}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventType), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	start := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	memberBusy := model.GoogleBusyCache{
		Periods:  []model.GoogleBusyPeriod{{EventID: "member-event", Start: start, End: end}},
		SyncedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := f.store.PutGoogleBusy("member@example.com", memberBusy); err != nil {
		t.Fatal(err)
	}
	public := expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/host-roles", nil), http.StatusOK)
	if !strings.Contains(string(public), "\"optionalHosts\"") || !strings.Contains(string(public), "\"busyRanges\":[]") {
		t.Fatalf("optional host blocked availability: %s", public)
	}

	adminBusy := model.GoogleBusyCache{
		Periods:  []model.GoogleBusyPeriod{{EventID: "admin-event", Start: start, End: end}},
		SyncedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := f.store.PutGoogleBusy(adminEmail, adminBusy); err != nil {
		t.Fatal(err)
	}
	public = expectStatus(t, f.request(http.MethodGet, "/api/public/event-types/host-roles", nil), http.StatusOK)
	if !strings.Contains(string(public), "\"start\":\""+start.Format(time.RFC3339)+"\"") {
		t.Fatalf("required host busy period was not exposed: %s", public)
	}
}

func TestCrossEventRequiredHostConflictIgnoresOptionalHosts(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC",
	}), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("First", []string{adminEmail}, 1)), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Required conflict", []string{adminEmail}, 1)), http.StatusCreated)
	optional := eventTypeBody("Optional conflict", []string{"member@example.com"}, 1)
	optional["optionalHostEmails"] = []string{adminEmail}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", optional), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	bookingTime := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	body := func(slug, email string) map[string]string {
		return map[string]string{
			"eventSlug": slug, "time": bookingTime, "attendeeName": "Guest", "attendeeEmail": email, "attendeeTimezone": "UTC",
		}
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", body("first", "first@example.com")), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", body("required-conflict", "second@example.com")), http.StatusConflict)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", body("optional-conflict", "third@example.com")), http.StatusCreated)
}

func TestGoogleCalendarSyncRetainsFailuresAndRemovesDeletedEvents(t *testing.T) {
	f := newFixture(t, true)
	mockGoogleProvider(t, adminEmail, "", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Synced", []string{adminEmail}, 1)), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	start := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	limit := 2
	ownedBooking := model.Booking{
		ID: "owned-booking", EventSlug: "synced", Time: start, EndTime: end,
		AttendeeName: "Guest", AttendeeEmail: "owned@example.com", RecipientEmails: []string{adminEmail},
		GoogleEventIDs: map[string]string{adminEmail: "owned"},
	}
	ownedKey := "synced" + start.Format(time.RFC3339) + "-" + end.Format(time.RFC3339)
	if err := f.store.CreateBooking(ownedKey, ownedBooking, []string{adminEmail}, &limit); err != nil {
		t.Fatal(err)
	}
	originalTransport := http.DefaultTransport
	cycle := 0
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/calendar/v3/calendars/primary/events" {
			return nil, fmt.Errorf("unexpected sync request: %s %s", request.Method, request.URL)
		}
		if request.URL.Query().Get("singleEvents") != "true" || request.URL.Query().Get("maxResults") != "2500" {
			t.Errorf("Google sync did not request expanded events: %s", request.URL)
		}
		pageToken := request.URL.Query().Get("pageToken")
		if pageToken == "" {
			cycle++
		}
		if cycle == 2 {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Status:     "500 Internal Server Error",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("{}")),
				Request:    request,
			}, nil
		}
		items := []map[string]any{}
		nextPageToken := ""
		if cycle == 1 && pageToken == "" {
			items = append(items,
				map[string]any{"id": "external", "status": "confirmed", "start": map[string]string{"dateTime": start.Format(time.RFC3339)}, "end": map[string]string{"dateTime": end.Format(time.RFC3339)}},
				map[string]any{"id": "owned", "status": "confirmed", "start": map[string]string{"dateTime": start.Format(time.RFC3339)}, "end": map[string]string{"dateTime": end.Format(time.RFC3339)}},
				map[string]any{"id": "transparent", "status": "confirmed", "transparency": "transparent", "start": map[string]string{"dateTime": start.Format(time.RFC3339)}, "end": map[string]string{"dateTime": end.Format(time.RFC3339)}},
			)
			nextPageToken = "second"
		} else if cycle == 1 {
			items = append(items,
				map[string]any{"id": "all-day", "status": "confirmed", "start": map[string]string{"date": start.Format(time.DateOnly)}, "end": map[string]string{"date": start.AddDate(0, 0, 1).Format(time.DateOnly)}},
				map[string]any{"id": "cancelled", "status": "cancelled", "start": map[string]string{"dateTime": start.Format(time.RFC3339)}, "end": map[string]string{"dateTime": end.Format(time.RFC3339)}},
			)
		}
		body, _ := json.Marshal(map[string]any{"timeZone": "UTC", "nextPageToken": nextPageToken, "items": items})
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	f.api.SyncGoogleCalendars(context.Background())
	cache, err := f.store.GetGoogleBusy(adminEmail)
	if err != nil || len(cache.Periods) != 2 {
		t.Fatalf("initial Google cache = %#v, err=%v", cache, err)
	}
	f.api.SyncGoogleCalendars(context.Background())
	cache, err = f.store.GetGoogleBusy(adminEmail)
	if err != nil || len(cache.Periods) != 2 {
		t.Fatalf("failed sync replaced Google cache = %#v, err=%v", cache, err)
	}
	f.api.SyncGoogleCalendars(context.Background())
	cache, err = f.store.GetGoogleBusy(adminEmail)
	if err != nil || len(cache.Periods) != 0 {
		t.Fatalf("deleted Google events remained cached = %#v, err=%v", cache, err)
	}
}

func TestBookingLiveGoogleAvailability(t *testing.T) {
	f := newFixture(t, true)
	mockGoogleProvider(t, adminEmail, "", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Live check", []string{adminEmail}, 1)), http.StatusCreated)

	candidate := time.Now().UTC().AddDate(0, 0, 2)
	start := time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 12, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	originalTransport := http.DefaultTransport
	call := 0
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		call++
		status := http.StatusOK
		body, _ := json.Marshal(map[string]any{
			"timeZone": "UTC",
			"items": []map[string]any{{
				"id": "busy", "status": "confirmed",
				"start": map[string]string{"dateTime": start.Format(time.RFC3339)},
				"end":   map[string]string{"dateTime": end.Format(time.RFC3339)},
			}},
		})
		if call == 2 {
			status = http.StatusInternalServerError
			body = []byte("{}")
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	request := map[string]string{
		"eventSlug": "live-check", "time": start.Format(time.RFC3339), "attendeeName": "Guest", "attendeeEmail": "guest@example.com", "attendeeTimezone": "UTC",
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", request), http.StatusConflict)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", request), http.StatusServiceUnavailable)
}

func TestGoogleOAuthAPIs(t *testing.T) {
	disabled := newFixture(t, false)
	expectStatus(t, disabled.request(http.MethodGet, "/api/auth/google/start", nil), http.StatusNotFound)
	expectStatus(t, disabled.request(http.MethodPost, "/api/auth/google/callback", nil), http.StatusNotFound)

	enabled := newFixtureAtBasePath(t, true, "/calendar")
	configBody := expectStatus(t, enabled.request(http.MethodGet, "/api/config/public", nil), http.StatusOK)
	if !strings.Contains(string(configBody), `"googleLoginEnabled":true`) {
		t.Fatalf("unexpected config: %s", configBody)
	}
	start := enabled.request(http.MethodGet, "/api/auth/google/start", nil)
	expectStatus(t, start, http.StatusFound)
	location, err := url.Parse(start.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	query := location.Query()
	if query.Get("state") == "" || query.Get("code_challenge") == "" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("OAuth redirect is missing state or PKCE: %s", location.String())
	}
	if !strings.Contains(query.Get("scope"), "calendar.events") {
		t.Fatalf("OAuth redirect is missing calendar scope: %s", query.Get("scope"))
	}
	if query.Get("redirect_uri") != "http://example.test/calendar/auth/google/callback" {
		t.Fatalf("unexpected OAuth redirect URI: %s", query.Get("redirect_uri"))
	}
	expectStatus(t, enabled.request(http.MethodPost, "/api/auth/google/callback", map[string]string{"error": "access_denied"}), http.StatusBadRequest)
	expectStatus(t, enabled.request(http.MethodPost, "/api/auth/google/callback", map[string]string{"state": "invalid", "code": "invalid"}), http.StatusBadRequest)
}

func TestGoogleOAuthImportsMissingFullName(t *testing.T) {
	f := newFixture(t, true)
	mockGoogleProvider(t, adminEmail, "Ada Lovelace", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	user, err := f.store.GetUser(adminEmail)
	if err != nil || user.FullName != "Ada Lovelace" {
		t.Fatalf("Google full name was not imported: user=%#v err=%v", user, err)
	}
}

func TestGoogleOAuthKeepsExistingFullName(t *testing.T) {
	f := newFixture(t, true)
	user, err := f.store.GetUser(adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	user.FullName = "Existing User"
	if err := f.store.PutUser(user); err != nil {
		t.Fatal(err)
	}
	mockGoogleProvider(t, adminEmail, "Google Profile", "", nil)
	expectStatus(t, googleCallback(f), http.StatusOK)
	user, err = f.store.GetUser(adminEmail)
	if err != nil || user.FullName != "Existing User" {
		t.Fatalf("Google replaced an existing full name: user=%#v err=%v", user, err)
	}
}

func TestGoogleOAuthRejectsUnknownUser(t *testing.T) {
	f := newFixture(t, true)
	mockGoogleProvider(t, "unknown@example.com", "Unknown User", "", nil)
	body := expectStatus(t, googleCallback(f), http.StatusForbidden)
	if !strings.Contains(string(body), "Ask an administrator to add your email") {
		t.Fatalf("unexpected unknown-user error: %s", body)
	}
	if _, err := f.store.GetUser("unknown@example.com"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown Google user was created: %v", err)
	}
}

func TestGoogleOAuthImportsMissingAvatar(t *testing.T) {
	f := newFixtureAtBasePath(t, true, "/calendar")
	pictureURL := "https://lh3.googleusercontent.com/profile.jpg"
	avatarRequests := mockGoogleProvider(t, adminEmail, "", pictureURL, jpegBytes(t, 640, 320))
	expectStatus(t, googleCallback(f), http.StatusOK)
	if *avatarRequests != 1 {
		t.Fatalf("expected one Google avatar request, got %d", *avatarRequests)
	}
	user, err := f.store.GetUser(adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	avatarFilename := requireAvatarFilename(t, user.AvatarPath, "admin__example.com")
	stored, err := os.Open(filepath.Join(f.dataPath, "content", "avatars", avatarFilename))
	if err != nil {
		t.Fatal(err)
	}
	defer stored.Close()
	config, err := jpeg.DecodeConfig(stored)
	if err != nil || config.Width != 512 || config.Height != 512 {
		t.Fatalf("Google avatar was not normalized to 512 by 512: config=%#v err=%v", config, err)
	}
}

func TestGoogleOAuthKeepsExistingAvatar(t *testing.T) {
	f := newFixture(t, true)
	const avatarFilename = "admin__example.com-1234abcd.jpg"
	existing := jpegBytes(t, 512, 512)
	if err := os.WriteFile(filepath.Join(f.dataPath, "content", "avatars", avatarFilename), existing, 0o600); err != nil {
		t.Fatal(err)
	}
	user, err := f.store.GetUser(adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	user.AvatarPath = avatarFilename
	if err := f.store.PutUser(user); err != nil {
		t.Fatal(err)
	}
	pictureURL := "https://lh3.googleusercontent.com/profile.jpg"
	avatarRequests := mockGoogleProvider(t, adminEmail, "", pictureURL, jpegBytes(t, 320, 640))
	expectStatus(t, googleCallback(f), http.StatusOK)
	if *avatarRequests != 0 {
		t.Fatalf("Google avatar was requested despite an existing avatar path")
	}
	stored, err := os.ReadFile(filepath.Join(f.dataPath, "content", "avatars", avatarFilename))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, existing) {
		t.Fatal("existing avatar file was overwritten")
	}
}

func mustDo(t *testing.T, client *http.Client, request *http.Request) *http.Response {
	t.Helper()
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func eventTypeBody(name string, recipients []string, inviteeLimit int) map[string]any {
	schedule := make([]map[string]any, 0, 7)
	for _, day := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
		schedule = append(schedule, map[string]any{
			"day": day, "enabled": true, "start": "00:00", "end": "23:59", "breaks": []any{},
		})
	}
	return map[string]any{
		"name":               name,
		"durationMinutes":    30,
		"bookingWindowDays":  60,
		"inviteeLimit":       inviteeLimit,
		"timezone":           "UTC",
		"requiredHostEmails": recipients,
		"optionalHostEmails": []string{},
		"schedule":           schedule,
	}
}

func avatarFilenameFromResponse(t *testing.T, body []byte, emailSlug string) string {
	t.Helper()
	var response struct {
		User struct {
			AvatarPath string `json:"avatarPath"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	return requireAvatarFilename(t, response.User.AvatarPath, emailSlug)
}

func logoFilenameFromResponse(t *testing.T, body []byte) string {
	t.Helper()
	var response struct {
		Branding model.Branding `json:"branding"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	return requireAvatarFilename(t, response.Branding.LogoPath, "logo")
}

func requireAvatarFilename(t *testing.T, filename, emailSlug string) string {
	t.Helper()
	prefix := emailSlug + "-"
	token, found := strings.CutPrefix(filename, prefix)
	if !found {
		t.Fatalf("avatar filename %q does not start with %q", filename, prefix)
	}
	token, found = strings.CutSuffix(token, ".jpg")
	if !found || len(token) != 8 {
		t.Fatalf("avatar filename %q does not end with an eight-character token and .jpg", filename)
	}
	if _, err := hex.DecodeString(token); err != nil {
		t.Fatalf("avatar filename %q does not contain a hexadecimal token", filename)
	}
	return filename
}

func jpegDataURL(t *testing.T, width, height int) string {
	t.Helper()
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpegBytes(t, width, height))
}

func jpegBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, width, height)), &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}

func googleCallback(f *fixture) *http.Response {
	f.t.Helper()
	start := f.request(http.MethodGet, "/api/auth/google/start", nil)
	expectStatus(f.t, start, http.StatusFound)
	location, err := url.Parse(start.Header.Get("Location"))
	if err != nil {
		f.t.Fatal(err)
	}
	return f.request(http.MethodPost, "/api/auth/google/callback", map[string]string{
		"state": location.Query().Get("state"),
		"code":  "test-code",
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func mockGoogleProvider(t *testing.T, email, fullName, pictureURL string, picture []byte) *int {
	t.Helper()
	originalTransport := http.DefaultTransport
	avatarRequests := 0
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var contentType string
		var body []byte
		switch request.URL.String() {
		case "https://oauth2.googleapis.com/token":
			contentType = "application/json"
			body = []byte(`{"access_token":"google-access-token","token_type":"Bearer","refresh_token":"google-refresh-token","expires_in":3600}`)
		case "https://openidconnect.googleapis.com/v1/userinfo":
			contentType = "application/json"
			body, _ = json.Marshal(map[string]any{
				"email": email, "email_verified": true, "name": fullName, "picture": pictureURL,
			})
		case pictureURL:
			avatarRequests++
			contentType = "image/jpeg"
			body = picture
		default:
			return nil, fmt.Errorf("unexpected Google request: %s", request.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Header:     http.Header{"Content-Type": []string{contentType}},
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })
	return &avatarRequests
}

func mockBookingDelivery(t *testing.T) (*int, *int, *int) {
	t.Helper()
	originalTransport := http.DefaultTransport
	calendarRequests := 0
	calendarUpdates := 0
	mailgunRequests := 0
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/calendar/v3/calendars/primary/events":
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"timeZone":"UTC","items":[]}`)),
				Request:    request,
			}, nil
		case request.Method == http.MethodPost && request.URL.Path == "/calendar/v3/calendars/primary/events":
			calendarRequests++
			body, _ := io.ReadAll(request.Body)
			if request.URL.Query().Get("sendUpdates") != "all" || !strings.Contains(string(body), "Cancel or update event") || !strings.Contains(string(body), "/event/") || !strings.Contains(string(body), `"email":"guest@example.com"`) || !strings.Contains(string(body), `"email":"friend@example.com"`) {
				t.Errorf("Google event did not include attendee invitations and management link: url=%s body=%s", request.URL, body)
			}
		case request.Method == http.MethodPatch && request.URL.Path == "/calendar/v3/calendars/primary/events/accepted":
			calendarUpdates++
			body, _ := io.ReadAll(request.Body)
			if request.URL.Query().Get("sendUpdates") != "all" || !strings.Contains(string(body), `"email":"friend@example.com"`) {
				t.Errorf("Google event update did not send attendee updates: url=%s body=%s", request.URL, body)
			} else if calendarUpdates == 1 && (!strings.Contains(string(body), `"summary":"Delivery test"`) || !strings.Contains(string(body), `"email":"second-friend@example.com"`)) {
				t.Errorf("Google guest update was not correct: %s", body)
			} else if calendarUpdates == 2 && (!strings.Contains(string(body), `"summary":"Canceled: Delivery test"`) || !strings.Contains(string(body), "Reason: Host unavailable")) {
				t.Errorf("Google cancellation event was not updated correctly: %s", body)
			}
		case request.URL.String() == "https://api.eu.mailgun.net/v3/mail.example.com/messages":
			body, _ := io.ReadAll(request.Body)
			values, _ := url.ParseQuery(string(body))
			wantSubject := "New Event -"
			if mailgunRequests == 1 {
				wantSubject = "Canceled Event -"
			}
			if !strings.HasPrefix(values.Get("subject"), wantSubject) || values.Get("html") == "" || values.Get("text") == "" {
				t.Errorf("Mailgun message is missing rendered parts: %s", body)
			}
			mailgunRequests++
		default:
			return nil, fmt.Errorf("unexpected delivery request: %s", request.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":"accepted"}`)),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })
	return &calendarRequests, &calendarUpdates, &mailgunRequests
}
