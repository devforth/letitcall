package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
	dataPath string
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
	result := &fixture{t: t, server: testServer, client: client, store: database, basePath: basePath, dataPath: dataPath}
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

func TestGoogleOAuthImportsMissingAvatar(t *testing.T) {
	f := newFixtureAtBasePath(t, true, "/calendar")
	pictureURL := "https://lh3.googleusercontent.com/profile.jpg"
	avatarRequests := mockGoogleProvider(t, adminEmail, pictureURL, jpegBytes(t, 640, 320))
	expectStatus(t, googleCallback(f), http.StatusSeeOther)
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
	avatarRequests := mockGoogleProvider(t, adminEmail, pictureURL, jpegBytes(t, 320, 640))
	expectStatus(t, googleCallback(f), http.StatusSeeOther)
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
	return f.request(http.MethodGet, "/api/auth/google/callback?state="+url.QueryEscape(location.Query().Get("state"))+"&code=test-code", nil)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func mockGoogleProvider(t *testing.T, email, pictureURL string, picture []byte) *int {
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
				"email": email, "email_verified": true, "picture": pictureURL,
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
