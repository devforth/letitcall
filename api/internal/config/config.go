package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvHTTPPort                 = "HTTP__PORT"
	EnvHTTPPublicURL            = "HTTP__PUBLIC__URL"
	EnvHTTPReadTimeout          = "HTTP__READ__TIMEOUT"
	EnvHTTPWriteTimeout         = "HTTP__WRITE__TIMEOUT"
	EnvHTTPIdleTimeout          = "HTTP__IDLE__TIMEOUT"
	EnvHTTPShutdownTimeout      = "HTTP__SHUTDOWN__TIMEOUT"
	EnvStorageLevelDBPath       = "STORAGE__LEVELDB__PATH"
	EnvFirstUserEmail           = "FIRSTUSER__CREDENTIALS__EMAIL"
	EnvFirstUserPassword        = "FIRSTUSER__CREDENTIALS__PASSWORD"
	EnvSessionTTL               = "LOGIN__SESSION__TTL"
	EnvSessionCookieSecure      = "LOGIN__SESSION__COOKIE__SECURE"
	EnvPasswordMaxAttempts      = "LOGIN__PASSWORD__MAX_ATTEMPTS"
	EnvPasswordLockout          = "LOGIN__PASSWORD__LOCKOUT"
	EnvGoogleClientID           = "LOGIN__OAUTH__GOOGLE__CLIENT_ID"
	EnvGoogleClientSecret       = "LOGIN__OAUTH__GOOGLE__CLIENT_SECRET"
	EnvGoogleRedirectURL        = "LOGIN__OAUTH__GOOGLE__REDIRECT_URL"
	EnvGoogleTokenEncryptionKey = "LOGIN__OAUTH__GOOGLE__TOKEN_ENCRYPTION_KEY"
)

type Config struct {
	HTTP      HTTP
	Storage   Storage
	FirstUser FirstUser
	Login     Login
}

type HTTP struct {
	Port            int
	PublicURL       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
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
	SessionCookieSecure bool
	PasswordMaxAttempts int
	PasswordLockout     time.Duration
	Google              GoogleOAuth
}

type GoogleOAuth struct {
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	TokenEncryptionKey string
}

func (g GoogleOAuth) Enabled() bool {
	return g.ClientID != ""
}

func Load() (Config, error) {
	publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv(EnvHTTPPublicURL)), "/")
	port, err := envInt(EnvHTTPPort, 80)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := envDuration(EnvHTTPReadTimeout, 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envDuration(EnvHTTPWriteTimeout, 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := envDuration(EnvHTTPIdleTimeout, 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := envDuration(EnvHTTPShutdownTimeout, 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := envDuration(EnvSessionTTL, 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := envBool(EnvSessionCookieSecure, strings.HasPrefix(publicURL, "https://"))
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

	redirectURL := strings.TrimSpace(os.Getenv(EnvGoogleRedirectURL))
	if redirectURL == "" && publicURL != "" {
		redirectURL = publicURL + "/api/auth/google/callback"
	}

	cfg := Config{
		HTTP: HTTP{
			Port:            port,
			PublicURL:       publicURL,
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Storage: Storage{LevelDBPath: envString(EnvStorageLevelDBPath, "./data")},
		FirstUser: FirstUser{
			Email:    strings.TrimSpace(os.Getenv(EnvFirstUserEmail)),
			Password: os.Getenv(EnvFirstUserPassword),
		},
		Login: Login{
			SessionTTL:          sessionTTL,
			SessionCookieSecure: cookieSecure,
			PasswordMaxAttempts: maxAttempts,
			PasswordLockout:     lockout,
			Google: GoogleOAuth{
				ClientID:           strings.TrimSpace(os.Getenv(EnvGoogleClientID)),
				ClientSecret:       os.Getenv(EnvGoogleClientSecret),
				RedirectURL:        redirectURL,
				TokenEncryptionKey: strings.TrimSpace(os.Getenv(EnvGoogleTokenEncryptionKey)),
			},
		},
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
	if c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.IdleTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("HTTP timeout settings must be positive durations")
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
	if c.HTTP.PublicURL != "" {
		parsed, err := url.Parse(c.HTTP.PublicURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && !(parsed.Scheme == "http" && isLocalhost(parsed.Hostname()))) {
			return fmt.Errorf("%s must be an HTTPS URL (HTTP is allowed only for localhost)", EnvHTTPPublicURL)
		}
	}

	google := c.Login.Google
	googleValues := []string{google.ClientID, google.ClientSecret, google.RedirectURL, google.TokenEncryptionKey}
	googleSet := 0
	for _, value := range googleValues {
		if value != "" {
			googleSet++
		}
	}
	if googleSet != 0 && googleSet != len(googleValues) {
		return errors.New("all LOGIN__OAUTH__GOOGLE settings must be set when Google OAuth is enabled")
	}
	if google.Enabled() {
		redirect, err := url.Parse(google.RedirectURL)
		if err != nil || redirect.Host == "" || (redirect.Scheme != "https" && !(redirect.Scheme == "http" && isLocalhost(redirect.Hostname()))) {
			return fmt.Errorf("%s must be an HTTPS URL (HTTP is allowed only for localhost)", EnvGoogleRedirectURL)
		}
		key, err := decodeEncryptionKey(google.TokenEncryptionKey)
		if err != nil || len(key) != 32 {
			return fmt.Errorf("%s must be a base64-encoded 32-byte key", EnvGoogleTokenEncryptionKey)
		}
	}
	return nil
}

func DecodeGoogleTokenEncryptionKey(value string) ([]byte, error) {
	key, err := decodeEncryptionKey(value)
	if err != nil {
		return nil, fmt.Errorf("decode Google token encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("Google token encryption key must be 32 bytes")
	}
	return key, nil
}

func decodeEncryptionKey(value string) ([]byte, error) {
	key, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return key, nil
	}
	return base64.StdEncoding.DecodeString(value)
}

func isLocalhost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
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

func envBool(name string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
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
