package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

type compatibilityInviteesCounter struct {
	Total  int `json:"total"`
	Active int `json:"active"`
	Limit  int `json:"limit"`
}

type compatibilityEventMembership struct {
	User      string `json:"user"`
	UserEmail string `json:"user_email"`
	UserName  string `json:"user_name"`
}

type compatibilityEventGuest struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type compatibilityCancellation struct {
	CanceledBy   string    `json:"canceled_by"`
	Reason       *string   `json:"reason"`
	CancelerType string    `json:"canceler_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type compatibilityScheduledEvent struct {
	URI              string                         `json:"uri"`
	Name             string                         `json:"name"`
	Status           string                         `json:"status"`
	BookingMethod    string                         `json:"booking_method"`
	StartTime        time.Time                      `json:"start_time"`
	EndTime          time.Time                      `json:"end_time"`
	EventType        string                         `json:"event_type"`
	InviteesCounter  compatibilityInviteesCounter   `json:"invitees_counter"`
	CreatedAt        time.Time                      `json:"created_at"`
	UpdatedAt        time.Time                      `json:"updated_at"`
	EventMemberships []compatibilityEventMembership `json:"event_memberships"`
	EventGuests      []compatibilityEventGuest      `json:"event_guests"`
	Cancellation     *compatibilityCancellation     `json:"cancellation,omitempty"`
	// TODO: Add location, calendar_event, and host meeting notes when bookings store those concepts.
}

type compatibilityQuestionAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Position int    `json:"position"`
}

type compatibilityInvitee struct {
	URI                 string                        `json:"uri"`
	Email               string                        `json:"email"`
	FirstName           *string                       `json:"first_name"`
	LastName            *string                       `json:"last_name"`
	Name                string                        `json:"name"`
	Status              string                        `json:"status"`
	QuestionsAndAnswers []compatibilityQuestionAnswer `json:"questions_and_answers"`
	Timezone            string                        `json:"timezone"`
	Event               string                        `json:"event"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	Rescheduled         bool                          `json:"rescheduled"`
	CancelURL           string                        `json:"cancel_url"`
	Cancellation        *compatibilityCancellation    `json:"cancellation,omitempty"`
	// TODO: Add tracking, reminders, reschedule links, routing forms, payments, no-show, and reconfirmation when stored.
}

type scheduledEventGroup struct {
	EventSlug string
	Start     time.Time
	End       time.Time
	Bookings  []model.Booking
}

func (s *Server) compatibilityListScheduledEvents(w http.ResponseWriter, r *http.Request) {
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
			internalError(w, err, "load scheduled event user")
			return
		}
		userEmail = user.Email
	}
	minStart, err := parseOptionalCompatibilityInstant(values, "min_start_time")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	maxStart, err := parseOptionalCompatibilityInstant(values, "max_start_time")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	status := values.Get("status")
	if status != "" && status != "active" && status != "canceled" {
		writeError(w, http.StatusBadRequest, "status must be active or canceled")
		return
	}
	groups, err := s.scheduledEventGroups()
	if err != nil {
		internalError(w, err, "group scheduled events")
		return
	}
	items := make([]compatibilityScheduledEvent, 0, len(groups))
	for _, group := range groups {
		if userEmail != "" && !groupHasHost(group, userEmail) {
			continue
		}
		if email := values.Get("invitee_email"); email != "" && !groupHasInvitee(group, email) {
			continue
		}
		if minStart != nil && group.Start.Before(*minStart) {
			continue
		}
		if maxStart != nil && !group.Start.Before(*maxStart) {
			continue
		}
		item, err := s.compatibilityScheduledEvent(group)
		if err != nil {
			internalError(w, err, "serialize scheduled event")
			return
		}
		if status != "" && item.Status != status {
			continue
		}
		items = append(items, item)
	}
	sortValue := values.Get("sort")
	if sortValue == "" || sortValue == "start_time:asc" {
		sort.SliceStable(items, func(i, j int) bool { return items[i].StartTime.Before(items[j].StartTime) })
	} else if sortValue == "start_time:desc" {
		sort.SliceStable(items, func(i, j int) bool { return items[i].StartTime.After(items[j].StartTime) })
	} else {
		writeError(w, http.StatusBadRequest, "sort must be start_time:asc or start_time:desc")
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

func (s *Server) compatibilityListInvitees(w http.ResponseWriter, r *http.Request) {
	groups, err := s.scheduledEventGroups()
	if err != nil {
		internalError(w, err, "group event invitees")
		return
	}
	var selected *scheduledEventGroup
	for index := range groups {
		if s.scheduledEventID(groups[index].EventSlug, groups[index].Start, groups[index].End) == r.PathValue("event_uuid") {
			selected = &groups[index]
			break
		}
	}
	if selected == nil {
		writeError(w, http.StatusNotFound, "scheduled event not found")
		return
	}
	values := r.URL.Query()
	status := values.Get("status")
	if status != "" && status != "active" && status != "canceled" {
		writeError(w, http.StatusBadRequest, "status must be active or canceled")
		return
	}
	items := make([]compatibilityInvitee, 0, len(selected.Bookings))
	for _, booking := range selected.Bookings {
		item, err := s.compatibilityInvitee(*selected, booking)
		if err != nil {
			internalError(w, err, "serialize event invitee")
			return
		}
		if status != "" && item.Status != status {
			continue
		}
		if email := values.Get("email"); email != "" && !strings.EqualFold(email, item.Email) {
			continue
		}
		items = append(items, item)
	}
	sortValue := values.Get("sort")
	if sortValue == "" || sortValue == "created_at:asc" {
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	} else if sortValue == "created_at:desc" {
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	} else {
		writeError(w, http.StatusBadRequest, "sort must be created_at:asc or created_at:desc")
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

func (s *Server) scheduledEventGroups() ([]scheduledEventGroup, error) {
	bookings, err := s.store.ListBookings()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]int)
	groups := make([]scheduledEventGroup, 0)
	for _, booking := range bookings {
		id := s.scheduledEventID(booking.EventSlug, booking.Time, booking.EndTime)
		index, ok := byID[id]
		if !ok {
			index = len(groups)
			byID[id] = index
			groups = append(groups, scheduledEventGroup{EventSlug: booking.EventSlug, Start: booking.Time, End: booking.EndTime})
		}
		groups[index].Bookings = append(groups[index].Bookings, booking)
	}
	return groups, nil
}

func (s *Server) compatibilityScheduledEvent(group scheduledEventGroup) (compatibilityScheduledEvent, error) {
	first := group.Bookings[0]
	createdAt, updatedAt := first.CreatedAt, first.UpdatedAt
	active := 0
	guests := make([]compatibilityEventGuest, 0)
	membershipEmails := make(map[string]bool)
	var latestCanceled *model.Booking
	for index := range group.Bookings {
		booking := &group.Bookings[index]
		if booking.CreatedAt.Before(createdAt) {
			createdAt = booking.CreatedAt
		}
		if booking.UpdatedAt.After(updatedAt) {
			updatedAt = booking.UpdatedAt
		}
		if booking.CanceledAt == nil {
			active++
		} else if latestCanceled == nil || booking.CanceledAt.After(*latestCanceled.CanceledAt) {
			latestCanceled = booking
		}
		for _, email := range booking.RecipientEmails {
			membershipEmails[strings.ToLower(email)] = true
		}
		for _, email := range booking.GuestEmails {
			guests = append(guests, compatibilityEventGuest{Email: email, CreatedAt: booking.CreatedAt, UpdatedAt: booking.UpdatedAt})
		}
	}
	memberships := make([]compatibilityEventMembership, 0, len(membershipEmails))
	for email := range membershipEmails {
		name := email
		if user, err := s.store.GetUser(email); err == nil && user.FullName != "" {
			name = user.FullName
		}
		memberships = append(memberships, compatibilityEventMembership{User: s.userURI(email), UserEmail: email, UserName: name})
	}
	sort.Slice(memberships, func(i, j int) bool { return memberships[i].UserEmail < memberships[j].UserEmail })
	limit := 1
	if eventType, err := s.store.GetEventType(group.EventSlug); err == nil && eventType.InviteeLimit != nil {
		limit = *eventType.InviteeLimit
	}
	status := "active"
	var cancellation *compatibilityCancellation
	if active == 0 {
		status = "canceled"
		cancellation = compatibilityBookingCancellation(*latestCanceled)
	}
	return compatibilityScheduledEvent{
		URI: s.scheduledEventURI(group.EventSlug, group.Start, group.End), Name: first.Title,
		Status: status, BookingMethod: "instant", StartTime: group.Start, EndTime: group.End,
		EventType:       s.eventTypeURI(group.EventSlug),
		InviteesCounter: compatibilityInviteesCounter{Total: len(group.Bookings), Active: active, Limit: limit},
		CreatedAt:       createdAt, UpdatedAt: updatedAt, EventMemberships: memberships,
		EventGuests: guests, Cancellation: cancellation,
	}, nil
}

func (s *Server) compatibilityInvitee(group scheduledEventGroup, booking model.Booking) (compatibilityInvitee, error) {
	parts := strings.Fields(booking.AttendeeName)
	firstName := parts[0]
	var lastName *string
	if len(parts) > 1 {
		value := strings.Join(parts[1:], " ")
		lastName = &value
	}
	questions := make([]compatibilityQuestionAnswer, 0, 1)
	if booking.Notes != "" {
		questions = append(questions, compatibilityQuestionAnswer{Question: bookingNotesQuestion, Answer: booking.Notes, Position: 0})
	}
	status := "active"
	if booking.CanceledAt != nil {
		status = "canceled"
	}
	secret, err := s.store.GetBookingSecret(booking.ID)
	if err != nil {
		return compatibilityInvitee{}, err
	}
	eventURI := s.scheduledEventURI(group.EventSlug, group.Start, group.End)
	return compatibilityInvitee{
		URI: s.inviteeURI(group.EventSlug, group.Start, group.End, booking.ID), Email: booking.AttendeeEmail,
		FirstName: &firstName, LastName: lastName, Name: booking.AttendeeName, Status: status,
		QuestionsAndAnswers: questions, Timezone: booking.AttendeeTimezone, Event: eventURI,
		CreatedAt: booking.CreatedAt, UpdatedAt: booking.UpdatedAt, Rescheduled: false,
		CancelURL: s.bookingManageURL(secret), Cancellation: compatibilityBookingCancellation(booking),
	}, nil
}

func compatibilityBookingCancellation(booking model.Booking) *compatibilityCancellation {
	if booking.CanceledAt == nil || booking.CanceledBy == nil {
		return nil
	}
	var reason *string
	if booking.CancellationReason != "" {
		value := booking.CancellationReason
		reason = &value
	}
	cancelerType := "invitee"
	if containsEmail(booking.RecipientEmails, booking.CanceledBy.Email) {
		cancelerType = "host"
	}
	return &compatibilityCancellation{
		CanceledBy: booking.CanceledBy.Email, Reason: reason, CancelerType: cancelerType, CreatedAt: *booking.CanceledAt,
	}
}

func groupHasHost(group scheduledEventGroup, email string) bool {
	for _, booking := range group.Bookings {
		if containsEmail(booking.RecipientEmails, email) {
			return true
		}
	}
	return false
}

func groupHasInvitee(group scheduledEventGroup, email string) bool {
	for _, booking := range group.Bookings {
		if strings.EqualFold(booking.AttendeeEmail, email) {
			return true
		}
	}
	return false
}
