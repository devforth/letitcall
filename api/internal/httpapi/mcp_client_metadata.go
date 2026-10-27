package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"syscall"
	"time"
)

const (
	clientMetadataMaxBytes  = 5 << 10
	clientMetadataTimeout   = 5 * time.Second
	clientNameMaxCharacters = 100
)

var loopbackHosts = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}

// mcpClient is an MCP client resolved from its Client ID Metadata Document.
type mcpClient struct {
	ClientID     string
	ClientName   string
	RedirectURIs []string
}

func isClientIDURL(clientID string) bool {
	parsed, err := url.Parse(clientID)
	return err == nil && parsed.Scheme == "https" && parsed.String() == clientID && parsed.User == nil &&
		parsed.Path != "" && parsed.Path != "/" && parsed.Fragment == "" && net.ParseIP(parsed.Hostname()) == nil
}

func isLoopbackURL(value *url.URL) bool {
	return value.Scheme == "http" && loopbackHosts[value.Hostname()]
}

func isAllowedRedirectURI(value *url.URL) bool {
	return value.Scheme == "https" || isLoopbackURL(value)
}

// redirectURIMatches lets a native client listening on a random loopback port differ from the registered port (RFC 8252).
func redirectURIMatches(requested *url.URL, registered string) bool {
	registeredURL, err := url.Parse(registered)
	if err != nil {
		return false
	}
	if !isLoopbackURL(registeredURL) {
		return requested.String() == registeredURL.String()
	}
	return requested.Scheme == registeredURL.Scheme && requested.Hostname() == registeredURL.Hostname() &&
		requested.Path == registeredURL.Path && requested.RawQuery == registeredURL.RawQuery && requested.Fragment == registeredURL.Fragment
}

// fetchMCPClient resolves client_id, a URL chosen by whoever sent the authorization link, so the fetch refuses
// private, loopback and other non-public addresses.
func fetchMCPClient(ctx context.Context, clientID string) (mcpClient, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, clientID, nil)
	if err != nil {
		return mcpClient{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := publicHTTPClient().Do(request)
	if err != nil {
		slog.Warn("fetch MCP client metadata document", "error", err, "clientId", clientID)
		return mcpClient{}, errors.New("could not fetch the client metadata document")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, clientMetadataMaxBytes+1))
	if err != nil || response.StatusCode != http.StatusOK || len(body) > clientMetadataMaxBytes {
		slog.Warn("fetch MCP client metadata document", "error", err, "status", response.Status, "bytes", len(body), "clientId", clientID)
		return mcpClient{}, errors.New("could not fetch the client metadata document")
	}
	var document struct {
		ClientID                string   `json:"client_id"`
		ClientName              string   `json:"client_name"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}
	if err := json.Unmarshal(body, &document); err != nil || document.ClientName == "" || len(document.RedirectURIs) == 0 || document.TokenEndpointAuthMethod != "none" {
		return mcpClient{}, errors.New(`client metadata document must have client_id, client_name, redirect_uris and token_endpoint_auth_method "none"`)
	}
	if document.ClientID != clientID {
		return mcpClient{}, errors.New("client_id in the client metadata document does not match its URL")
	}
	name := []rune(document.ClientName)
	return mcpClient{ClientID: clientID, ClientName: string(name[:min(len(name), clientNameMaxCharacters)]), RedirectURIs: document.RedirectURIs}, nil
}

// publicHTTPClient checks the address it actually dials, which closes DNS rebinding.
func publicHTTPClient() *http.Client {
	client := &http.Client{
		Timeout:       clientMetadataTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     http.DefaultTransport,
	}
	// Tests replace http.DefaultTransport with a stub that never dials.
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		public := transport.Clone()
		public.Proxy = nil
		public.DialContext = (&net.Dialer{Timeout: clientMetadataTimeout, Control: dialPublicAddress}).DialContext
		client.Transport = public
	}
	return client
}

func dialPublicAddress(_, address string, _ syscall.RawConn) error {
	addressPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return err
	}
	ip := addressPort.Addr().Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return fmt.Errorf("%s is not a public address", ip)
	}
	return nil
}
