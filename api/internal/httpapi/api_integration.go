package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

type apiTokenSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Server) getAPIIntegration(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.store.ListAPITokens(userFromRequest(r).Email)
	if err != nil {
		internalError(w, err, "list API tokens")
		return
	}
	summaries := make([]apiTokenSummary, 0, len(tokens))
	for _, token := range tokens {
		summaries = append(summaries, summarizeAPIToken(token))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"baseURL":    s.compatibilityBaseURL(),
		"openAPIURL": s.compatibilityBaseURL() + "/openapi.json",
		"swaggerURL": s.compatibilityBaseURL() + "/swagger/",
		"tokens":     summaries,
	})
}

type createAPITokenRequest struct {
	Name string `json:"name"`
}

func (s *Server) createAPIToken(w http.ResponseWriter, r *http.Request) {
	var request createAPITokenRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	name := strings.TrimSpace(request.Name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		writeError(w, http.StatusBadRequest, "name must be between 1 and 100 characters")
		return
	}
	random, err := security.RandomToken(32)
	if err != nil {
		internalError(w, err, "generate API token")
		return
	}
	value := "lic_" + random
	token := model.APIToken{
		ID:        security.TokenDigest(value),
		Name:      name,
		UserEmail: userFromRequest(r).Email,
		CreatedAt: s.now().UTC().Truncate(time.Second),
	}
	if err := s.store.CreateAPIToken(token); err != nil {
		internalError(w, err, "store API token")
		return
	}
	if err := s.recordAuditLog(r, "generated_token", "api_token", token.ID, summarizeAPIToken(token)); err != nil {
		internalError(w, err, "record API token generation audit log")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"apiToken": summarizeAPIToken(token), "token": value})
}

func (s *Server) deleteAPIToken(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token, err := s.store.GetAPIToken(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "API token not found")
		return
	}
	if err != nil {
		internalError(w, err, "load API token for deletion")
		return
	}
	err = s.store.DeleteAPIToken(id, userFromRequest(r).Email)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "API token not found")
		return
	}
	if err != nil {
		internalError(w, err, "delete API token")
		return
	}
	if err := s.recordAuditLog(r, "revoked_token", "api_token", id, summarizeAPIToken(token)); err != nil {
		internalError(w, err, "record API token revocation audit log")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func summarizeAPIToken(token model.APIToken) apiTokenSummary {
	return apiTokenSummary{ID: token.ID, Name: token.Name, CreatedAt: token.CreatedAt}
}
