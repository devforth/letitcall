package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/httpapi"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

const (
	adminEmail    = "admin@example.com"
	adminPassword = "AdminPassword123!"
)

type fixture struct {
	t        *testing.T
	server   *httptest.Server
	client   *http.Client
	store    *store.Store
	basePath string
}

func newFixture(t *testing.T, googleEnabled bool) *fixture {
	return newFixtureAtBasePath(t, googleEnabled, "")
}

func newFixtureAtBasePath(t *testing.T, googleEnabled bool, basePath string) *fixture {
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
	cfg.HTTP.BasePath = basePath
	if googleEnabled {
		cfg.Login.Google = config.GoogleOAuth{
			ClientID:     "google-client-id",
			ClientSecret: "google-client-secret",
		}
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
	result := &fixture{t: t, server: testServer, client: client, store: database, basePath: basePath}
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
		HTTP:    config.HTTP{Port: 80},
		Storage: config.Storage{LevelDBPath: dataPath},
		Login: config.Login{
			SessionTTL:          time.Hour,
			PasswordMaxAttempts: 20,
			PasswordLockout:     time.Minute,
		},
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
	if !strings.Contains(string(publicConfig), `"googleLoginEnabled":false`) {
		t.Fatalf("unexpected public config: %s", publicConfig)
	}
	expectStatus(t, f.request(http.MethodGet, "/api/does-not-exist", nil), http.StatusNotFound)
	portal := expectStatus(t, f.request(http.MethodGet, "/users", nil), http.StatusOK)
	if !strings.Contains(string(portal), "Let It Call") {
		t.Fatalf("SPA fallback did not serve the portal: %s", portal)
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

func TestAuthenticationAllowsConfiguredFirstUserIdentifier(t *testing.T) {
	f := newFixture(t, false)
	if err := f.store.DeleteUser(adminEmail); err != nil {
		t.Fatal(err)
	}
	user, err := security.NewFirstUser("admin", "admin", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := f.store.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	expectStatus(t, f.login("admin", "admin"), http.StatusOK)
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
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "Europe/London",
	}
	created := expectStatus(t, f.request(http.MethodPost, "/api/users", createBody), http.StatusCreated)
	if !strings.Contains(string(created), `"googleConnected":false`) {
		t.Fatalf("unexpected create response: %s", created)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/users", createBody), http.StatusConflict)
	expectStatus(t, f.request(http.MethodPatch, "/api/users/member@example.com", map[string]string{"timezone": "America/New_York"}), http.StatusOK)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/admin@example.com", nil), http.StatusConflict)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNotFound)
}

func TestBookingAPIs(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.request(http.MethodGet, "/api/bookings", nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	fractional := map[string]string{
		"time": "2026-08-10T12:00:00.000Z", "attendeeEmail": "guest@example.com", "title": "Planning",
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", fractional), http.StatusBadRequest)
	booking := map[string]string{
		"time": "2026-08-10T14:00:00+02:00", "attendeeEmail": "guest@example.com", "title": "Planning",
	}
	created := expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking), http.StatusCreated)
	if !strings.Contains(string(created), `"time":"2026-08-10T12:00:00Z"`) {
		t.Fatalf("booking was not normalized to UTC seconds: %s", created)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", booking), http.StatusConflict)
	listed := expectStatus(t, f.request(http.MethodGet, "/api/bookings", nil), http.StatusOK)
	if !strings.Contains(string(listed), "guest@example.com") {
		t.Fatalf("booking missing from list: %s", listed)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/bookings/"+url.PathEscape("2026-08-10T12:00:00Z"), nil), http.StatusNoContent)
	expectStatus(t, f.request(http.MethodDelete, "/api/bookings/"+url.PathEscape("2026-08-10T12:00:00Z"), nil), http.StatusNotFound)
}

func TestGoogleOAuthAPIs(t *testing.T) {
	disabled := newFixture(t, false)
	expectStatus(t, disabled.request(http.MethodGet, "/api/auth/google/start", nil), http.StatusNotFound)
	expectStatus(t, disabled.request(http.MethodGet, "/api/auth/google/callback", nil), http.StatusNotFound)

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
	if query.Get("redirect_uri") != enabled.server.URL+"/calendar/api/auth/google/callback" {
		t.Fatalf("unexpected OAuth redirect URI: %s", query.Get("redirect_uri"))
	}
	expectStatus(t, enabled.request(http.MethodGet, "/api/auth/google/callback?state=invalid&code=invalid", nil), http.StatusBadRequest)
}

func mustDo(t *testing.T, client *http.Client, request *http.Request) *http.Response {
	t.Helper()
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
