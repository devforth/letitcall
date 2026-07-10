package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/content"
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
	FullName string `json:"fullName"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
	Avatar   string `json:"avatar"`
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
	user.FullName = strings.TrimSpace(request.FullName)
	var avatar content.Avatar
	if request.Avatar != "" {
		avatar, err = s.avatars.Prepare(user.Email, request.Avatar)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		user.AvatarPath = avatar.Filename
	}
	if err := s.store.CreateUser(user); errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "a user with this email already exists")
		return
	} else if err != nil {
		internalError(w, err, "create user")
		return
	}
	if avatar.Filename != "" {
		if err := s.avatars.Write(avatar); err != nil {
			_ = s.store.DeleteUser(user.Email)
			internalError(w, err, "store user avatar")
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user.Public()})
}

type updateUserRequest struct {
	FullName *string `json:"fullName"`
	Password *string `json:"password"`
	Timezone *string `json:"timezone"`
	Avatar   *string `json:"avatar"`
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
	if request.FullName == nil && request.Password == nil && request.Timezone == nil && request.Avatar == nil {
		writeError(w, http.StatusBadRequest, "fullName, password, timezone, or avatar is required")
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
	previousAvatarFilename := user.AvatarPath
	if request.FullName != nil {
		user.FullName = strings.TrimSpace(*request.FullName)
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
	var avatar content.Avatar
	if request.Avatar != nil {
		avatar, err = s.avatars.Prepare(user.Email, *request.Avatar)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		user.AvatarPath = avatar.Filename
	}
	user.UpdatedAt = s.now().UTC().Truncate(time.Second)
	if avatar.Filename != "" {
		if err := s.avatars.Write(avatar); err != nil {
			internalError(w, err, "store user avatar")
			return
		}
	}
	if err := s.store.PutUser(user); err != nil {
		internalError(w, err, "update user")
		return
	}
	if avatar.Filename != "" && previousAvatarFilename != "" {
		if err := s.avatars.Remove(previousAvatarFilename); err != nil {
			slog.Error("remove previous user avatar", "error", err, "filename", previousAvatarFilename)
		}
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
	if err := s.store.RemoveEventTypeRecipient(email, s.now().UTC().Truncate(time.Second)); errors.Is(err, store.ErrLastRecipient) {
		writeError(w, http.StatusConflict, "user is the only recipient for an event type")
		return
	} else if err != nil {
		internalError(w, err, "remove user from event types")
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
