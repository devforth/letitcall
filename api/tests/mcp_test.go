package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mcpClientID     = "https://agent.example/client-metadata.json"
	chatGPTClientID = "https://chatgpt.example/oauth/client.json"
	mcpRedirectURI  = "http://127.0.0.1:5555/callback"
	mcpVerifier     = "test-code-verifier-with-enough-entropy-0123456789"
)

func TestMCPMetadataIsServedInsideAndOutsideBasePath(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/letitcall")
	mcpURL := "http://example.test/letitcall/api/mcp"
	for _, path := range []string{"/letitcall/api/mcp/oauth-protected-resource", "/.well-known/oauth-protected-resource/letitcall/api/mcp"} {
		body := expectStatus(t, mustDo(t, f.client, mustRequest(t, http.MethodGet, f.server.URL+path, nil)), http.StatusOK)
		if !strings.Contains(string(body), `"resource":"`+mcpURL+`"`) || !strings.Contains(string(body), `"authorization_servers":["`+mcpURL+`"]`) {
			t.Fatalf("unexpected protected resource metadata at %s: %s", path, body)
		}
	}
	for _, path := range []string{"/letitcall/api/mcp/.well-known/openid-configuration", "/.well-known/oauth-authorization-server/letitcall/api/mcp"} {
		body := expectStatus(t, mustDo(t, f.client, mustRequest(t, http.MethodGet, f.server.URL+path, nil)), http.StatusOK)
		if !strings.Contains(string(body), `"issuer":"`+mcpURL+`"`) || !strings.Contains(string(body), `"token_endpoint":"`+mcpURL+`/oauth/token"`) ||
			!strings.Contains(string(body), `"client_id_metadata_document_supported":true`) {
			t.Fatalf("unexpected authorization server metadata at %s: %s", path, body)
		}
	}
	response := mustDo(t, f.client, mustRequest(t, http.MethodPost, f.server.URL+"/letitcall/api/mcp", strings.NewReader(`{}`)))
	if challenge := response.Header.Get("WWW-Authenticate"); challenge != `Bearer resource_metadata="`+mcpURL+`/oauth-protected-resource"` {
		t.Fatalf("unexpected MCP challenge: %q", challenge)
	}
	if body := expectStatus(t, response, http.StatusUnauthorized); string(body) != "{\"error\":\"authentication required\"}\n" {
		t.Fatalf("unexpected MCP authentication error: %s", body)
	}
}

func TestMCPOAuthConnectsAgentThatUsesTools(t *testing.T) {
	f := newFixture(t, false)
	mockMCPClientMetadata(t)
	expectStatus(t, f.request(http.MethodGet, "/api/mcp/oauth/authorize?"+authorizeQuery(mcpClientID, "http://evil.example/callback").Encode(), nil), http.StatusBadRequest)

	requestID := f.startMCPAuthorization(authorizeQuery(mcpClientID, mcpRedirectURI))
	expectStatus(t, f.request(http.MethodGet, "/api/mcp/oauth/authorization?request="+requestID, nil), http.StatusUnauthorized)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	described := expectStatus(t, f.request(http.MethodGet, "/api/mcp/oauth/authorization?request="+requestID, nil), http.StatusOK)
	if string(described) != `{"clientHost":"agent.example","clientName":"Test Agent","loopbackRedirect":true,"redirectHost":"127.0.0.1:5555"}`+"\n" {
		t.Fatalf("unexpected authorization description: %s", described)
	}
	code := f.approveMCPAuthorization(requestID)
	tokens := f.exchangeMCPToken(url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {mcpVerifier},
		"client_id": {mcpClientID}, "redirect_uri": {mcpRedirectURI}, "resource": {"http://example.test/api/mcp"},
	}, http.StatusOK)
	f.exchangeMCPToken(url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {mcpVerifier},
		"client_id": {mcpClientID}, "redirect_uri": {mcpRedirectURI},
	}, http.StatusBadRequest)

	session := f.connectMCP(tokens.AccessToken)
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 9 {
		t.Fatalf("unexpected MCP tools: %d", len(listed.Tools))
	}
	start := time.Now().UTC().AddDate(0, 0, 1).Truncate(time.Hour).Format(time.RFC3339)
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "create_busy_time", Arguments: map[string]any{
		"start": start, "durationMinutes": 45, "comment": "Focus time",
	}})
	if err != nil || result.IsError {
		t.Fatalf("create_busy_time failed: %v %+v", err, result)
	}
	busyTimes := expectStatus(t, f.request(http.MethodGet, "/api/busy-times", nil), http.StatusOK)
	if !strings.Contains(string(busyTimes), `"userEmail":"`+adminEmail+`","start":"`+start+`","durationMinutes":45,"comment":"Focus time"`) {
		t.Fatalf("MCP busy time was not stored: %s", busyTimes)
	}
	eventType := eventTypeBody("Agent Call", []string{adminEmail}, 1)
	eventType["oneTime"] = true
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{Name: "create_event_type", Arguments: eventType})
	if err != nil || result.IsError {
		t.Fatalf("create_event_type failed: %v %+v", err, result)
	}
	created := expectStatus(t, f.request(http.MethodGet, "/api/event-types/agent-call", nil), http.StatusOK)
	if !strings.Contains(string(created), `"oneTime":true`) || !strings.Contains(string(created), `"createdBy":"`+adminEmail+`"`) {
		t.Fatalf("MCP event type was not created as the user: %s", created)
	}
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{Name: "delete_busy_time", Arguments: map[string]any{"id": "missing"}})
	if err != nil || !result.IsError || result.Content[0].(*mcp.TextContent).Text != "busy time not found" {
		t.Fatalf("API errors must be tool errors: %v %+v", err, result)
	}

	refreshed := f.exchangeMCPToken(url.Values{"grant_type": {"refresh_token"}, "refresh_token": {tokens.RefreshToken}, "client_id": {mcpClientID}}, http.StatusOK)
	f.exchangeMCPToken(url.Values{"grant_type": {"refresh_token"}, "refresh_token": {tokens.RefreshToken}, "client_id": {mcpClientID}}, http.StatusBadRequest)
	expectStatus(t, f.mcpRequest(tokens.AccessToken), http.StatusUnauthorized)
	expectStatus(t, f.mcpRequest(refreshed.AccessToken), http.StatusOK)

	connections := expectStatus(t, f.request(http.MethodGet, "/api/mcp/connections", nil), http.StatusOK)
	var listing struct {
		MCPURL      string `json:"mcpURL"`
		Connections []struct {
			ID         string `json:"id"`
			ClientName string `json:"clientName"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(connections, &listing); err != nil {
		t.Fatal(err)
	}
	if listing.MCPURL != "http://example.test/api/mcp" || len(listing.Connections) != 1 || listing.Connections[0].ClientName != "Test Agent" {
		t.Fatalf("unexpected AI connections: %s", connections)
	}
	expectStatus(t, f.request(http.MethodDelete, "/api/mcp/connections/"+listing.Connections[0].ID, nil), http.StatusNoContent)
	expectStatus(t, f.mcpRequest(refreshed.AccessToken), http.StatusUnauthorized)
	auditLogs := expectStatus(t, f.request(http.MethodGet, "/api/audit-logs", nil), http.StatusOK)
	if !strings.Contains(string(auditLogs), `"action":"connected","resource":"ai_agent"`) || !strings.Contains(string(auditLogs), `"action":"disconnected","resource":"ai_agent"`) {
		t.Fatalf("AI agent connections were not audited: %s", auditLogs)
	}
}

func TestMCPAuthorizationCanBeDenied(t *testing.T) {
	f := newFixture(t, false)
	mockMCPClientMetadata(t)
	requestID := f.startMCPAuthorization(authorizeQuery(mcpClientID, mcpRedirectURI))
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	resolved := expectStatus(t, f.request(http.MethodPost, "/api/mcp/oauth/authorization", map[string]any{"request": requestID, "approved": false}), http.StatusOK)
	if !strings.Contains(string(resolved), "error=access_denied") || !strings.Contains(string(resolved), "state=agent-state") {
		t.Fatalf("denial did not redirect with access_denied: %s", resolved)
	}
	expectStatus(t, f.request(http.MethodPost, "/api/mcp/oauth/authorization", map[string]any{"request": requestID, "approved": true}), http.StatusBadRequest)

	unregistered := f.startMCPAuthorization(authorizeQuery(mcpClientID, "http://127.0.0.1:5555/other"))
	described := expectStatus(t, f.request(http.MethodGet, "/api/mcp/oauth/authorization?request="+unregistered, nil), http.StatusBadRequest)
	if !strings.Contains(string(described), "redirect_uri is not listed in the client metadata document") {
		t.Fatalf("unexpected unregistered redirect error: %s", described)
	}
}

func TestMCPAcceptsClientsSupportingNoneAmongAuthMethods(t *testing.T) {
	f := newFixture(t, false)
	mockMCPClientMetadata(t)
	requestID := f.startMCPAuthorization(authorizeQuery(chatGPTClientID, "https://chatgpt.example/connector_platform_oauth_redirect"))
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	described := expectStatus(t, f.request(http.MethodGet, "/api/mcp/oauth/authorization?request="+requestID, nil), http.StatusOK)
	if !strings.Contains(string(described), `"clientName":"ChatGPT"`) {
		t.Fatalf("a client listing none among its methods was rejected: %s", described)
	}
}

type mcpTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func authorizeQuery(clientID, redirectURI string) url.Values {
	challenge := sha256.Sum256([]byte(mcpVerifier))
	return url.Values{
		"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {redirectURI}, "state": {"agent-state"},
		"code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"},
	}
}

func (f *fixture) startMCPAuthorization(query url.Values) string {
	f.t.Helper()
	response := f.request(http.MethodGet, "/api/mcp/oauth/authorize?"+query.Encode(), nil)
	expectStatus(f.t, response, http.StatusFound)
	location, err := url.Parse(response.Header.Get("Location"))
	if err != nil || location.Path != "/auth/connect-ai" || location.Query().Get("request") == "" {
		f.t.Fatalf("authorize did not redirect to the consent page: %q", response.Header.Get("Location"))
	}
	return location.Query().Get("request")
}

func (f *fixture) approveMCPAuthorization(requestID string) string {
	f.t.Helper()
	body := expectStatus(f.t, f.request(http.MethodPost, "/api/mcp/oauth/authorization", map[string]any{"request": requestID, "approved": true}), http.StatusOK)
	var resolved struct {
		RedirectURL string `json:"redirectUrl"`
	}
	if err := json.Unmarshal(body, &resolved); err != nil {
		f.t.Fatal(err)
	}
	redirect, err := url.Parse(resolved.RedirectURL)
	if err != nil || !strings.HasPrefix(resolved.RedirectURL, mcpRedirectURI+"?") || redirect.Query().Get("state") != "agent-state" ||
		redirect.Query().Get("iss") != "http://example.test/api/mcp" || redirect.Query().Get("code") == "" {
		f.t.Fatalf("unexpected approval redirect: %s", body)
	}
	return redirect.Query().Get("code")
}

func (f *fixture) exchangeMCPToken(form url.Values, status int) mcpTokenResponse {
	f.t.Helper()
	request := mustRequest(f.t, http.MethodPost, f.server.URL+"/api/mcp/oauth/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	body := expectStatus(f.t, mustDo(f.t, f.client, request), status)
	var tokens mcpTokenResponse
	if status != http.StatusOK {
		if string(body) != "{\"error\":\"invalid_grant\"}\n" {
			f.t.Fatalf("unexpected token error: %s", body)
		}
		return tokens
	}
	if err := json.Unmarshal(body, &tokens); err != nil {
		f.t.Fatal(err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" || tokens.TokenType != "Bearer" || tokens.ExpiresIn != 3600 {
		f.t.Fatalf("unexpected token response: %s", body)
	}
	return tokens
}

func (f *fixture) mcpRequest(accessToken string) *http.Response {
	f.t.Helper()
	request := mustRequest(f.t, http.MethodPost, f.server.URL+"/api/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	return mustDo(f.t, f.server.Client(), request)
}

func (f *fixture) connectMCP(accessToken string) *mcp.ClientSession {
	f.t.Helper()
	transport := f.server.Client().Transport
	client := mcp.NewClient(&mcp.Implementation{Name: "test-agent", Version: "1.0.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: f.server.URL + "/api/mcp",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			request.Header.Set("Authorization", "Bearer "+accessToken)
			return transport.RoundTrip(request)
		})},
	}, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { session.Close() })
	return session
}

func mustRequest(t *testing.T, method, target string, body io.Reader) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, target, body)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func mockMCPClientMetadata(t *testing.T) {
	t.Helper()
	originalTransport := http.DefaultTransport
	documents := map[string]map[string]any{
		mcpClientID: {
			"client_id": mcpClientID, "client_name": "Test Agent", "redirect_uris": []string{"http://127.0.0.1/callback"},
			"token_endpoint_auth_method": "none",
		},
		// Shaped as the ChatGPT document: it prefers private_key_jwt but also supports none.
		chatGPTClientID: {
			"client_id": chatGPTClientID, "client_name": "ChatGPT", "redirect_uris": []string{"https://chatgpt.example/connector_platform_oauth_redirect"},
			"token_endpoint_auth_method": "private_key_jwt", "token_endpoint_auth_methods_supported": []string{"none", "private_key_jwt"},
		},
	}
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		value, ok := documents[request.URL.String()]
		if !ok {
			return nil, fmt.Errorf("unexpected client metadata request: %s", request.URL)
		}
		document, _ := json.Marshal(value)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(document)),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })
}
