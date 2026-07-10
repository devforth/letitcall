package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/letitcall/letitcall/api/internal/calendar"
	"github.com/letitcall/letitcall/api/internal/mailing"
	"github.com/letitcall/letitcall/api/internal/model"
	"golang.org/x/oauth2"
)

func (s *Server) deliverBooking(ctx context.Context, eventType model.EventType, booking model.Booking) {
	recipients := make([]model.User, 0, len(eventType.RecipientEmails))
	for _, email := range eventType.RecipientEmails {
		user, err := s.store.GetUser(email)
		if err != nil {
			slog.Error("load booking recipient", "error", err, "email", email)
			return
		}
		recipients = append(recipients, user)
	}
	var wait sync.WaitGroup
	errorsChannel := make(chan error, 2)
	wait.Add(2)
	go func() {
		defer wait.Done()
		errorsChannel <- s.addToCalendars(ctx, recipients, booking)
	}()
	go func() {
		defer wait.Done()
		errorsChannel <- s.sendToEmail(ctx, recipients, booking)
	}()
	wait.Wait()
	close(errorsChannel)
	var deliveryErrors []error
	for err := range errorsChannel {
		if err != nil {
			deliveryErrors = append(deliveryErrors, err)
		}
	}
	if err := errors.Join(deliveryErrors...); err != nil {
		slog.Error("deliver booking", "error", err, "booking", booking.ID)
	}
}

func (s *Server) addToCalendars(ctx context.Context, recipients []model.User, booking model.Booking) error {
	var results []error
	for _, user := range recipients {
		if !user.GoogleConnected {
			continue
		}
		tokenJSON, err := s.tokenCipher.Decrypt(user.EncryptedGoogleToken)
		if err != nil {
			results = append(results, fmt.Errorf("decrypt Google token for %s: %w", user.Email, err))
			continue
		}
		var token oauth2.Token
		if err := json.Unmarshal(tokenJSON, &token); err != nil {
			results = append(results, fmt.Errorf("decode Google token for %s: %w", user.Email, err))
			continue
		}
		if err := calendar.AddGoogleEvent(ctx, s.oauth.Client(ctx, &token), calendar.Event{
			Name:          booking.Title,
			AttendeeEmail: booking.AttendeeEmail,
			Start:         booking.Time,
			End:           booking.EndTime,
		}); err != nil {
			results = append(results, fmt.Errorf("add Google event for %s: %w", user.Email, err))
		}
	}
	return errors.Join(results...)
}

func (s *Server) sendToEmail(ctx context.Context, recipients []model.User, booking model.Booking) error {
	var results []error
	for _, user := range recipients {
		if err := s.mailer.SendBooking(ctx, mailing.Booking{
			EventName:      booking.Title,
			AttendeeEmail:  booking.AttendeeEmail,
			Start:          booking.Time,
			End:            booking.EndTime,
			RecipientEmail: user.Email,
			Timezone:       user.Timezone,
		}); err != nil {
			results = append(results, fmt.Errorf("email %s: %w", user.Email, err))
		}
	}
	return errors.Join(results...)
}
