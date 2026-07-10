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
	Time          string `json:"time"`
	AttendeeEmail string `json:"attendeeEmail"`
	Title         string `json:"title"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	var request createBookingRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	key, bookingTime, err := bookingKey(request.Time)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	attendeeEmail, err := security.NormalizeEmail(request.AttendeeEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, "attendeeEmail must be a valid address")
		return
	}
	title := strings.TrimSpace(request.Title)
	if title == "" || len(title) > 200 {
		writeError(w, http.StatusBadRequest, "title must be between 1 and 200 characters")
		return
	}
	booking := model.Booking{
		Time:          bookingTime,
		OwnerEmail:    userFromRequest(r).Email,
		AttendeeEmail: attendeeEmail,
		Title:         title,
		CreatedAt:     s.now().UTC().Truncate(time.Second),
	}
	if err := s.store.CreateBooking(key, booking); errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "a booking already exists at this UTC time")
		return
	} else if err != nil {
		internalError(w, err, "create booking")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

func (s *Server) deleteBooking(w http.ResponseWriter, r *http.Request) {
	key, _, err := bookingKey(r.PathValue("time"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	booking, err := s.store.GetBooking(key)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	if err != nil {
		internalError(w, err, "load booking for deletion")
		return
	}
	if booking.OwnerEmail != userFromRequest(r).Email {
		writeError(w, http.StatusForbidden, "only the booking owner can delete it")
		return
	}
	if err := s.store.DeleteBooking(key); err != nil {
		internalError(w, err, "delete booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
