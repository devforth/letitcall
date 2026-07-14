package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

type compatibilityAvailableTime struct {
	Status            string    `json:"status"`
	InviteesRemaining int       `json:"invitees_remaining"`
	StartTime         time.Time `json:"start_time"`
	SchedulingURL     string    `json:"scheduling_url"`
}

func (s *Server) compatibilityAvailableTimes(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	slug, err := s.eventSlugForURI(values.Get("event_type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "event_type must be a valid event type URI")
		return
	}
	eventType, err := s.store.GetEventType(slug)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load event type availability")
		return
	}
	start, err := parseCompatibilityInstant(values.Get("start_time"), "start_time")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	end, err := parseCompatibilityInstant(values.Get("end_time"), "end_time")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !end.After(start) || end.Sub(start) > 7*24*time.Hour {
		writeError(w, http.StatusBadRequest, "availability range must be positive and no longer than seven days")
		return
	}
	collection, err := s.compatibilitySlots(eventType, start, end)
	if err != nil {
		internalError(w, err, "calculate compatibility availability")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"collection": collection})
}

func (s *Server) compatibilitySlots(eventType model.EventType, start, end time.Time) ([]compatibilityAvailableTime, error) {
	location, err := time.LoadLocation(eventType.Timezone)
	if err != nil {
		return nil, err
	}
	busyRanges, remaining, err := s.publicAvailability(eventType)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	today := localDate(now.In(location), location)
	firstDate := localDate(start.In(location), location)
	lastDate := localDate(end.In(location), location)
	collection := make([]compatibilityAvailableTime, 0)
	for date := firstDate; !date.After(lastDate); date = date.AddDate(0, 0, 1) {
		if date.Before(today) || date.After(today.AddDate(0, 0, eventType.BookingWindowDays)) {
			continue
		}
		day := scheduleForWeekday(eventType.Schedule, date.Weekday())
		if !day.Enabled {
			continue
		}
		workingStart, _ := minuteOfDay(day.Start)
		workingEnd, _ := minuteOfDay(day.End)
		for minute := workingStart; minute+eventType.DurationMinutes <= workingEnd; minute += eventType.DurationMinutes {
			if overlapsBreak(minute, minute+eventType.DurationMinutes, day.Breaks) {
				continue
			}
			candidate := time.Date(date.Year(), date.Month(), date.Day(), minute/60, minute%60, 0, 0, location).UTC()
			candidateEnd := candidate.Add(time.Duration(eventType.DurationMinutes) * time.Minute)
			if candidate.Before(start) || !candidate.Before(end) || !candidate.After(now) || overlapsBusy(candidate, candidateEnd, busyRanges) {
				continue
			}
			capacity := 1
			if eventType.InviteeLimit != nil {
				capacity = *eventType.InviteeLimit
				if value, ok := remaining[candidate.Format(time.RFC3339)]; ok {
					capacity = value
				}
			}
			collection = append(collection, compatibilityAvailableTime{
				Status: "available", InviteesRemaining: capacity, StartTime: candidate,
				SchedulingURL: s.cfg.HTTP.BaseURL + "/book/" + eventType.EventSlug,
			})
		}
	}
	return collection, nil
}

func scheduleForWeekday(schedule []model.ScheduleDay, weekday time.Weekday) model.ScheduleDay {
	name := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}[weekday]
	for _, day := range schedule {
		if day.Day == name {
			return day
		}
	}
	return model.ScheduleDay{}
}

func overlapsBreak(start, end int, breaks []model.TimeRange) bool {
	for _, pause := range breaks {
		pauseStart, _ := minuteOfDay(pause.Start)
		pauseEnd, _ := minuteOfDay(pause.End)
		if start < pauseEnd && end > pauseStart {
			return true
		}
	}
	return false
}

func overlapsBusy(start, end time.Time, ranges []busyRange) bool {
	for _, value := range ranges {
		if start.Before(value.End) && end.After(value.Start) {
			return true
		}
	}
	return false
}
