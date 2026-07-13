package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/content"
	"github.com/letitcall/letitcall/api/internal/mailing"
	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"github.com/letitcall/letitcall/api/internal/web"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	sessionCookieName     = "letitcall_session"
	portalBasePlaceholder = "/__LETITCALL_BASE_PATH__"
)

type contextKey string

const userContextKey contextKey = "authenticated-user"

type Server struct {
	cfg         config.Config
	store       *store.Store
	avatars     *content.Avatars
	logos       *content.Logos
	oauth       *oauth2.Config
	tokenCipher *security.TokenCipher
	limiter     *security.LoginLimiter
	mailer      mailing.Sender
	dummyHash   string
	now         func() time.Time
	webhookHTTP *http.Client
	webhookWake chan struct{}
}

func New(cfg config.Config, database *store.Store) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate API configuration: %w", err)
	}
	dummyHash, err := security.HashPassword("not-a-real-password")
	if err != nil {
		return nil, err
	}
	avatars, err := content.NewAvatars(cfg.Storage.LevelDBPath)
	if err != nil {
		return nil, err
	}
	logos, err := content.NewLogos(cfg.Storage.LevelDBPath)
	if err != nil {
		return nil, err
	}
	renderer, err := mailing.NewRenderer()
	if err != nil {
		return nil, err
	}
	server := &Server{
		cfg:         cfg,
		store:       database,
		avatars:     avatars,
		logos:       logos,
		mailer:      mailing.New(cfg.Mailing.Mailgun, renderer),
		limiter:     security.NewLoginLimiter(cfg.Login.PasswordMaxAttempts, cfg.Login.PasswordLockout),
		dummyHash:   dummyHash,
		now:         time.Now,
		webhookHTTP: &http.Client{Timeout: 10 * time.Second},
		webhookWake: make(chan struct{}, 1),
	}
	key, err := security.LoadGoogleTokenKey(cfg.Storage.LevelDBPath)
	if err != nil {
		return nil, err
	}
	server.tokenCipher, err = security.NewTokenCipher(key)
	if err != nil {
		return nil, err
	}
	if cfg.Login.Google.Enabled() {
		server.oauth = &oauth2.Config{
			ClientID:     cfg.Login.Google.ClientID,
			ClientSecret: cfg.Login.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"openid",
				"email",
				"profile",
				"https://www.googleapis.com/auth/calendar.events",
			},
		}
	}
	return server, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/config/public", s.publicConfig)
	mux.HandleFunc("GET "+compatibilityAPIPath+"/openapi.json", s.openAPISpecification)
	swagger := s.swaggerHandler()
	mux.Handle("GET "+compatibilityAPIPath+"/swagger/", swagger)
	mux.Handle("GET "+compatibilityAPIPath+"/swagger/{asset}", swagger)
	mux.Handle("GET "+compatibilityAPIPath+"/users/me", s.requireAPIToken(http.HandlerFunc(s.compatibilityCurrentUser)))
	mux.Handle("GET "+compatibilityAPIPath+"/event_types", s.requireAPIToken(http.HandlerFunc(s.compatibilityListEventTypes)))
	mux.Handle("GET "+compatibilityAPIPath+"/event_types/{uuid}", s.requireAPIToken(http.HandlerFunc(s.compatibilityGetEventType)))
	mux.Handle("GET "+compatibilityAPIPath+"/event_type_available_times", s.requireAPIToken(http.HandlerFunc(s.compatibilityAvailableTimes)))
	mux.Handle("POST "+compatibilityAPIPath+"/webhook_subscriptions", s.requireAPIToken(http.HandlerFunc(s.createWebhookSubscription)))
	mux.Handle("GET "+compatibilityAPIPath+"/webhook_subscriptions", s.requireAPIToken(http.HandlerFunc(s.listWebhookSubscriptions)))
	mux.Handle("GET "+compatibilityAPIPath+"/scheduled_events", s.requireAPIToken(http.HandlerFunc(s.compatibilityListScheduledEvents)))
	mux.Handle("GET "+compatibilityAPIPath+"/scheduled_events/{event_uuid}/invitees", s.requireAPIToken(http.HandlerFunc(s.compatibilityListInvitees)))
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("GET /api/auth/google/start", s.googleStart)
	mux.HandleFunc("POST "+googleAPICallbackPath, s.googleCallback)
	mux.Handle("GET /api/auth/session", s.requireAuth(http.HandlerFunc(s.session)))
	mux.Handle("POST /api/auth/logout", s.requireAuth(http.HandlerFunc(s.logout)))
	mux.Handle("GET /api/branding", s.requireAuth(http.HandlerFunc(s.getBranding)))
	mux.Handle("PUT /api/branding", s.requireAuth(http.HandlerFunc(s.updateBranding)))
	mux.Handle("GET /api/integration", s.requireAuth(http.HandlerFunc(s.getAPIIntegration)))
	mux.Handle("POST /api/integration/tokens", s.requireAuth(http.HandlerFunc(s.createAPIToken)))
	mux.Handle("DELETE /api/integration/tokens/{id}", s.requireAuth(http.HandlerFunc(s.deleteAPIToken)))
	mux.Handle("GET /api/users", s.requireAuth(http.HandlerFunc(s.listUsers)))
	mux.Handle("POST /api/users", s.requireAuth(http.HandlerFunc(s.createUser)))
	mux.Handle("PATCH /api/users/{email}", s.requireAuth(http.HandlerFunc(s.updateUser)))
	mux.Handle("GET /api/users/{email}/deletion-impact", s.requireAuth(http.HandlerFunc(s.getUserDeletionImpact)))
	mux.Handle("POST /api/users/{email}/reassign-bookings", s.requireAuth(http.HandlerFunc(s.reassignUserBookings)))
	mux.Handle("DELETE /api/users/{email}", s.requireAuth(http.HandlerFunc(s.deleteUser)))
	mux.Handle("GET /api/bookings", s.requireAuth(http.HandlerFunc(s.listBookings)))
	mux.HandleFunc("POST /api/bookings", s.createBooking)
	mux.Handle("DELETE /api/bookings/{id}", s.requireAuth(http.HandlerFunc(s.deleteBooking)))
	mux.Handle("GET /api/events/{secret}", s.optionalAuth(http.HandlerFunc(s.getManagedBooking)))
	mux.Handle("PATCH /api/events/{secret}", s.optionalAuth(http.HandlerFunc(s.updateManagedBooking)))
	mux.Handle("POST /api/events/{secret}/cancel", s.optionalAuth(http.HandlerFunc(s.cancelManagedBooking)))
	mux.Handle("GET /api/event-types", s.requireAuth(http.HandlerFunc(s.listEventTypes)))
	mux.Handle("POST /api/event-types", s.requireAuth(http.HandlerFunc(s.createEventType)))
	mux.Handle("GET /api/event-types/{slug}", s.requireAuth(http.HandlerFunc(s.getEventType)))
	mux.Handle("PUT /api/event-types/{slug}", s.requireAuth(http.HandlerFunc(s.updateEventType)))
	mux.Handle("DELETE /api/event-types/{slug}", s.requireAuth(http.HandlerFunc(s.deleteEventType)))
	mux.HandleFunc("GET /api/public/event-types/{slug}", s.getPublicEventType)
	mux.HandleFunc("GET /content/avatars/{filename}", s.serveAvatar)
	mux.HandleFunc("GET /content/logos/{filename}", s.serveLogo)
	mux.HandleFunc("/content/", http.NotFound)
	mux.HandleFunc("/", s.servePortal)
	handler := s.middleware(mux)
	basePath := s.cfg.HTTP.BasePath()
	if basePath == "" {
		return handler
	}
	mounted := http.NewServeMux()
	mounted.Handle(basePath+"/", http.StripPrefix(basePath, handler))
	return mounted
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic serving request", "error", recovered, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		session, err := s.store.GetSession(cookie.Value, s.now())
		if err != nil {
			s.clearCookie(w, r)
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := s.store.GetUser(session.Email)
		if err != nil {
			_ = s.store.DeleteSession(cookie.Value)
			s.clearCookie(w, r)
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !s.validOrigin(r) {
			writeError(w, http.StatusForbidden, "request origin is not allowed")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) requireAPIToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		token, err := s.store.GetAPIToken(security.TokenDigest(parts[1]))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := s.store.GetUser(token.UserEmail)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) optionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !s.validOrigin(r) {
			writeError(w, http.StatusForbidden, "request origin is not allowed")
			return
		}
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		session, err := s.store.GetSession(cookie.Value, s.now())
		if err != nil {
			s.clearCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}
		user, err := s.store.GetUser(session.Email)
		if err != nil {
			_ = s.store.DeleteSession(cookie.Value)
			s.clearCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) validOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return origin == requestOrigin(r)
}

func (s *Server) servePortal(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "API endpoint not found")
		return
	}
	assetPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if assetPath == "." || assetPath == "" {
		assetPath = "index.html"
	}
	if info, err := fs.Stat(web.Assets, assetPath); err != nil || info.IsDir() {
		assetPath = "index.html"
	}
	contents, err := fs.ReadFile(web.Assets, assetPath)
	if err != nil {
		internalError(w, err, "read portal asset")
		return
	}
	contents = bytes.ReplaceAll(contents, []byte(portalBasePlaceholder), []byte(s.cfg.HTTP.BasePath()))
	if assetPath == "index.html" {
		branding, err := s.store.GetBranding()
		if err != nil {
			internalError(w, err, "load branding")
			return
		}
		contents = bytes.ReplaceAll(contents, []byte(model.DefaultBrandName), []byte(html.EscapeString(branding.Name)))
	}
	if contentType := mime.TypeByExtension(path.Ext(assetPath)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	_, _ = w.Write(contents)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) publicConfig(w http.ResponseWriter, _ *http.Request) {
	branding, err := s.store.GetBranding()
	if err != nil {
		internalError(w, err, "load branding")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"brandName":          branding.Name,
		"logoPath":           branding.LogoPath,
		"googleLoginEnabled": s.cfg.Login.Google.Enabled(),
	})
}

func userFromRequest(r *http.Request) model.User {
	return r.Context().Value(userContextKey).(model.User)
}

func authenticatedUser(r *http.Request) (model.User, bool) {
	user, ok := r.Context().Value(userContextKey).(model.User)
	return user, ok
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("write JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	if mediaType := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0])); mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return errors.New("invalid content type")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return errors.New("multiple or malformed trailing JSON values")
	}
	return nil
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, value string, expires time.Time, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     s.cookiePath(),
		Expires:  expires,
		MaxAge:   max(1, maxAge),
		HttpOnly: true,
		Secure:   requestScheme(r) == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) clearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Path:     s.cookiePath(),
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   requestScheme(r) == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) cookiePath() string {
	basePath := s.cfg.HTTP.BasePath()
	if basePath == "" {
		return "/"
	}
	return basePath
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request, email string) error {
	token, err := security.RandomToken(32)
	if err != nil {
		return err
	}
	expires := s.now().UTC().Add(s.cfg.Login.SessionTTL)
	if err := s.store.PutSession(token, model.Session{Email: email, ExpiresAt: expires}); err != nil {
		return err
	}
	s.setSessionCookie(w, r, token, expires, int(s.cfg.Login.SessionTTL.Seconds()))
	return nil
}

func requestOrigin(r *http.Request) string {
	return requestScheme(r) + "://" + r.Host
}

func requestScheme(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		return forwarded
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func internalError(w http.ResponseWriter, err error, operation string) {
	slog.Error(operation, "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func bookingKey(value string) (string, time.Time, error) {
	if strings.Contains(value, ".") {
		return "", time.Time{}, errors.New("booking time must not contain milliseconds")
	}
	if !strings.HasSuffix(value, "Z") {
		return "", time.Time{}, errors.New("booking time must use UTC")
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("booking time must use RFC3339: %w", err)
	}
	utc := parsed.UTC().Truncate(time.Second)
	if utc.Format(time.RFC3339) != value {
		return "", time.Time{}, errors.New("booking time must use UTC RFC3339 seconds")
	}
	return utc.Format(time.RFC3339), utc, nil
}
