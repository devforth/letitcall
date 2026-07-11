package model

import "time"

type User struct {
	Email                string    `json:"email"`
	FullName             string    `json:"fullName"`
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
	FullName        string    `json:"fullName"`
	Timezone        string    `json:"timezone"`
	AvatarPath      string    `json:"avatarPath,omitempty"`
	GoogleConnected bool      `json:"googleConnected"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		Email:           u.Email,
		FullName:        u.FullName,
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
	ID                 string            `json:"id"`
	EventSlug          string            `json:"eventSlug"`
	Time               time.Time         `json:"time"`
	EndTime            time.Time         `json:"endTime"`
	AttendeeName       string            `json:"attendeeName"`
	AttendeeEmail      string            `json:"attendeeEmail"`
	AttendeeTimezone   string            `json:"attendeeTimezone"`
	GuestEmails        []string          `json:"guestEmails"`
	Notes              string            `json:"notes,omitempty"`
	Title              string            `json:"title"`
	RecipientEmails    []string          `json:"recipientEmails"`
	GoogleEventIDs     map[string]string `json:"googleEventIds,omitempty"`
	CanceledAt         *time.Time        `json:"canceledAt,omitempty"`
	CanceledBy         *BookingActor     `json:"canceledBy,omitempty"`
	CancellationReason string            `json:"cancellationReason,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

type BookingActor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type EventType struct {
	EventSlug          string        `json:"eventSlug"`
	Name               string        `json:"name"`
	DurationMinutes    int           `json:"durationMinutes"`
	BookingWindowDays  int           `json:"bookingWindowDays"`
	InviteeLimit       *int          `json:"inviteeLimit"`
	Timezone           string        `json:"timezone"`
	RequiredHostEmails []string      `json:"requiredHostEmails"`
	OptionalHostEmails []string      `json:"optionalHostEmails"`
	Schedule           []ScheduleDay `json:"schedule"`
	CreatedBy          string        `json:"createdBy"`
	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}

func (e EventType) HostEmails() []string {
	return append(append([]string(nil), e.RequiredHostEmails...), e.OptionalHostEmails...)
}

type GoogleBusyPeriod struct {
	EventID string    `json:"eventId"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

type GoogleBusyCache struct {
	Periods  []GoogleBusyPeriod `json:"periods"`
	SyncedAt time.Time          `json:"syncedAt"`
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
