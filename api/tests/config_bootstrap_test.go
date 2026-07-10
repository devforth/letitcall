package tests

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/bootstrap"
	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/store"
)

func TestConfigUsesStrictEnvironmentNamesAndDefaults(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvHTTPPort, "8080")
	t.Setenv(config.EnvStorageLevelDBPath, t.TempDir())
	t.Setenv(config.EnvFirstUserEmail, "owner@example.com")
	t.Setenv(config.EnvFirstUserPassword, "OwnerPassword123!")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Port != 8080 || cfg.FirstUser.Email != "owner@example.com" {
		t.Fatalf("configuration did not load structured environment variables: %#v", cfg)
	}
}

func TestConfigRejectsPartialGoogleOAuthSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvGoogleClientID, "client-id-only")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "all LOGIN__OAUTH__GOOGLE settings") {
		t.Fatalf("expected partial Google config error, got %v", err)
	}
}

func TestConfigAcceptsCompleteSecureGoogleOAuthSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvHTTPPublicURL, "https://calendar.example.com")
	t.Setenv(config.EnvGoogleClientID, "client-id")
	t.Setenv(config.EnvGoogleClientSecret, "client-secret")
	t.Setenv(config.EnvGoogleTokenEncryptionKey, base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Login.Google.RedirectURL != "https://calendar.example.com/api/auth/google/callback" {
		t.Fatalf("unexpected derived redirect URL: %s", cfg.Login.Google.RedirectURL)
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		config.EnvHTTPPort,
		config.EnvHTTPPublicURL,
		config.EnvHTTPReadTimeout,
		config.EnvHTTPWriteTimeout,
		config.EnvHTTPIdleTimeout,
		config.EnvHTTPShutdownTimeout,
		config.EnvStorageLevelDBPath,
		config.EnvFirstUserEmail,
		config.EnvFirstUserPassword,
		config.EnvSessionTTL,
		config.EnvSessionCookieSecure,
		config.EnvPasswordMaxAttempts,
		config.EnvPasswordLockout,
		config.EnvGoogleClientID,
		config.EnvGoogleClientSecret,
		config.EnvGoogleRedirectURL,
		config.EnvGoogleTokenEncryptionKey,
	} {
		t.Setenv(name, "")
	}
	t.Setenv(config.EnvStorageLevelDBPath, t.TempDir())
}

func TestFirstUserBootstrapSeedsOnlyAnEmptyUsersTable(t *testing.T) {
	database, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	credentials := config.FirstUser{Email: "owner@example.com", Password: "OwnerPassword123!"}
	if err := bootstrap.EnsureFirstUser(database, credentials, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := bootstrap.EnsureFirstUser(database, config.FirstUser{}, time.Now()); err != nil {
		t.Fatalf("existing users should not require bootstrap credentials: %v", err)
	}
	count, err := database.UserCount()
	if err != nil || count != 1 {
		t.Fatalf("expected exactly one bootstrapped user, count=%d err=%v", count, err)
	}
}

func TestFirstUserBootstrapRequiresCredentialsForEmptyTable(t *testing.T) {
	database, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := bootstrap.EnsureFirstUser(database, config.FirstUser{}, time.Now()); err == nil {
		t.Fatal("expected empty store bootstrap to require credentials")
	}
}
