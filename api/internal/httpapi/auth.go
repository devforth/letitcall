package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.validOrigin(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(request.Email))
	limitKey := remoteIP(r) + "|" + normalizedEmail
	now := s.now()
	if !s.limiter.Allowed(limitKey, now) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts; try again later")
		return
	}

	user, err := s.store.GetUser(normalizedEmail)
	hash := s.dummyHash
	if err == nil {
		hash = user.PasswordHash
	} else if !errors.Is(err, store.ErrNotFound) {
		internalError(w, err, "load user for login")
		return
	}
	if err != nil || !security.CheckPassword(hash, request.Password) {
		s.limiter.Failure(limitKey, now)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	s.limiter.Success(limitKey)
	if err := s.createSession(w, user.Email); err != nil {
		internalError(w, err, "create session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user.Public()})
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": userFromRequest(r).Public()})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if err := s.store.DeleteSession(cookie.Value); err != nil {
			internalError(w, err, "delete session")
			return
		}
	}
	clearCookie(w, s.cfg.Login.SessionCookieSecure)
	w.WriteHeader(http.StatusNoContent)
}
