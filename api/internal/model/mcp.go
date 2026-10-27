package model

import "time"

// MCPAuthorization is a pending OAuth authorization request, or the authorization code issued once it is approved.
type MCPAuthorization struct {
	ClientID      string    `json:"clientId"`
	ClientName    string    `json:"clientName,omitempty"`
	RedirectURI   string    `json:"redirectUri"`
	CodeChallenge string    `json:"codeChallenge"`
	State         string    `json:"state,omitempty"`
	UserEmail     string    `json:"userEmail,omitempty"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// MCPConnection is an OAuth grant of an AI agent acting as a user through MCP.
type MCPConnection struct {
	ID                    string    `json:"id"`
	UserEmail             string    `json:"userEmail"`
	ClientID              string    `json:"clientId"`
	ClientName            string    `json:"clientName"`
	AccessTokenDigest     string    `json:"accessTokenDigest"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshTokenDigest    string    `json:"refreshTokenDigest"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
	CreatedAt             time.Time `json:"createdAt"`
}
