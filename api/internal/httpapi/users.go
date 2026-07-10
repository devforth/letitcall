package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

func (s *Server) listUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := s.store.ListUsers()
	if err != nil {
		internalError(w, err, "list users")
		return
	}
	publicUsers := make([]model.PublicUser, 0, len(users))
	for _, user := range users {
		publicUsers = append(publicUsers, user.Public())
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": publicUsers})
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	user, err := security.NewUser(request.Email, request.Password, request.Timezone, s.now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.CreateUser(user); errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "a user with this email already exists")
		return
	} else if err != nil {
		internalError(w, err, "create user")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user.Public()})
}

type updateUserRequest struct {
	Password *string `json:"password"`
	Timezone *string `json:"timezone"`
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	email, err := security.NormalizeEmail(r.PathValue("email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var request updateUserRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	if request.Password == nil && request.Timezone == nil {
		writeError(w, http.StatusBadRequest, "password or timezone is required")
		return
	}
	user, err := s.store.GetUser(email)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		internalError(w, err, "load user for update")
		return
	}
	if request.Password != nil {
		hash, hashErr := security.HashPassword(*request.Password)
		if hashErr != nil {
			writeError(w, http.StatusBadRequest, hashErr.Error())
			return
		}
		user.PasswordHash = hash
	}
	if request.Timezone != nil {
		timezone, timezoneErr := security.ValidateTimezone(*request.Timezone)
		if timezoneErr != nil {
			writeError(w, http.StatusBadRequest, timezoneErr.Error())
			return
		}
		user.Timezone = timezone
	}
	user.UpdatedAt = s.now().UTC().Truncate(time.Second)
	if err := s.store.PutUser(user); err != nil {
		internalError(w, err, "update user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user.Public()})
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	email, err := security.NormalizeEmail(r.PathValue("email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.EqualFold(email, userFromRequest(r).Email) {
		writeError(w, http.StatusConflict, "you cannot delete your own user")
		return
	}
	if err := s.store.DeleteUserIfMoreThanOne(email); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err := s.store.DeleteSessionsForUser(email); err != nil {
		internalError(w, err, "delete user sessions")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
