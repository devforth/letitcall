package model

import "time"

type User struct {
	Email                string    `json:"email"`
	PasswordHash         string    `json:"passwordHash"`
	Timezone             string    `json:"timezone"`
	AvatarPath           string    `json:"avatarPath,omitempty"`
	GoogleConnected      bool      `json:"googleConnected"`
	EncryptedGoogleToken string    `json:"encryptedGoogleToken,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type PublicUser struct {
	Email           string    `json:"email"`
	Timezone        string    `json:"timezone"`
	AvatarPath      string    `json:"avatarPath,omitempty"`
	GoogleConnected bool      `json:"googleConnected"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		Email:           u.Email,
		Timezone:        u.Timezone,
		AvatarPath:      u.AvatarPath,
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
	ID              string    `json:"id"`
	EventSlug       string    `json:"eventSlug"`
	Time            time.Time `json:"time"`
	EndTime         time.Time `json:"endTime"`
	AttendeeEmail   string    `json:"attendeeEmail"`
	Title           string    `json:"title"`
	RecipientEmails []string  `json:"recipientEmails"`
	CreatedAt       time.Time `json:"createdAt"`
}

type EventType struct {
	EventSlug         string        `json:"eventSlug"`
	Name              string        `json:"name"`
	DurationMinutes   int           `json:"durationMinutes"`
	BookingWindowDays *int          `json:"bookingWindowDays"`
	InviteeLimit      *int          `json:"inviteeLimit"`
	Timezone          string        `json:"timezone"`
	RecipientEmails   []string      `json:"recipientEmails"`
	Schedule          []ScheduleDay `json:"schedule"`
	CreatedBy         string        `json:"createdBy"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type ScheduleDay struct {
	Day     string      `json:"day"`
	Enabled bool        `json:"enabled"`
	Start   string      `json:"start,omitempty"`
	End     string      `json:"end,omitempty"`
	Breaks  []TimeRange `json:"breaks"`
}

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
