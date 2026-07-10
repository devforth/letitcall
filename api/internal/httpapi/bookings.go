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

func (s *Server) listBookings(w http.ResponseWriter, _ *http.Request) {
	bookings, err := s.store.ListBookings()
	if err != nil {
		internalError(w, err, "list bookings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

type createBookingRequest struct {
	EventSlug     string `json:"eventSlug"`
	Time          string `json:"time"`
	AttendeeName  string `json:"attendeeName"`
	AttendeeEmail string `json:"attendeeEmail"`
	Notes         string `json:"notes"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	var request createBookingRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	eventType, err := s.store.GetEventType(strings.TrimSpace(request.EventSlug))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "event type not found")
		return
	}
	if err != nil {
		internalError(w, err, "load event type for booking")
		return
	}
	_, bookingTime, err := bookingKey(request.Time)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.validateBookingTime(eventType, bookingTime); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	attendeeEmail, err := security.NormalizeEmail(request.AttendeeEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, "attendeeEmail must be a valid address")
		return
	}
	attendeeName := strings.TrimSpace(request.AttendeeName)
	if attendeeName == "" || len(attendeeName) > 200 {
		writeError(w, http.StatusBadRequest, "attendeeName must be between 1 and 200 characters")
		return
	}
	notes := strings.TrimSpace(request.Notes)
	if len(notes) > 2000 {
		writeError(w, http.StatusBadRequest, "notes must not exceed 2000 characters")
		return
	}
	id, err := security.RandomToken(12)
	if err != nil {
		internalError(w, err, "generate booking ID")
		return
	}
	booking := model.Booking{
		ID:              id,
		EventSlug:       eventType.EventSlug,
		Time:            bookingTime,
		EndTime:         bookingTime.Add(time.Duration(eventType.DurationMinutes) * time.Minute),
		AttendeeName:    attendeeName,
		AttendeeEmail:   attendeeEmail,
		Notes:           notes,
		Title:           eventType.Name,
		RecipientEmails: append([]string(nil), eventType.RecipientEmails...),
		CreatedAt:       s.now().UTC().Truncate(time.Second),
	}
	slotKey := eventType.EventSlug + booking.Time.Format(time.RFC3339) + "-" + booking.EndTime.Format(time.RFC3339)
	if err := s.store.CreateBooking(slotKey, booking, eventType.InviteeLimit); errors.Is(err, store.ErrCapacity) {
		writeError(w, http.StatusConflict, "invitee limit has been reached for this time")
		return
	} else if errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "this invitee already has a booking at this time")
		return
	} else if err != nil {
		internalError(w, err, "create booking")
		return
	}
	s.deliverBooking(r.Context(), eventType, booking)
	writeJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

func (s *Server) deleteBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.store.GetBooking(id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	} else if err != nil {
		internalError(w, err, "load booking for deletion")
		return
	}
	if err := s.store.DeleteBooking(id); err != nil {
		internalError(w, err, "delete booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) validateBookingTime(eventType model.EventType, bookingTime time.Time) error {
	if !bookingTime.After(s.now().UTC()) {
		return errors.New("booking time must be in the future")
	}
	location, err := time.LoadLocation(eventType.Timezone)
	if err != nil {
		return err
	}
	local := bookingTime.In(location)
	localEnd := local.Add(time.Duration(eventType.DurationMinutes) * time.Minute)
	dayName := strings.ToLower(local.Weekday().String())
	var day model.ScheduleDay
	for _, candidate := range eventType.Schedule {
		if candidate.Day == dayName {
			day = candidate
			break
		}
	}
	if !day.Enabled {
		return errors.New("booking time is outside event availability")
	}
	startMinute := local.Hour()*60 + local.Minute()
	endMinute := localEnd.Hour()*60 + localEnd.Minute()
	workingStart, _ := minuteOfDay(day.Start)
	workingEnd, _ := minuteOfDay(day.End)
	if local.YearDay() != localEnd.YearDay() || startMinute < workingStart || endMinute > workingEnd {
		return errors.New("booking time is outside event availability")
	}
	for _, pause := range day.Breaks {
		pauseStart, _ := minuteOfDay(pause.Start)
		pauseEnd, _ := minuteOfDay(pause.End)
		if startMinute < pauseEnd && endMinute > pauseStart {
			return errors.New("booking time overlaps an availability break")
		}
	}
	today := localDate(s.now().In(location), location)
	bookingDate := localDate(local, location)
	if bookingDate.Before(today) {
		return errors.New("booking time must be in the future")
	}
	if eventType.BookingWindowDays != nil && bookingDate.After(today.AddDate(0, 0, *eventType.BookingWindowDays)) {
		return errors.New("booking time is outside the booking window")
	}
	return nil
}

func localDate(value time.Time, location *time.Location) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}
