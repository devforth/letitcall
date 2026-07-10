package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"golang.org/x/oauth2"
)

const (
	googleUserInfoURL  = "https://openidconnect.googleapis.com/v1/userinfo"
	googleCallbackPath = "/api/auth/google/callback"
)

func (s *Server) googleStart(w http.ResponseWriter, r *http.Request) {
	if s.oauth == nil {
		writeError(w, http.StatusNotFound, "Google login is not enabled")
		return
	}
	state, err := security.RandomToken(32)
	if err != nil {
		internalError(w, err, "generate OAuth state")
		return
	}
	verifier, err := security.RandomToken(48)
	if err != nil {
		internalError(w, err, "generate OAuth code verifier")
		return
	}
	if err := s.store.PutOAuthState(state, model.OAuthState{
		CodeVerifier: verifier,
		ExpiresAt:    s.now().UTC().Add(10 * time.Minute),
	}); err != nil {
		internalError(w, err, "store OAuth state")
		return
	}
	challenge := sha256.Sum256([]byte(verifier))
	oauthConfig := s.googleOAuthConfig(r)
	authorizationURL := oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:])),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, authorizationURL, http.StatusFound)
}

func (s *Server) googleCallback(w http.ResponseWriter, r *http.Request) {
	if s.oauth == nil {
		writeError(w, http.StatusNotFound, "Google login is not enabled")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		writeError(w, http.StatusBadRequest, "Google authorization was not completed")
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		writeError(w, http.StatusBadRequest, "OAuth state and code are required")
		return
	}
	storedState, err := s.store.ConsumeOAuthState(state, s.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "OAuth state is invalid or expired")
		return
	}
	oauthConfig := s.googleOAuthConfig(r)
	token, err := oauthConfig.Exchange(r.Context(), code, oauth2.VerifierOption(storedState.CodeVerifier))
	if err != nil {
		writeError(w, http.StatusBadGateway, "Google token exchange failed")
		return
	}
	identity, err := fetchGoogleIdentity(r.Context(), oauthConfig, token)
	if err != nil {
		internalError(w, err, "fetch Google identity")
		return
	}
	if !identity.EmailVerified {
		writeError(w, http.StatusForbidden, "Google email must be verified")
		return
	}
	email, err := security.NormalizeEmail(identity.Email)
	if err != nil {
		writeError(w, http.StatusForbidden, "Google did not return a valid email")
		return
	}
	user, err := s.store.GetUser(email)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusForbidden, "this Google account is not an existing user")
		return
	}
	if err != nil {
		internalError(w, err, "load OAuth user")
		return
	}

	if token.RefreshToken == "" && user.EncryptedGoogleToken != "" {
		if previousJSON, decryptErr := s.tokenCipher.Decrypt(user.EncryptedGoogleToken); decryptErr == nil {
			var previous oauth2.Token
			if json.Unmarshal(previousJSON, &previous) == nil {
				token.RefreshToken = previous.RefreshToken
			}
		}
	}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		internalError(w, err, "encode Google token")
		return
	}
	encryptedToken, err := s.tokenCipher.Encrypt(tokenJSON)
	if err != nil {
		internalError(w, err, "encrypt Google token")
		return
	}
	user.GoogleConnected = true
	user.EncryptedGoogleToken = encryptedToken
	user.UpdatedAt = s.now().UTC().Truncate(time.Second)
	if err := s.store.PutUser(user); err != nil {
		internalError(w, err, "save Google connection")
		return
	}
	if err := s.createSession(w, r, user.Email); err != nil {
		internalError(w, err, "create OAuth session")
		return
	}
	http.Redirect(w, r, s.cfg.HTTP.BasePath+"/", http.StatusSeeOther)
}

func (s *Server) googleOAuthConfig(r *http.Request) *oauth2.Config {
	config := *s.oauth
	config.RedirectURL = requestOrigin(r) + s.cfg.HTTP.BasePath + googleCallbackPath
	return &config
}

type googleIdentity struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func fetchGoogleIdentity(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (googleIdentity, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return googleIdentity{}, err
	}
	response, err := oauthConfig.Client(ctx, token).Do(request)
	if err != nil {
		return googleIdentity{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return googleIdentity{}, fmt.Errorf("Google userinfo returned %s", response.Status)
	}
	var identity googleIdentity
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&identity); err != nil {
		return googleIdentity{}, err
	}
	return identity, nil
}
