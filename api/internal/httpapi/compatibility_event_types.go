package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

const bookingNotesQuestion = "Please share anything that will help prepare for our meeting."

type compatibilityUser struct {
	URI                 string    `json:"uri"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	Timezone            string    `json:"timezone"`
	AvatarURL           *string   `json:"avatar_url"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	CurrentOrganization string    `json:"current_organization"`
	// TODO: Add slug and scheduling_url when users have public scheduling pages.
}

type compatibilityEventTypeProfile struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type compatibilityCustomQuestion struct {
	UUID          *string  `json:"uuid"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Position      int      `json:"position"`
	Enabled       bool     `json:"enabled"`
	Required      bool     `json:"required"`
	AnswerChoices []string `json:"answer_choices"`
	IncludeOther  bool     `json:"include_other"`
}

type compatibilityEventType struct {
	URI             string                        `json:"uri"`
	Name            string                        `json:"name"`
	Active          bool                          `json:"active"`
	BookingMethod   string                        `json:"booking_method"`
	Slug            string                        `json:"slug"`
	SchedulingURL   string                        `json:"scheduling_url"`
	Duration        int                           `json:"duration"`
	Kind            string                        `json:"kind"`
	PoolingType     *string                       `json:"pooling_type"`
	Type            string                        `json:"type"`
	KindDescription string                        `json:"kind_description"`
	CreatedAt       time.Time                     `json:"created_at"`
	UpdatedAt       time.Time                     `json:"updated_at"`
	Profile         compatibilityEventTypeProfile `json:"profile"`
	CustomQuestions []compatibilityCustomQuestion `json:"custom_questions"`
	// TODO: Add color, descriptions, secret, deleted_at, and admin_managed when represented by the event type model.
}

func (s *Server) compatibilityCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := userFromRequest(r)
	var avatarURL *string
	if user.AvatarPath != "" {
		value := s.cfg.HTTP.BaseURL + "/content/avatars/" + user.AvatarPath
		avatarURL = &value
	}
	writeJSON(w, http.StatusOK, map[string]any{"resource": compatibilityUser{
		URI: s.userURI(user.Email), Name: user.FullName, Email: user.Email, Timezone: user.Timezone,
		AvatarURL: avatarURL, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
		CurrentOrganization: s.organizationURI(),
	}})
}

func (s *Server) compatibilityListEventTypes(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	userURI, organization := values.Get("user"), values.Get("organization")
	if userURI == "" && organization == "" {
		writeError(w, http.StatusBadRequest, "user or organization is required")
		return
	}
	if organization != "" && organization != s.organizationURI() {
		writeError(w, http.StatusNotFound, "organization not found")
		return
	}
	var userEmail string
	if userURI != "" {
		user, err := s.userForURI(userURI)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			internalError(w, err, "load compatibility event type user")
			return
		}
		userEmail = user.Email
	}
	if active := values.Get("active"); active != "" {
		parsed, err := strconv.ParseBool(active)
		if err != nil {
			writeError(w, http.StatusBadRequest, "active must be true or false")
			return
		}
		if !parsed {
			count, offset, err := paginationRequest(values)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			page, pagination := compatibilityPage(r, []compatibilityEventType{}, count, offset, s.compatibilityBaseURL())
			writeJSON(w, http.StatusOK, map[string]any{"collection": page, "pagination": pagination})
			return
		}
	}
	if adminManaged := values.Get("admin_managed"); adminManaged != "" {
		parsed, err := strconv.ParseBool(adminManaged)
		if err != nil {
			writeError(w, http.StatusBadRequest, "admin_managed must be true or false")
			return
		}
		if parsed {
			writeJSON(w, http.StatusOK, map[string]any{"collection": []compatibilityEventType{}, "pagination": compatibilityPagination{Count: 0}})
			return
		}
	}
	eventTypes, err := s.store.ListEventTypes()
	if err != nil {
		internalError(w, err, "list compatibility event types")
		return
	}
	items := make([]compatibilityEventType, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		if userEmail != "" && !containsEmail(eventType.HostEmails(), userEmail) {
			continue
		}
		items = append(items, s.compatibilityEventType(eventType))
	}
	sortValue := values.Get("sort")
	if sortValue == "" || sortValue == "name" || sortValue == "name:asc" {
		sort.SliceStable(items, func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) })
	} else if sortValue == "name:desc" {
		sort.SliceStable(items, func(i, j int) bool { return strings.ToLower(items[i].Name) > strings.ToLower(items[j].Name) })
	} else {
		writeError(w, http.StatusBadRequest, "sort must be name:asc or name:desc")
		return
	}
	count, offset, err := paginationRequest(values)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	page, pagination := compatibilityPage(r, items, count, offset, s.compatibilityBaseURL())
	writeJSON(w, http.StatusOK, map[string]any{"collection": page, "pagination": pagination})
}

func (s *Server) compatibilityGetEventType(w http.ResponseWriter, r *http.Request) {
	eventType, err := s.store.GetEventType(r.PathValue("uuid"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load compatibility event type")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"resource": s.compatibilityEventType(eventType)})
}

func (s *Server) compatibilityEventType(eventType model.EventType) compatibilityEventType {
	kind := "solo"
	kindDescription := "One-on-One"
	if eventType.InviteeLimit != nil {
		kind = "group"
		kindDescription = "Group"
	}
	var poolingType *string
	if len(eventType.RequiredHostEmails) > 1 {
		value := "collective"
		poolingType = &value
		if kind == "solo" {
			kindDescription = "Collective"
		}
	}
	profileName := eventType.CreatedBy
	if creator, err := s.store.GetUser(eventType.CreatedBy); err == nil && creator.FullName != "" {
		profileName = creator.FullName
	}
	return compatibilityEventType{
		URI: s.eventTypeURI(eventType.EventSlug), Name: eventType.Name, Active: true,
		BookingMethod: "instant", Slug: eventType.EventSlug,
		SchedulingURL: s.cfg.HTTP.BaseURL + "/book/" + eventType.EventSlug,
		Duration:      eventType.DurationMinutes, Kind: kind, PoolingType: poolingType,
		Type: "StandardEventType", KindDescription: kindDescription,
		CreatedAt: eventType.CreatedAt, UpdatedAt: eventType.UpdatedAt,
		Profile: compatibilityEventTypeProfile{Type: "User", Name: profileName, Owner: s.userURI(eventType.CreatedBy)},
		CustomQuestions: []compatibilityCustomQuestion{{
			Name: bookingNotesQuestion, Type: "text", Position: 0, Enabled: true,
			Required: false, AnswerChoices: []string{}, IncludeOther: false,
		}},
	}
}
