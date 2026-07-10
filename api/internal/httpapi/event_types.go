package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

var weekdays = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

type eventTypeRequest struct {
	Name              string              `json:"name"`
	DurationMinutes   int                 `json:"durationMinutes"`
	BookingWindowDays *int                `json:"bookingWindowDays"`
	InviteeLimit      *int                `json:"inviteeLimit"`
	Timezone          string              `json:"timezone"`
	RecipientEmails   []string            `json:"recipientEmails"`
	Schedule          []model.ScheduleDay `json:"schedule"`
}

func (s *Server) listEventTypes(w http.ResponseWriter, _ *http.Request) {
	eventTypes, err := s.store.ListEventTypes()
	if err != nil {
		internalError(w, err, "list event types")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"eventTypes": eventTypes})
}

func (s *Server) getEventType(w http.ResponseWriter, r *http.Request) {
	eventType, err := s.store.GetEventType(r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load event type")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"eventType": eventType})
}

func (s *Server) createEventType(w http.ResponseWriter, r *http.Request) {
	var request eventTypeRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	eventType, err := s.eventTypeFromRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	eventType.EventSlug = slugifyEventName(eventType.Name)
	if eventType.EventSlug == "" {
		writeError(w, http.StatusBadRequest, "name must contain at least one letter or number")
		return
	}
	now := s.now().UTC().Truncate(time.Second)
	eventType.CreatedBy = userFromRequest(r).Email
	eventType.CreatedAt = now
	eventType.UpdatedAt = now
	if err := s.store.CreateEventType(eventType); errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, fmt.Sprintf("an event type with slug %q already exists", eventType.EventSlug))
		return
	} else if err != nil {
		internalError(w, err, "create event type")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"eventType": eventType})
}

func (s *Server) updateEventType(w http.ResponseWriter, r *http.Request) {
	existing, err := s.store.GetEventType(r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load event type for update")
		return
	}
	var request eventTypeRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	eventType, err := s.eventTypeFromRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	eventType.EventSlug = existing.EventSlug
	eventType.CreatedBy = existing.CreatedBy
	eventType.CreatedAt = existing.CreatedAt
	eventType.UpdatedAt = s.now().UTC().Truncate(time.Second)
	if err := s.store.PutEventType(eventType); err != nil {
		internalError(w, err, "update event type")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"eventType": eventType})
}

func (s *Server) getPublicEventType(w http.ResponseWriter, r *http.Request) {
	eventType, err := s.store.GetEventType(r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load public event type")
		return
	}
	type host struct {
		Email      string `json:"email"`
		AvatarPath string `json:"avatarPath,omitempty"`
	}
	hosts := make([]host, 0, len(eventType.RecipientEmails))
	for _, email := range eventType.RecipientEmails {
		user, err := s.store.GetUser(email)
		if err != nil {
			internalError(w, err, "load public event type host")
			return
		}
		hosts = append(hosts, host{Email: user.Email, AvatarPath: user.AvatarPath})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"eventType": map[string]any{
			"eventSlug":         eventType.EventSlug,
			"name":              eventType.Name,
			"durationMinutes":   eventType.DurationMinutes,
			"bookingWindowDays": eventType.BookingWindowDays,
			"inviteeLimit":      eventType.InviteeLimit,
			"timezone":          eventType.Timezone,
			"schedule":          eventType.Schedule,
			"hosts":             hosts,
		},
	})
}

func (s *Server) eventTypeFromRequest(request eventTypeRequest) (model.EventType, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" || len(name) > 200 {
		return model.EventType{}, errors.New("name must be between 1 and 200 characters")
	}
	if request.DurationMinutes < 1 || request.DurationMinutes > 1440 {
		return model.EventType{}, errors.New("durationMinutes must be between 1 and 1440")
	}
	if request.BookingWindowDays != nil && *request.BookingWindowDays < 1 {
		return model.EventType{}, errors.New("bookingWindowDays must be a positive integer or null")
	}
	if request.InviteeLimit != nil && *request.InviteeLimit < 1 {
		return model.EventType{}, errors.New("inviteeLimit must be a positive integer or null")
	}
	timezone := strings.TrimSpace(request.Timezone)
	if _, err := time.LoadLocation(timezone); err != nil {
		return model.EventType{}, errors.New("timezone must be a valid IANA timezone")
	}
	recipients, err := s.validateRecipients(request.RecipientEmails)
	if err != nil {
		return model.EventType{}, err
	}
	schedule, err := validateSchedule(request.Schedule)
	if err != nil {
		return model.EventType{}, err
	}
	return model.EventType{
		Name:              name,
		DurationMinutes:   request.DurationMinutes,
		BookingWindowDays: request.BookingWindowDays,
		InviteeLimit:      request.InviteeLimit,
		Timezone:          timezone,
		RecipientEmails:   recipients,
		Schedule:          schedule,
	}, nil
}

func (s *Server) validateRecipients(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, errors.New("at least one recipient is required")
	}
	recipients := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		email, err := security.NormalizeEmail(value)
		if err != nil {
			return nil, errors.New("recipientEmails must contain valid user emails")
		}
		if seen[email] {
			return nil, errors.New("recipientEmails must not contain duplicates")
		}
		if _, err := s.store.GetUser(email); errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("recipient user %q does not exist", email)
		} else if err != nil {
			return nil, err
		}
		seen[email] = true
		recipients = append(recipients, email)
	}
	sort.Strings(recipients)
	return recipients, nil
}

func validateSchedule(values []model.ScheduleDay) ([]model.ScheduleDay, error) {
	if len(values) != len(weekdays) {
		return nil, errors.New("schedule must contain all seven weekdays")
	}
	byDay := make(map[string]model.ScheduleDay, len(values))
	enabledDays := 0
	for _, value := range values {
		value.Day = strings.ToLower(strings.TrimSpace(value.Day))
		if !contains(weekdays, value.Day) || byDay[value.Day].Day != "" {
			return nil, errors.New("schedule must contain each weekday exactly once")
		}
		if !value.Enabled {
			if value.Start != "" || value.End != "" || len(value.Breaks) != 0 {
				return nil, fmt.Errorf("disabled %s must not contain hours", value.Day)
			}
			value.Breaks = []model.TimeRange{}
			byDay[value.Day] = value
			continue
		}
		enabledDays++
		start, err := minuteOfDay(value.Start)
		if err != nil {
			return nil, fmt.Errorf("%s start must use HH:MM", value.Day)
		}
		end, err := minuteOfDay(value.End)
		if err != nil || end <= start {
			return nil, fmt.Errorf("%s end must be after start", value.Day)
		}
		previousBreakEnd := start
		for _, pause := range value.Breaks {
			pauseStart, startErr := minuteOfDay(pause.Start)
			pauseEnd, endErr := minuteOfDay(pause.End)
			if startErr != nil || endErr != nil || pauseStart < previousBreakEnd || pauseEnd <= pauseStart || pauseEnd > end {
				return nil, fmt.Errorf("%s breaks must be ordered, non-overlapping, and inside working hours", value.Day)
			}
			previousBreakEnd = pauseEnd
		}
		if value.Breaks == nil {
			value.Breaks = []model.TimeRange{}
		}
		byDay[value.Day] = value
	}
	if enabledDays == 0 {
		return nil, errors.New("schedule must enable at least one day")
	}
	result := make([]model.ScheduleDay, 0, len(weekdays))
	for _, day := range weekdays {
		result = append(result, byDay[day])
	}
	return result, nil
}

func minuteOfDay(value string) (int, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil || parsed.Format("15:04") != value {
		return 0, errors.New("invalid time")
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

func slugifyEventName(name string) string {
	var slug strings.Builder
	separator := false
	for _, character := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if separator && slug.Len() > 0 {
				slug.WriteByte('-')
			}
			slug.WriteRune(character)
			separator = false
		} else if slug.Len() > 0 {
			separator = true
		}
	}
	return strings.TrimSuffix(slug.String(), "-")
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
