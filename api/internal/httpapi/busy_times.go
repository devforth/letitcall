package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

type busyTimeRequest struct {
	UserEmail       string `json:"userEmail"`
	Start           string `json:"start"`
	DurationMinutes int    `json:"durationMinutes"`
	Comment         string `json:"comment"`
}

func (s *Server) listBusyTimes(w http.ResponseWriter, _ *http.Request) {
	busyTimes, err := s.store.ListBusyTimes()
	if err != nil {
		internalError(w, err, "list busy times")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"busyTimes": busyTimes})
}

func (s *Server) createBusyTime(w http.ResponseWriter, r *http.Request) {
	var request busyTimeRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	email, err := security.NormalizeEmail(request.UserEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, "userEmail must be a valid user email")
		return
	}
	if _, err := s.store.GetUser(email); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("user %q does not exist", email))
		return
	} else if err != nil {
		internalError(w, err, "load busy time user")
		return
	}
	start, err := parseCompatibilityInstant(request.Start, "start")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.DurationMinutes < 1 {
		writeError(w, http.StatusBadRequest, "durationMinutes must be a positive integer")
		return
	}
	comment := strings.TrimSpace(request.Comment)
	if len(comment) > 2000 {
		writeError(w, http.StatusBadRequest, "comment must not exceed 2000 characters")
		return
	}
	id, err := security.RandomToken(12)
	if err != nil {
		internalError(w, err, "generate busy time ID")
		return
	}
	busyTime := model.BusyTime{
		ID:              id,
		UserEmail:       email,
		Start:           start,
		DurationMinutes: request.DurationMinutes,
		Comment:         comment,
		CreatedAt:       s.now().UTC().Truncate(time.Second),
	}
	if err := s.store.CreateBusyTime(busyTime); err != nil {
		internalError(w, err, "create busy time")
		return
	}
	if err := s.recordAuditLog(r, "created", "busy_time", busyTime.ID, busyTime); err != nil {
		internalError(w, err, "record busy time creation audit log")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"busyTime": busyTime})
}

func (s *Server) deleteBusyTime(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	busyTime, err := s.store.GetBusyTime(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "busy time not found")
		return
	} else if err != nil {
		internalError(w, err, "load busy time for deletion")
		return
	}
	if err := s.store.DeleteBusyTime(id); err != nil {
		internalError(w, err, "delete busy time")
		return
	}
	if err := s.recordAuditLog(r, "deleted", "busy_time", id, busyTime); err != nil {
		internalError(w, err, "record busy time deletion audit log")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
