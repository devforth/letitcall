package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/calendar"
	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

type userDeletionImpact struct {
	RequiresReassignment bool       `json:"requiresReassignment"`
	FutureBookingCount   int        `json:"futureBookingCount"`
	EarliestBookingAt    *time.Time `json:"earliestBookingAt,omitempty"`
	LatestBookingAt      *time.Time `json:"latestBookingAt,omitempty"`
}

func (s *Server) getUserDeletionImpact(w http.ResponseWriter, r *http.Request) {
	email, err := security.NormalizeEmail(r.PathValue("email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := s.store.GetUser(email); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		internalError(w, err, "load user for deletion impact")
		return
	}
	bookings, err := s.futureSoleHostBookings(email, s.now().UTC().Truncate(time.Second))
	if err != nil {
		internalError(w, err, "calculate user deletion impact")
		return
	}
	impact := userDeletionImpact{
		RequiresReassignment: len(bookings) > 0,
		FutureBookingCount:   len(bookings),
	}
	if len(bookings) > 0 {
		earliest := bookings[0].Time.UTC().Truncate(time.Second)
		latest := bookings[len(bookings)-1].Time.UTC().Truncate(time.Second)
		impact.EarliestBookingAt = &earliest
		impact.LatestBookingAt = &latest
	}
	writeJSON(w, http.StatusOK, impact)
}

type reassignUserBookingsRequest struct {
	NewHostEmail string `json:"newHostEmail"`
}

func (s *Server) reassignUserBookings(w http.ResponseWriter, r *http.Request) {
	oldEmail, err := security.NormalizeEmail(r.PathValue("email"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	oldHost, err := s.store.GetUser(oldEmail)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		internalError(w, err, "load user for booking reassignment")
		return
	}
	var request reassignUserBookingsRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	newEmail, err := security.NormalizeEmail(request.NewHostEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, "newHostEmail must be a valid address")
		return
	}
	if strings.EqualFold(oldEmail, newEmail) {
		writeError(w, http.StatusBadRequest, "newHostEmail must select another user")
		return
	}
	newHost, err := s.store.GetUser(newEmail)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "new host user not found")
		return
	}
	if err != nil {
		internalError(w, err, "load new booking host")
		return
	}
	now := s.now().UTC().Truncate(time.Second)
	bookings, err := s.futureSoleHostBookings(oldEmail, now)
	if err != nil {
		internalError(w, err, "load bookings for host reassignment")
		return
	}
	if len(bookings) == 0 {
		writeError(w, http.StatusConflict, "user has no upcoming bookings to reassign")
		return
	}

	var oldCalendar, newCalendar *http.Client
	if oldHost.GoogleConnected {
		oldCalendar, err = s.googleClient(r.Context(), oldHost)
		if err != nil {
			internalError(w, err, "connect old host Google Calendar")
			return
		}
	}
	if newHost.GoogleConnected {
		newCalendar, err = s.googleClient(r.Context(), newHost)
		if err != nil {
			internalError(w, err, "connect new host Google Calendar")
			return
		}
	}

	for _, booking := range bookings {
		updated := booking
		updated.RecipientEmails = replaceEmail(updated.RecipientEmails, oldEmail, newEmail)
		updated.GoogleEventIDs = cloneEventIDs(updated.GoogleEventIDs)
		if newCalendar != nil && updated.GoogleEventIDs[newEmail] == "" {
			secret, err := s.store.GetBookingSecret(booking.ID)
			if err != nil {
				internalError(w, err, "load booking management link for host reassignment")
				return
			}
			eventID, err := calendar.AddGoogleEvent(r.Context(), newCalendar, calendar.Event{
				Name:           booking.Title,
				Description:    bookingCalendarDescription(booking, s.bookingManageURL(secret)),
				AttendeeEmails: bookingAttendeeEmails(booking),
				Start:          booking.Time,
				End:            booking.EndTime,
			})
			if err != nil {
				internalError(w, err, "add reassigned booking to new host Google Calendar")
				return
			}
			updated.GoogleEventIDs[newEmail] = eventID
		}
		if oldEventID := updated.GoogleEventIDs[oldEmail]; oldEventID != "" {
			if err := calendar.DeleteGoogleEvent(r.Context(), oldCalendar, oldEventID); err != nil {
				internalError(w, err, "remove reassigned booking from old host Google Calendar")
				return
			}
			delete(updated.GoogleEventIDs, oldEmail)
		}
		updated.UpdatedAt = now
		if _, err := s.store.ModifyBooking(booking.ID, func(candidate *model.Booking, _ []model.Booking) error {
			*candidate = updated
			return nil
		}); err != nil {
			internalError(w, err, "store reassigned booking")
			return
		}
	}
	if err := s.store.ReassignSoleRequiredHost(oldEmail, newEmail, now); err != nil {
		internalError(w, err, "reassign event type host")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reassignedBookingCount": len(bookings)})
}

func (s *Server) futureSoleHostBookings(email string, now time.Time) ([]model.Booking, error) {
	eventTypes, err := s.store.ListEventTypes()
	if err != nil {
		return nil, err
	}
	soleHostEventSlugs := make(map[string]bool)
	for _, eventType := range eventTypes {
		if len(eventType.RequiredHostEmails) == 1 && strings.EqualFold(eventType.RequiredHostEmails[0], email) {
			soleHostEventSlugs[eventType.EventSlug] = true
		}
	}
	bookings, err := s.store.ListBookings()
	if err != nil {
		return nil, err
	}
	impacted := make([]model.Booking, 0)
	for _, booking := range bookings {
		if booking.CanceledAt == nil && booking.Time.After(now) && soleHostEventSlugs[booking.EventSlug] && deletionEmailListContains(booking.RecipientEmails, email) {
			impacted = append(impacted, booking)
		}
	}
	return impacted, nil
}

func deletionEmailListContains(values []string, email string) bool {
	for _, value := range values {
		if strings.EqualFold(value, email) {
			return true
		}
	}
	return false
}

func replaceEmail(values []string, oldEmail, newEmail string) []string {
	replaced := make([]string, 0, len(values))
	for _, value := range values {
		if strings.EqualFold(value, oldEmail) {
			value = newEmail
		}
		if !deletionEmailListContains(replaced, value) {
			replaced = append(replaced, value)
		}
	}
	return replaced
}

func cloneEventIDs(values map[string]string) map[string]string {
	cloned := make(map[string]string, len(values))
	for email, eventID := range values {
		cloned[email] = eventID
	}
	return cloned
}
