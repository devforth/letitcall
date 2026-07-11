package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/bootstrap"
	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/security"
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
	if cfg.HTTP.Port != 8080 || cfg.HTTP.BaseURL != config.DefaultBaseURL || cfg.Branding.Name != config.DefaultBrandName || cfg.FirstUser.Email != "owner@example.com" {
		t.Fatalf("configuration did not load structured environment variables: %#v", cfg)
	}
}

func TestDotEnvLoadOrderAndMissingFiles(t *testing.T) {
	clearConfigEnvironment(t)
	if err := os.Unsetenv(config.EnvHTTPPort); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv(config.EnvHTTPBaseURL); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Error(err)
		}
	})
	if err := config.LoadDotEnv(); err != nil {
		t.Fatal(err)
	}
	local := config.EnvHTTPPort + "=41784\n" + config.EnvHTTPBaseURL + "=https://local.example/local\n"
	if err := os.WriteFile(filepath.Join(directory, ".env.local"), []byte(local), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte(config.EnvHTTPPort+"=8080\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := config.LoadDotEnv(); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Port != 8080 || cfg.HTTP.BaseURL != "https://local.example/local" {
		t.Fatalf("unexpected dotenv configuration: %#v", cfg.HTTP)
	}
}

func TestConfigRejectsPartialGoogleOAuthSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvGoogleClientID, "client-id-only")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "client ID and client secret must be set together") {
		t.Fatalf("expected partial Google config error, got %v", err)
	}
}

func TestConfigLoadsBaseURLAndGoogleOAuthSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvHTTPBaseURL, "https://calls.example.com/calendar/")
	t.Setenv(config.EnvGoogleClientID, "client-id")
	t.Setenv(config.EnvGoogleClientSecret, "client-secret")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.BaseURL != "https://calls.example.com/calendar" || cfg.HTTP.BasePath() != "/calendar" || !cfg.Login.Google.Enabled() {
		t.Fatalf("unexpected base URL or Google settings: %#v", cfg)
	}
}

func TestConfigLoadsMailgunSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvMailgunAPIKey, "mailgun-key")
	t.Setenv(config.EnvMailgunBaseURL, "https://api.eu.mailgun.net/")
	t.Setenv(config.EnvMailgunDomain, "mail.example.com")
	t.Setenv(config.EnvMailgunFrom, "Let It Call <bookings@example.com>")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Mailing.Mailgun.Enabled() || cfg.Mailing.Mailgun.BaseURL != "https://api.eu.mailgun.net" || cfg.Mailing.Mailgun.Domain != "mail.example.com" {
		t.Fatalf("Mailgun settings were not loaded: %#v", cfg.Mailing.Mailgun)
	}
}

func TestConfigLoadsBrandName(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvBrandName, "DevForth")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Branding.Name != "DevForth" {
		t.Fatalf("brand name was not loaded: %#v", cfg.Branding)
	}
}

func TestConfigRejectsPartialMailgunSettings(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvMailgunAPIKey, "mailgun-key")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "must be set together") {
		t.Fatalf("expected partial Mailgun config error, got %v", err)
	}
}

func TestConfigRejectsMailgunBaseURLWithPath(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvMailgunAPIKey, "mailgun-key")
	t.Setenv(config.EnvMailgunBaseURL, "https://api.eu.mailgun.net/v3")
	t.Setenv(config.EnvMailgunDomain, "mail.example.com")
	t.Setenv(config.EnvMailgunFrom, "Bookings <bookings@example.com>")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "without a path") {
		t.Fatalf("expected invalid Mailgun base URL error, got %v", err)
	}
}

func TestConfigRejectsBaseURLWithoutOrigin(t *testing.T) {
	clearConfigEnvironment(t)
	t.Setenv(config.EnvHTTPBaseURL, "/calendar")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "must be a full HTTP or HTTPS URL") {
		t.Fatalf("expected invalid base URL error, got %v", err)
	}
}

func clearConfigEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		config.EnvHTTPPort,
		config.EnvHTTPBaseURL,
		config.EnvBrandName,
		config.EnvStorageLevelDBPath,
		config.EnvFirstUserEmail,
		config.EnvFirstUserPassword,
		config.EnvSessionTTL,
		config.EnvPasswordMaxAttempts,
		config.EnvPasswordLockout,
		config.EnvGoogleClientID,
		config.EnvGoogleClientSecret,
		config.EnvMailgunAPIKey,
		config.EnvMailgunBaseURL,
		config.EnvMailgunDomain,
		config.EnvMailgunFrom,
	} {
		t.Setenv(name, "")
	}
	t.Setenv(config.EnvStorageLevelDBPath, t.TempDir())
}

func TestGoogleTokenKeyIsGeneratedOnceInDataPath(t *testing.T) {
	dataPath := t.TempDir()
	first, err := security.LoadGoogleTokenKey(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := security.LoadGoogleTokenKey(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 || !bytes.Equal(first, second) {
		t.Fatal("Google token key was not persisted")
	}
	info, err := os.Stat(filepath.Join(dataPath, security.GoogleTokenKeyFile))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected Google token key permissions: %o", info.Mode().Perm())
	}
}

func TestFirstUserBootstrapSeedsOnlyAnEmptyUsersTable(t *testing.T) {
	database, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	credentials := config.FirstUser{Email: "admin", Password: "admin"}
	if err := bootstrap.EnsureFirstUser(database, credentials, time.Now()); err != nil {
		t.Fatal(err)
	}
	user, err := database.GetUser("admin")
	if err != nil || !security.CheckPassword(user.PasswordHash, "admin") {
		t.Fatalf("configured first-user credentials were not stored: user=%#v err=%v", user, err)
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
