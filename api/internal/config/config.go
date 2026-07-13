package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvHTTPPort            = "HTTP__PORT"
	EnvHTTPBaseURL         = "HTTP__BASE__URL"
	EnvStorageLevelDBPath  = "STORAGE__LEVELDB__PATH"
	EnvFirstUserEmail      = "FIRSTUSER__CREDENTIALS__EMAIL"
	EnvFirstUserPassword   = "FIRSTUSER__CREDENTIALS__PASSWORD"
	EnvSessionTTL          = "LOGIN__SESSION__TTL"
	EnvPasswordMaxAttempts = "LOGIN__PASSWORD__MAX_ATTEMPTS"
	EnvPasswordLockout     = "LOGIN__PASSWORD__LOCKOUT"
	EnvGoogleClientID      = "LOGIN__OAUTH__GOOGLE__CLIENT_ID"
	EnvGoogleClientSecret  = "LOGIN__OAUTH__GOOGLE__CLIENT_SECRET"
	EnvAuditLogMaxItems    = "AUDITLOG__RETENTION__LEVELDB__MAX_ITEMS"
	EnvMailgunAPIKey       = "MAILING__SENDING__MAILGUN__API_KEY"
	EnvMailgunBaseURL      = "MAILING__SENDING__MAILGUN__BASE_URL"
	EnvMailgunDomain       = "MAILING__SENDING__MAILGUN__DOMAIN"
	EnvMailgunFrom         = "MAILING__SENDING__MAILGUN__FROM"
)

type Config struct {
	HTTP      HTTP
	Storage   Storage
	FirstUser FirstUser
	Login     Login
	AuditLog  AuditLog
	Mailing   Mailing
}

type HTTP struct {
	Port    int
	BaseURL string
}

const DefaultBaseURL = "http://127.0.0.1:41783"

func (h HTTP) BasePath() string {
	parsed, _ := url.Parse(h.BaseURL)
	return strings.TrimRight(parsed.Path, "/")
}

type Storage struct {
	LevelDBPath string
}

type FirstUser struct {
	Email    string
	Password string
}

type Login struct {
	SessionTTL          time.Duration
	PasswordMaxAttempts int
	PasswordLockout     time.Duration
	Google              GoogleOAuth
}

type GoogleOAuth struct {
	ClientID     string
	ClientSecret string
}

const DefaultAuditLogMaxItems = 10_000

type AuditLog struct {
	MaxItems int
}

type Mailing struct {
	Mailgun Mailgun
}

type Mailgun struct {
	APIKey  string
	BaseURL string
	Domain  string
	From    string
}

func (m Mailgun) Enabled() bool {
	return m.APIKey != ""
}

func (g GoogleOAuth) Enabled() bool {
	return g.ClientID != ""
}

func Load() (Config, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(EnvHTTPBaseURL)), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
		slog.Warn("HTTP base URL is not configured; using local URL", "environment", EnvHTTPBaseURL, "baseURL", baseURL)
	}
	port, err := envInt(EnvHTTPPort, 80)
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := envDuration(EnvSessionTTL, 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	maxAttempts, err := envInt(EnvPasswordMaxAttempts, 5)
	if err != nil {
		return Config{}, err
	}
	lockout, err := envDuration(EnvPasswordLockout, 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	auditLogMaxItems, err := envInt(EnvAuditLogMaxItems, DefaultAuditLogMaxItems)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTP{
			Port:    port,
			BaseURL: baseURL,
		},
		Storage: Storage{LevelDBPath: envString(EnvStorageLevelDBPath, "./data")},
		FirstUser: FirstUser{
			Email:    strings.TrimSpace(os.Getenv(EnvFirstUserEmail)),
			Password: os.Getenv(EnvFirstUserPassword),
		},
		Login: Login{
			SessionTTL:          sessionTTL,
			PasswordMaxAttempts: maxAttempts,
			PasswordLockout:     lockout,
			Google: GoogleOAuth{
				ClientID:     strings.TrimSpace(os.Getenv(EnvGoogleClientID)),
				ClientSecret: os.Getenv(EnvGoogleClientSecret),
			},
		},
		AuditLog: AuditLog{MaxItems: auditLogMaxItems},
		Mailing: Mailing{Mailgun: Mailgun{
			APIKey:  os.Getenv(EnvMailgunAPIKey),
			BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv(EnvMailgunBaseURL)), "/"),
			Domain:  strings.TrimSpace(os.Getenv(EnvMailgunDomain)),
			From:    strings.TrimSpace(os.Getenv(EnvMailgunFrom)),
		}},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", EnvHTTPPort)
	}
	baseURL, err := url.Parse(c.HTTP.BaseURL)
	if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return fmt.Errorf("%s must be a full HTTP or HTTPS URL without a query or fragment", EnvHTTPBaseURL)
	}
	if strings.TrimSpace(c.Storage.LevelDBPath) == "" {
		return fmt.Errorf("%s cannot be empty", EnvStorageLevelDBPath)
	}
	if (c.FirstUser.Email == "") != (c.FirstUser.Password == "") {
		return fmt.Errorf("%s and %s must be set together", EnvFirstUserEmail, EnvFirstUserPassword)
	}
	if c.Login.SessionTTL <= 0 {
		return fmt.Errorf("%s must be a positive duration", EnvSessionTTL)
	}
	if c.Login.PasswordMaxAttempts < 1 {
		return fmt.Errorf("%s must be at least 1", EnvPasswordMaxAttempts)
	}
	if c.Login.PasswordLockout <= 0 {
		return fmt.Errorf("%s must be a positive duration", EnvPasswordLockout)
	}
	if c.AuditLog.MaxItems < 1 {
		return fmt.Errorf("%s must be at least 1", EnvAuditLogMaxItems)
	}
	google := c.Login.Google
	if (google.ClientID == "") != (google.ClientSecret == "") {
		return errors.New("LOGIN__OAUTH__GOOGLE client ID and client secret must be set together")
	}
	mailgun := c.Mailing.Mailgun
	configuredMailgunValues := 0
	for _, value := range []string{mailgun.APIKey, mailgun.BaseURL, mailgun.Domain, mailgun.From} {
		if value != "" {
			configuredMailgunValues++
		}
	}
	if configuredMailgunValues != 0 && configuredMailgunValues != 4 {
		return fmt.Errorf("%s, %s, %s, and %s must be set together", EnvMailgunAPIKey, EnvMailgunBaseURL, EnvMailgunDomain, EnvMailgunFrom)
	}
	if mailgun.Enabled() {
		baseURL, err := url.Parse(mailgun.BaseURL)
		if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" || (baseURL.Path != "" && baseURL.Path != "/") || baseURL.RawQuery != "" || baseURL.Fragment != "" {
			return fmt.Errorf("%s must be a full HTTP or HTTPS origin without a path, query, or fragment", EnvMailgunBaseURL)
		}
	}
	return nil
}

func envString(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envInt(name string, fallback int) (int, error) {
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a Go duration such as 10s or 24h: %w", name, err)
	}
	return parsed, nil
}
