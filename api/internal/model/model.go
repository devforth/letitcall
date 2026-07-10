package model

import "time"

type User struct {
	Email                string    `json:"email"`
	PasswordHash         string    `json:"passwordHash"`
	Timezone             string    `json:"timezone"`
	GoogleConnected      bool      `json:"googleConnected"`
	EncryptedGoogleToken string    `json:"encryptedGoogleToken,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type PublicUser struct {
	Email           string    `json:"email"`
	Timezone        string    `json:"timezone"`
	GoogleConnected bool      `json:"googleConnected"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		Email:           u.Email,
		Timezone:        u.Timezone,
		GoogleConnected: u.GoogleConnected,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

type Session struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type OAuthState struct {
	CodeVerifier string    `json:"codeVerifier"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type Booking struct {
	Time          time.Time `json:"time"`
	OwnerEmail    string    `json:"ownerEmail"`
	AttendeeEmail string    `json:"attendeeEmail"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"createdAt"`
}
