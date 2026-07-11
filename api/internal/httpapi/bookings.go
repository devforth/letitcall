package httpapi

import (
	"errors"
	"log/slog"
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
	type bookingListItem struct {
		model.Booking
		ManageURL string `json:"manageURL"`
	}
	items := make([]bookingListItem, 0, len(bookings))
	for _, booking := range bookings {
		secretToken, err := s.store.GetBookingSecret(booking.ID)
		if errors.Is(err, store.ErrNotFound) {
			items = append(items, bookingListItem{Booking: booking})
			continue
		}
		if err != nil {
			internalError(w, err, "load booking management link")
			return
		}
		items = append(items, bookingListItem{Booking: booking, ManageURL: s.bookingManageURL(secretToken)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookings": items})
}

type createBookingRequest struct {
	EventSlug        string   `json:"eventSlug"`
	Time             string   `json:"time"`
	AttendeeName     string   `json:"attendeeName"`
	AttendeeEmail    string   `json:"attendeeEmail"`
	AttendeeTimezone string   `json:"attendeeTimezone"`
	GuestEmails      []string `json:"guestEmails"`
	Notes            string   `json:"notes"`
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
	guestEmails, err := normalizeGuestEmails(request.GuestEmails, attendeeEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	notes := strings.TrimSpace(request.Notes)
	if len(notes) > 2000 {
		writeError(w, http.StatusBadRequest, "notes must not exceed 2000 characters")
		return
	}
	attendeeTimezone := strings.TrimSpace(request.AttendeeTimezone)
	if _, err := time.LoadLocation(attendeeTimezone); err != nil {
		writeError(w, http.StatusBadRequest, "attendeeTimezone must be an IANA timezone")
		return
	}
	id, err := security.RandomToken(12)
	if err != nil {
		internalError(w, err, "generate booking ID")
		return
	}
	secretToken, err := security.RandomToken(32)
	if err != nil {
		internalError(w, err, "generate booking secret token")
		return
	}
	now := s.now().UTC().Truncate(time.Second)
	booking := model.Booking{
		ID:               id,
		EventSlug:        eventType.EventSlug,
		Time:             bookingTime,
		EndTime:          bookingTime.Add(time.Duration(eventType.DurationMinutes) * time.Minute),
		AttendeeName:     attendeeName,
		AttendeeEmail:    attendeeEmail,
		AttendeeTimezone: attendeeTimezone,
		GuestEmails:      guestEmails,
		Notes:            notes,
		Title:            eventType.Name,
		RecipientEmails:  eventType.HostEmails(),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	busy, err := s.liveGoogleBusy(r.Context(), eventType, booking.Time, booking.EndTime)
	if err != nil {
		slog.Error("check live Google Calendar availability", "error", err, "eventType", eventType.EventSlug)
		writeError(w, http.StatusServiceUnavailable, "calendar availability could not be verified")
		return
	}
	if busy {
		writeError(w, http.StatusConflict, "selected time is no longer available")
		return
	}
	slotKey := eventType.EventSlug + booking.Time.Format(time.RFC3339) + "-" + booking.EndTime.Format(time.RFC3339)
	if err := s.store.CreateBookingWithSecret(slotKey, booking, eventType.RequiredHostEmails, eventType.InviteeLimit, secretToken); errors.Is(err, store.ErrCapacity) {
		writeError(w, http.StatusConflict, "invitee limit has been reached for this time")
		return
	} else if errors.Is(err, store.ErrBusy) {
		writeError(w, http.StatusConflict, "selected time is no longer available")
		return
	} else if errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "this invitee already has a booking at this time")
		return
	} else if err != nil {
		internalError(w, err, "create booking")
		return
	}
	s.deliverBooking(r.Context(), eventType, booking, secretToken)
	writeJSON(w, http.StatusCreated, map[string]any{"booking": booking, "manageURL": s.bookingManageURL(secretToken)})
}

type updateManagedBookingRequest struct {
	Notes       string   `json:"notes"`
	GuestEmails []string `json:"guestEmails"`
}

type cancelManagedBookingRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) getManagedBooking(w http.ResponseWriter, r *http.Request) {
	booking, eventType, ok := s.managedBooking(w, r.PathValue("secret"))
	if !ok {
		return
	}
	_, authenticated := authenticatedUser(r)
	guestLimit, err := s.managedBookingGuestLimit(booking, eventType.InviteeLimit)
	if err != nil {
		internalError(w, err, "calculate managed booking guest limit")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"booking":       booking,
		"inviteeLimit":  eventType.InviteeLimit,
		"guestLimit":    guestLimit,
		"authenticated": authenticated,
	})
}

func (s *Server) managedBookingGuestLimit(booking model.Booking, inviteeLimit *int) (*int, error) {
	if inviteeLimit == nil {
		return nil, nil
	}
	bookings, err := s.store.ListBookings()
	if err != nil {
		return nil, err
	}
	otherInvitees := 0
	for _, candidate := range bookings {
		if candidate.ID != booking.ID && candidate.EventSlug == booking.EventSlug && candidate.Time.Equal(booking.Time) && candidate.CanceledAt == nil {
			otherInvitees += 1 + len(candidate.GuestEmails)
		}
	}
	limit := max(0, *inviteeLimit-otherInvitees-1)
	return &limit, nil
}

func (s *Server) updateManagedBooking(w http.ResponseWriter, r *http.Request) {
	booking, eventType, ok := s.managedBooking(w, r.PathValue("secret"))
	if !ok {
		return
	}
	var request updateManagedBookingRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	notes := strings.TrimSpace(request.Notes)
	if len(notes) > 2000 {
		writeError(w, http.StatusBadRequest, "notes must not exceed 2000 characters")
		return
	}
	guestEmails, err := normalizeGuestEmails(request.GuestEmails, booking.AttendeeEmail)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := s.store.ModifyBooking(booking.ID, func(candidate *model.Booking, slotBookings []model.Booking) error {
		if candidate.CanceledAt != nil {
			return store.ErrCanceled
		}
		if eventType.InviteeLimit != nil {
			occupied := 0
			for _, slotBooking := range slotBookings {
				if slotBooking.ID != candidate.ID && slotBooking.CanceledAt == nil {
					occupied += 1 + len(slotBooking.GuestEmails)
				}
			}
			if occupied+1+len(guestEmails) > *eventType.InviteeLimit {
				return store.ErrCapacity
			}
		}
		candidate.Notes = notes
		candidate.GuestEmails = guestEmails
		candidate.UpdatedAt = s.now().UTC().Truncate(time.Second)
		return nil
	})
	if errors.Is(err, store.ErrCanceled) {
		writeError(w, http.StatusConflict, "booking is canceled")
		return
	}
	if errors.Is(err, store.ErrCapacity) {
		writeError(w, http.StatusConflict, "invitee limit has been reached for this time")
		return
	}
	if err != nil {
		internalError(w, err, "update managed booking")
		return
	}
	if err := s.updateCalendarEvents(r.Context(), updated, s.bookingManageURL(r.PathValue("secret"))); err != nil {
		internalError(w, err, "update booking Google events")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"booking": updated})
}

func (s *Server) cancelManagedBooking(w http.ResponseWriter, r *http.Request) {
	booking, _, ok := s.managedBooking(w, r.PathValue("secret"))
	if !ok {
		return
	}
	var request cancelManagedBookingRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	reason := strings.TrimSpace(request.Reason)
	if len(reason) > 2000 {
		writeError(w, http.StatusBadRequest, "reason must not exceed 2000 characters")
		return
	}
	actor := model.BookingActor{Name: booking.AttendeeName, Email: booking.AttendeeEmail}
	if user, authenticated := authenticatedUser(r); authenticated {
		actor = model.BookingActor{Name: user.FullName, Email: user.Email}
		if actor.Name == "" {
			actor.Name = user.Email
		}
	}
	updated, err := s.store.ModifyBooking(booking.ID, func(candidate *model.Booking, _ []model.Booking) error {
		if candidate.CanceledAt != nil {
			return store.ErrCanceled
		}
		now := s.now().UTC().Truncate(time.Second)
		candidate.CanceledAt = &now
		candidate.CanceledBy = &actor
		candidate.CancellationReason = reason
		candidate.UpdatedAt = now
		return nil
	})
	if errors.Is(err, store.ErrCanceled) {
		writeError(w, http.StatusConflict, "booking is already canceled")
		return
	}
	if err != nil {
		internalError(w, err, "cancel managed booking")
		return
	}
	manageURL := s.bookingManageURL(r.PathValue("secret"))
	if err := errors.Join(
		s.updateCalendarEvents(r.Context(), updated, manageURL),
		s.sendCancellationEmail(r.Context(), updated, manageURL),
	); err != nil {
		internalError(w, err, "deliver booking cancellation")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"booking": updated})
}

func (s *Server) managedBooking(w http.ResponseWriter, secretToken string) (model.Booking, model.EventType, bool) {
	booking, err := s.store.GetBookingBySecret(secretToken)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "booking not found")
		return model.Booking{}, model.EventType{}, false
	}
	if err != nil {
		internalError(w, err, "load booking secret link")
		return model.Booking{}, model.EventType{}, false
	}
	eventType, err := s.store.GetEventType(booking.EventSlug)
	if err != nil {
		internalError(w, err, "load managed booking event type")
		return model.Booking{}, model.EventType{}, false
	}
	return booking, eventType, true
}

func normalizeGuestEmails(values []string, attendeeEmail string) ([]string, error) {
	guests := make([]string, 0, len(values))
	seen := map[string]bool{attendeeEmail: true}
	for _, value := range values {
		email, err := security.NormalizeEmail(value)
		if err != nil {
			return nil, errors.New("guestEmails must contain valid addresses")
		}
		if seen[email] {
			return nil, errors.New("guestEmails must not contain the invitee or duplicates")
		}
		seen[email] = true
		guests = append(guests, email)
	}
	return guests, nil
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
	if bookingDate.After(today.AddDate(0, 0, eventType.BookingWindowDays)) {
		return errors.New("booking time is outside the booking window")
	}
	return nil
}

func localDate(value time.Time, location *time.Location) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}
