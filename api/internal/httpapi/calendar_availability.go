package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/calendar"
	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

const calendarSyncInterval = 20 * time.Second

type busyRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (s *Server) RunCalendarSync(ctx context.Context) {
	for {
		startedAt := time.Now()
		s.SyncGoogleCalendars(ctx)
		wait := time.Until(startedAt.Add(calendarSyncInterval))
		if wait < 0 {
			wait = 0
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *Server) SyncGoogleCalendars(ctx context.Context) {
	if s.oauth == nil {
		return
	}
	days, err := s.calendarSyncDays()
	if err != nil {
		slog.Error("calculate Google Calendar sync window", "error", err)
		return
	}
	if days == 0 {
		return
	}
	users, err := s.store.ListUsers()
	if err != nil {
		slog.Error("list Google Calendar sync users", "error", err)
		return
	}
	connected := make([]model.User, 0, len(users))
	for _, user := range users {
		if user.GoogleConnected {
			connected = append(connected, user)
		}
	}
	for index, user := range connected {
		if ctx.Err() != nil {
			return
		}
		s.syncGoogleCalendar(ctx, user, days)
		if index < len(connected)-1 {
			timer := time.NewTimer(calendarSyncInterval / time.Duration(len(connected)))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func (s *Server) syncGoogleCalendar(ctx context.Context, user model.User, days int) {
	client, err := s.googleClient(ctx, user)
	if err != nil {
		slog.Error("create Google Calendar sync client", "error", err, "user", user.Email)
		return
	}
	start := s.now().UTC().Truncate(time.Second)
	periods, err := calendar.ListGoogleBusy(ctx, client, start, start.AddDate(0, 0, days+1))
	if err != nil {
		slog.Error("sync Google Calendar", "error", err, "user", user.Email)
		return
	}
	owned, err := s.googleEventIDs(user.Email)
	if err != nil {
		slog.Error("load LetItCall Google event IDs", "error", err, "user", user.Email)
		return
	}
	cache := model.GoogleBusyCache{Periods: make([]model.GoogleBusyPeriod, 0, len(periods)), SyncedAt: s.now().UTC().Truncate(time.Second)}
	for _, period := range periods {
		if !owned[period.EventID] {
			cache.Periods = append(cache.Periods, model.GoogleBusyPeriod{EventID: period.EventID, Start: period.Start, End: period.End})
		}
	}
	if err := s.store.PutGoogleBusy(user.Email, cache); err != nil {
		slog.Error("store Google Calendar busy periods", "error", err, "user", user.Email)
	}
}

func (s *Server) calendarSyncDays() (int, error) {
	eventTypes, err := s.store.ListEventTypes()
	if err != nil {
		return 0, err
	}
	days := 0
	for _, eventType := range eventTypes {
		days = max(days, eventType.BookingWindowDays)
	}
	return days, nil
}

func (s *Server) googleEventIDs(email string) (map[string]bool, error) {
	bookings, err := s.store.ListBookings()
	if err != nil {
		return nil, err
	}
	ids := make(map[string]bool)
	for _, booking := range bookings {
		if id := booking.GoogleEventIDs[email]; id != "" {
			ids[id] = true
		}
	}
	return ids, nil
}

func (s *Server) liveGoogleBusy(ctx context.Context, eventType model.EventType, start, end time.Time) (bool, error) {
	for _, email := range eventType.RequiredHostEmails {
		user, err := s.store.GetUser(email)
		if err != nil {
			return false, err
		}
		if !user.GoogleConnected {
			continue
		}
		client, err := s.googleClient(ctx, user)
		if err != nil {
			return false, err
		}
		periods, err := calendar.ListGoogleBusy(ctx, client, start, end)
		if err != nil {
			return false, err
		}
		owned, err := s.googleEventIDs(user.Email)
		if err != nil {
			return false, err
		}
		for _, period := range periods {
			if !owned[period.EventID] && start.Before(period.End) && end.After(period.Start) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *Server) publicAvailability(eventType model.EventType) ([]busyRange, map[string]int, error) {
	ranges := make([]busyRange, 0)
	for _, email := range eventType.RequiredHostEmails {
		cache, err := s.store.GetGoogleBusy(email)
		if errors.Is(err, store.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		for _, period := range cache.Periods {
			ranges = append(ranges, busyRange{Start: period.Start, End: period.End})
		}
	}

	required := make(map[string]bool, len(eventType.RequiredHostEmails))
	for _, email := range eventType.RequiredHostEmails {
		required[strings.ToLower(email)] = true
	}
	bookings, err := s.store.ListBookings()
	if err != nil {
		return nil, nil, err
	}
	type slot struct {
		start    time.Time
		end      time.Time
		invitees int
	}
	slots := make(map[string]slot)
	for _, booking := range bookings {
		if booking.CanceledAt != nil {
			continue
		}
		if booking.EventSlug == eventType.EventSlug {
			key := booking.Time.Format(time.RFC3339)
			value := slots[key]
			value.start = booking.Time
			value.end = booking.EndTime
			value.invitees += 1 + len(booking.GuestEmails)
			slots[key] = value
			continue
		}
		for _, email := range booking.RecipientEmails {
			if required[strings.ToLower(email)] {
				ranges = append(ranges, busyRange{Start: booking.Time, End: booking.EndTime})
				break
			}
		}
	}
	remaining := make(map[string]int)
	for key, slot := range slots {
		if eventType.InviteeLimit == nil {
			ranges = append(ranges, busyRange{Start: slot.start, End: slot.end})
			continue
		}
		remaining[key] = max(0, *eventType.InviteeLimit-slot.invitees)
		if slot.invitees >= *eventType.InviteeLimit {
			ranges = append(ranges, busyRange{Start: slot.start, End: slot.end})
		}
	}
	return ranges, remaining, nil
}
