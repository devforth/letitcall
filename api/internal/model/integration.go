package model

import "time"

type APIToken struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserEmail string    `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
}

type WebhookSubscription struct {
	ID                  string    `json:"id"`
	CallbackURL         string    `json:"callbackUrl"`
	Events              []string  `json:"events"`
	Scope               string    `json:"scope"`
	UserEmail           string    `json:"userEmail,omitempty"`
	CreatorEmail        string    `json:"creatorEmail"`
	EncryptedSigningKey string    `json:"encryptedSigningKey,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type WebhookDelivery struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscriptionId"`
	Payload        string    `json:"payload"`
	Attempts       int       `json:"attempts"`
	NextAttemptAt  time.Time `json:"nextAttemptAt"`
	CreatedAt      time.Time `json:"createdAt"`
}
