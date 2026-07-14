package model

import (
	"encoding/json"
	"time"
)

type AuditLogActor struct {
	Email      string `json:"email"`
	FullName   string `json:"fullName"`
	AvatarPath string `json:"avatarPath,omitempty"`
}

type AuditLog struct {
	ID         string          `json:"id"`
	Actor      AuditLogActor   `json:"actor"`
	Action     string          `json:"action"`
	Resource   string          `json:"resource"`
	ResourceID string          `json:"resourceId"`
	CreatedAt  time.Time       `json:"createdAt"`
	Payload    json.RawMessage `json:"payload"`
}
