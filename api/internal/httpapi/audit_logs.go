package httpapi

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
)

type auditChange struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

func (s *Server) listAuditLogs(w http.ResponseWriter, _ *http.Request) {
	entries, err := s.store.ListAuditLogs()
	if err != nil {
		internalError(w, err, "list audit logs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"auditLogs": entries})
}

func (s *Server) recordAuditLog(r *http.Request, action, resource, resourceID string, payload any) error {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	id, err := security.RandomToken(18)
	if err != nil {
		return err
	}
	actor := userFromRequest(r)
	return s.store.AppendAuditLog(model.AuditLog{
		ID: id,
		Actor: model.AuditLogActor{
			Email:      actor.Email,
			FullName:   actor.FullName,
			AvatarPath: actor.AvatarPath,
		},
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		CreatedAt:  s.now().UTC().Truncate(time.Second),
		Payload:    encodedPayload,
	}, s.cfg.AuditLog.MaxItems)
}

func auditDiff(before, after any, excluded ...string) (map[string]auditChange, error) {
	beforeFields, err := auditFields(before)
	if err != nil {
		return nil, err
	}
	afterFields, err := auditFields(after)
	if err != nil {
		return nil, err
	}
	for _, field := range excluded {
		delete(beforeFields, field)
		delete(afterFields, field)
	}
	changes := make(map[string]auditChange)
	for field, afterValue := range afterFields {
		beforeValue := beforeFields[field]
		if !reflect.DeepEqual(beforeValue, afterValue) {
			changes[field] = auditChange{Before: beforeValue, After: afterValue}
		}
	}
	return changes, nil
}

func auditFields(value any) (map[string]any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}
