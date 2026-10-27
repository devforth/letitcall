package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	mcpPath                        = "/api/mcp"
	mcpConsentPagePath             = "/auth/connect-ai"
	mcpAuthorizationRequestTTL     = 10 * time.Minute
	mcpAuthorizationCodeTTL        = 5 * time.Minute
	mcpAccessTokenTTL              = time.Hour
	mcpRefreshTokenTTL             = 7 * 24 * time.Hour
	mcpAuthorizationRequestPrefix  = "request/"
	mcpAuthorizationCodePrefix     = "code/"
	mcpProtectedResourceMetadata   = mcpPath + "/oauth-protected-resource"
	mcpAuthorizationServerMetadata = mcpPath + "/.well-known/openid-configuration"
	mcpJWKSPath                    = mcpPath + "/oauth/jwks"
	mcpAuthorizePath               = mcpPath + "/oauth/authorize"
	mcpTokenPath                   = mcpPath + "/oauth/token"
)

var errInvalidGrant = errors.New("invalid_grant")

// mcpURL is the MCP resource and the OAuth issuer: OAuth accepts tokens only for this exact URL.
func (s *Server) mcpURL() string {
	return s.cfg.HTTP.BaseURL + mcpPath
}

func (s *Server) mcpProtectedResourceMetadata(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, oauthex.ProtectedResourceMetadata{
		Resource:               s.mcpURL(),
		AuthorizationServers:   []string{s.mcpURL()},
		BearerMethodsSupported: []string{"header"},
	})
}

func (s *Server) mcpAuthorizationServerMetadata(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		oauthex.AuthServerMeta
		// Required by OpenID Connect Discovery, which clients validate this document against. No ID tokens are
		// issued, so the key set is empty.
		SubjectTypesSupported            []string `json:"subject_types_supported"`
		IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	}{
		AuthServerMeta: oauthex.AuthServerMeta{
			Issuer:                                     s.mcpURL(),
			AuthorizationEndpoint:                      s.cfg.HTTP.BaseURL + mcpAuthorizePath,
			TokenEndpoint:                              s.cfg.HTTP.BaseURL + mcpTokenPath,
			JWKSURI:                                    s.cfg.HTTP.BaseURL + mcpJWKSPath,
			ResponseTypesSupported:                     []string{"code"},
			GrantTypesSupported:                        []string{"authorization_code", "refresh_token"},
			CodeChallengeMethodsSupported:              []string{"S256"},
			TokenEndpointAuthMethodsSupported:          []string{"none"},
			ClientIDMetadataDocumentSupported:          true,
			AuthorizationResponseIssParameterSupported: true,
		},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
	})
}

func (s *Server) mcpJWKS(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"keys": []any{}})
}

// mcpAuthorize validates the authorization request and sends the browser to the consent page. It needs no sign-in,
// so the client metadata document is fetched only once a signed-in user opens the consent page.
func (s *Server) mcpAuthorize(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("response_type") != "code" || query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" {
		writeError(w, http.StatusBadRequest, "response_type must be code with an S256 code_challenge")
		return
	}
	if resource := query.Get("resource"); resource != "" && resource != s.mcpURL() {
		writeError(w, http.StatusBadRequest, "resource must be "+s.mcpURL())
		return
	}
	redirectURL, err := url.Parse(query.Get("redirect_uri"))
	if err != nil || !isAllowedRedirectURI(redirectURL) {
		writeError(w, http.StatusBadRequest, "redirect_uri must use https unless it points to localhost")
		return
	}
	if !isClientIDURL(query.Get("client_id")) {
		writeError(w, http.StatusBadRequest, "client_id must be the https URL of a client metadata document")
		return
	}
	requestID, err := security.RandomToken(32)
	if err != nil {
		internalError(w, err, "generate MCP authorization request ID")
		return
	}
	if err := s.store.PutMCPAuthorization(mcpAuthorizationRequestPrefix+requestID, model.MCPAuthorization{
		ClientID:      query.Get("client_id"),
		RedirectURI:   query.Get("redirect_uri"),
		CodeChallenge: query.Get("code_challenge"),
		State:         query.Get("state"),
		ExpiresAt:     s.now().UTC().Add(mcpAuthorizationRequestTTL).Truncate(time.Second),
	}); err != nil {
		internalError(w, err, "store MCP authorization request")
		return
	}
	http.Redirect(w, r, s.cfg.HTTP.BaseURL+mcpConsentPagePath+"?request="+url.QueryEscape(requestID), http.StatusFound)
}

// verifiedMCPAuthorization checks the requested redirect_uri against the client metadata document.
func (s *Server) verifiedMCPAuthorization(w http.ResponseWriter, r *http.Request, authorization model.MCPAuthorization) (mcpClient, *url.URL, bool) {
	client, err := fetchMCPClient(r.Context(), authorization.ClientID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return mcpClient{}, nil, false
	}
	redirectURL, _ := url.Parse(authorization.RedirectURI)
	for _, registered := range client.RedirectURIs {
		if redirectURIMatches(redirectURL, registered) {
			return client, redirectURL, true
		}
	}
	writeError(w, http.StatusBadRequest, "redirect_uri is not listed in the client metadata document")
	return mcpClient{}, nil, false
}

func (s *Server) describeMCPAuthorization(w http.ResponseWriter, r *http.Request) {
	authorization, err := s.store.GetMCPAuthorization(mcpAuthorizationRequestPrefix+r.URL.Query().Get("request"), s.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "authorization request is invalid or expired, start connecting again")
		return
	}
	client, redirectURL, ok := s.verifiedMCPAuthorization(w, r, authorization)
	if !ok {
		return
	}
	clientURL, _ := url.Parse(client.ClientID)
	writeJSON(w, http.StatusOK, map[string]any{
		"clientName":       client.ClientName,
		"clientHost":       clientURL.Host,
		"redirectHost":     redirectURL.Host,
		"loopbackRedirect": isLoopbackURL(redirectURL),
	})
}

type resolveMCPAuthorizationRequest struct {
	Request  string `json:"request"`
	Approved bool   `json:"approved"`
}

func (s *Server) resolveMCPAuthorization(w http.ResponseWriter, r *http.Request) {
	var request resolveMCPAuthorizationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	authorization, err := s.store.ConsumeMCPAuthorization(mcpAuthorizationRequestPrefix+request.Request, s.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "authorization request is invalid or expired, start connecting again")
		return
	}
	// Denying redirects too, so the redirect_uri is checked against the client metadata either way.
	client, redirectURL, ok := s.verifiedMCPAuthorization(w, r, authorization)
	if !ok {
		return
	}
	query := redirectURL.Query()
	if request.Approved {
		code, err := security.RandomToken(32)
		if err != nil {
			internalError(w, err, "generate MCP authorization code")
			return
		}
		authorization.ClientName = client.ClientName
		authorization.UserEmail = userFromRequest(r).Email
		authorization.ExpiresAt = s.now().UTC().Add(mcpAuthorizationCodeTTL).Truncate(time.Second)
		if err := s.store.PutMCPAuthorization(mcpAuthorizationCodePrefix+security.TokenDigest(code), authorization); err != nil {
			internalError(w, err, "store MCP authorization code")
			return
		}
		if err := s.recordAuditLog(r, "connected", "ai_agent", client.ClientID, map[string]string{"clientName": client.ClientName}); err != nil {
			internalError(w, err, "record AI agent connection audit log")
			return
		}
		query.Set("code", code)
	} else {
		query.Set("error", "access_denied")
	}
	if authorization.State != "" {
		query.Set("state", authorization.State)
	}
	query.Set("iss", s.mcpURL())
	redirectURL.RawQuery = query.Encode()
	writeJSON(w, http.StatusOK, map[string]string{"redirectUrl": redirectURL.String()})
}

// mcpToken answers OAuth errors with the RFC 6749 error code as the error field.
func (s *Server) mcpToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if resource := r.PostForm.Get("resource"); resource != "" && resource != s.mcpURL() {
		writeError(w, http.StatusBadRequest, "invalid_target")
		return
	}
	var tokens mcpTokens
	var err error
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		tokens, err = s.redeemMCPAuthorizationCode(r.PostForm)
	case "refresh_token":
		tokens, err = s.refreshMCPTokens(r.PostForm)
	default:
		writeError(w, http.StatusBadRequest, "unsupported_grant_type")
		return
	}
	if errors.Is(err, errInvalidGrant) {
		writeError(w, http.StatusBadRequest, errInvalidGrant.Error())
		return
	}
	if err != nil {
		internalError(w, err, "issue MCP tokens")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  tokens.access,
		"token_type":    "Bearer",
		"expires_in":    int(mcpAccessTokenTTL.Seconds()),
		"refresh_token": tokens.refresh,
	})
}

func (s *Server) redeemMCPAuthorizationCode(form url.Values) (mcpTokens, error) {
	authorization, err := s.store.ConsumeMCPAuthorization(mcpAuthorizationCodePrefix+security.TokenDigest(form.Get("code")), s.now())
	if errors.Is(err, store.ErrNotFound) {
		return mcpTokens{}, errInvalidGrant
	}
	if err != nil {
		return mcpTokens{}, err
	}
	challenge := sha256.Sum256([]byte(form.Get("code_verifier")))
	if authorization.ClientID != form.Get("client_id") || authorization.RedirectURI != form.Get("redirect_uri") ||
		base64.RawURLEncoding.EncodeToString(challenge[:]) != authorization.CodeChallenge {
		return mcpTokens{}, errInvalidGrant
	}
	id, err := security.RandomToken(12)
	if err != nil {
		return mcpTokens{}, err
	}
	connection := model.MCPConnection{
		ID:         id,
		UserEmail:  authorization.UserEmail,
		ClientID:   authorization.ClientID,
		ClientName: authorization.ClientName,
		CreatedAt:  s.now().UTC().Truncate(time.Second),
	}
	tokens, err := s.issueMCPTokens(&connection)
	if err != nil {
		return mcpTokens{}, err
	}
	return tokens, s.store.PutMCPConnection(connection)
}

// refreshMCPTokens rotates both tokens, so a refresh token works once.
func (s *Server) refreshMCPTokens(form url.Values) (mcpTokens, error) {
	refreshToken := form.Get("refresh_token")
	id, _, _ := strings.Cut(refreshToken, ".")
	var tokens mcpTokens
	_, err := s.store.ModifyMCPConnection(id, func(connection *model.MCPConnection) error {
		if connection.ClientID != form.Get("client_id") || connection.RefreshTokenDigest != security.TokenDigest(refreshToken) ||
			!connection.RefreshTokenExpiresAt.After(s.now()) {
			return errInvalidGrant
		}
		var err error
		tokens, err = s.issueMCPTokens(connection)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return mcpTokens{}, errInvalidGrant
	}
	return tokens, err
}

type mcpTokens struct {
	access  string
	refresh string
}

// issueMCPTokens prefixes tokens with the connection ID, which locates the connection that stores their digests.
func (s *Server) issueMCPTokens(connection *model.MCPConnection) (mcpTokens, error) {
	access, err := security.RandomToken(32)
	if err != nil {
		return mcpTokens{}, err
	}
	refresh, err := security.RandomToken(32)
	if err != nil {
		return mcpTokens{}, err
	}
	tokens := mcpTokens{access: connection.ID + "." + access, refresh: connection.ID + "." + refresh}
	now := s.now().UTC().Truncate(time.Second)
	connection.AccessTokenDigest = security.TokenDigest(tokens.access)
	connection.AccessTokenExpiresAt = now.Add(mcpAccessTokenTTL)
	connection.RefreshTokenDigest = security.TokenDigest(tokens.refresh)
	connection.RefreshTokenExpiresAt = now.Add(mcpRefreshTokenTTL)
	return tokens, nil
}
