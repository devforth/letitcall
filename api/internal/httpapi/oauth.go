package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/content"
	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"golang.org/x/oauth2"
)

const (
	googleUserInfoURL        = "https://openidconnect.googleapis.com/v1/userinfo"
	googlePortalCallbackPath = "/auth/google/callback"
	googleAPICallbackPath    = "/api/auth/google/callback"
)

type googleCallbackRequest struct {
	State string `json:"state"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

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
	if !s.validOrigin(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	var request googleCallbackRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	if request.Error != "" {
		writeError(w, http.StatusBadRequest, "Google authorization was not completed")
		return
	}
	state := request.State
	code := request.Code
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
		writeError(w, http.StatusForbidden, "This Google account does not belong to an existing user. Ask an administrator to add your email before signing in.")
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
	var googleAvatar content.Avatar
	if user.FullName == "" {
		user.FullName = strings.TrimSpace(identity.Name)
	}
	if user.AvatarPath == "" && identity.Picture != "" {
		googleAvatar, err = fetchGoogleAvatar(r.Context(), identity.Picture, user.Email, s.avatars)
		if err != nil {
			internalError(w, err, "fetch Google avatar")
			return
		}
		user.AvatarPath = googleAvatar.Filename
	}
	user.GoogleConnected = true
	user.EncryptedGoogleToken = encryptedToken
	user.UpdatedAt = s.now().UTC().Truncate(time.Second)
	if googleAvatar.Filename != "" {
		if err := s.avatars.Write(googleAvatar); err != nil {
			internalError(w, err, "store Google avatar")
			return
		}
	}
	if err := s.store.PutUser(user); err != nil {
		internalError(w, err, "save Google connection")
		return
	}
	if err := s.createSession(w, r, user.Email); err != nil {
		internalError(w, err, "create OAuth session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user.Public()})
}

func (s *Server) googleOAuthConfig(r *http.Request) *oauth2.Config {
	config := *s.oauth
	config.RedirectURL = s.cfg.HTTP.BaseURL + googlePortalCallbackPath
	return &config
}

type googleIdentity struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
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

func fetchGoogleAvatar(ctx context.Context, pictureURL, email string, avatars *content.Avatars) (content.Avatar, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return content.Avatar{}, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return content.Avatar{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return content.Avatar{}, fmt.Errorf("Google avatar returned %s", response.Status)
	}
	source, _, err := image.Decode(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return content.Avatar{}, fmt.Errorf("decode Google avatar: %w", err)
	}
	return avatars.PrepareImage(email, source)
}
