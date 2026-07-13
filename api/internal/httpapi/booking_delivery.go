package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/letitcall/letitcall/api/internal/calendar"
	"github.com/letitcall/letitcall/api/internal/mailing"
	"github.com/letitcall/letitcall/api/internal/model"
	"golang.org/x/oauth2"
)

func (s *Server) deliverBooking(ctx context.Context, eventType model.EventType, booking model.Booking, secretToken string) {
	recipients, err := s.bookingRecipients(eventType.HostEmails())
	if err != nil {
		slog.Error("load booking recipients", "error", err, "booking", booking.ID)
		return
	}
	manageURL := s.bookingManageURL(secretToken)
	var wait sync.WaitGroup
	errorsChannel := make(chan error, 2)
	wait.Add(2)
	go func() {
		defer wait.Done()
		errorsChannel <- s.addToCalendars(ctx, recipients, booking, manageURL)
	}()
	go func() {
		defer wait.Done()
		errorsChannel <- s.sendBookingEmail(ctx, recipients, booking, manageURL)
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

func (s *Server) bookingRecipients(emails []string) ([]model.User, error) {
	recipients := make([]model.User, 0, len(emails))
	for _, email := range emails {
		user, err := s.store.GetUser(email)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", email, err)
		}
		recipients = append(recipients, user)
	}
	return recipients, nil
}

func (s *Server) addToCalendars(ctx context.Context, recipients []model.User, booking model.Booking, manageURL string) error {
	eventIDs := make(map[string]string)
	var results []error
	for _, user := range recipients {
		if !user.GoogleConnected {
			continue
		}
		client, err := s.googleClient(ctx, user)
		if err != nil {
			results = append(results, err)
			continue
		}
		eventID, err := calendar.AddGoogleEvent(ctx, client, calendar.Event{
			Name:           booking.Title,
			Description:    bookingCalendarDescription(booking, manageURL),
			AttendeeEmails: bookingAttendeeEmails(booking),
			Start:          booking.Time,
			End:            booking.EndTime,
		})
		if err != nil {
			results = append(results, fmt.Errorf("add Google event for %s: %w", user.Email, err))
			continue
		}
		eventIDs[user.Email] = eventID
	}
	if len(eventIDs) > 0 {
		_, err := s.store.ModifyBooking(booking.ID, func(stored *model.Booking, _ []model.Booking) error {
			stored.GoogleEventIDs = eventIDs
			return nil
		})
		if err != nil {
			results = append(results, fmt.Errorf("store Google event IDs: %w", err))
		}
	}
	return errors.Join(results...)
}

func (s *Server) updateCalendarEvents(ctx context.Context, booking model.Booking, manageURL string) error {
	var results []error
	for email, eventID := range booking.GoogleEventIDs {
		user, err := s.store.GetUser(email)
		if err != nil {
			results = append(results, fmt.Errorf("load Google event owner %s: %w", email, err))
			continue
		}
		client, err := s.googleClient(ctx, user)
		if err != nil {
			results = append(results, err)
			continue
		}
		name := booking.Title
		if booking.CanceledAt != nil {
			name = "Canceled: " + name
		}
		if err := calendar.UpdateGoogleEvent(ctx, client, eventID, calendar.Event{
			Name:           name,
			Description:    bookingCalendarDescription(booking, manageURL),
			AttendeeEmails: bookingAttendeeEmails(booking),
		}); err != nil {
			results = append(results, fmt.Errorf("update Google event for %s: %w", email, err))
		}
	}
	return errors.Join(results...)
}

func bookingAttendeeEmails(booking model.Booking) []string {
	return append([]string{booking.AttendeeEmail}, booking.GuestEmails...)
}

func (s *Server) googleClient(ctx context.Context, user model.User) (*http.Client, error) {
	tokenJSON, err := s.tokenCipher.Decrypt(user.EncryptedGoogleToken)
	if err != nil {
		return nil, fmt.Errorf("decrypt Google token for %s: %w", user.Email, err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(tokenJSON, &token); err != nil {
		return nil, fmt.Errorf("decode Google token for %s: %w", user.Email, err)
	}
	return s.oauth.Client(ctx, &token), nil
}

func bookingCalendarDescription(booking model.Booking, manageURL string) string {
	lines := []string{
		"Booked by " + booking.AttendeeName + " <" + booking.AttendeeEmail + ">",
	}
	if len(booking.GuestEmails) > 0 {
		lines = append(lines, "Guests: "+strings.Join(booking.GuestEmails, ", "))
	}
	if booking.Notes != "" {
		lines = append(lines, "", "Booking details:", booking.Notes)
	}
	lines = append(lines, "", "Cancel or update event", manageURL)
	if booking.CanceledAt != nil && booking.CanceledBy != nil {
		lines = append(lines, "", "Canceled by "+booking.CanceledBy.Name+" <"+booking.CanceledBy.Email+"> at "+booking.CanceledAt.UTC().Format("2006-01-02 15:04:05 UTC"))
		if booking.CancellationReason != "" {
			lines = append(lines, "Reason: "+booking.CancellationReason)
		}
	}
	return strings.Join(lines, "\n")
}

func (s *Server) sendBookingEmail(ctx context.Context, recipients []model.User, booking model.Booking, manageURL string) error {
	branding, err := s.store.GetBranding()
	if err != nil {
		return err
	}
	var results []error
	for _, user := range recipients {
		if err := s.mailer.SendBooking(ctx, mailingBooking(branding.Name, user, booking, manageURL)); err != nil {
			results = append(results, fmt.Errorf("email %s: %w", user.Email, err))
		}
	}
	return errors.Join(results...)
}

func (s *Server) sendCancellationEmail(ctx context.Context, booking model.Booking, manageURL string) error {
	recipients, err := s.bookingRecipients(booking.RecipientEmails)
	if err != nil {
		return err
	}
	branding, err := s.store.GetBranding()
	if err != nil {
		return err
	}
	var results []error
	for _, user := range recipients {
		message := mailing.Cancellation{
			Booking:    mailingBooking(branding.Name, user, booking, manageURL),
			CanceledBy: booking.CanceledBy.Name + " <" + booking.CanceledBy.Email + ">",
			Reason:     booking.CancellationReason,
		}
		if err := s.mailer.SendCancellation(ctx, message); err != nil {
			results = append(results, fmt.Errorf("email %s: %w", user.Email, err))
		}
	}
	return errors.Join(results...)
}

func mailingBooking(brandName string, user model.User, booking model.Booking, manageURL string) mailing.Booking {
	recipientName := user.FullName
	if recipientName == "" {
		recipientName = user.Email
	}
	return mailing.Booking{
		BrandName:        brandName,
		EventName:        booking.Title,
		AttendeeName:     booking.AttendeeName,
		AttendeeEmail:    booking.AttendeeEmail,
		AttendeeTimezone: booking.AttendeeTimezone,
		Notes:            booking.Notes,
		Start:            booking.Time,
		End:              booking.EndTime,
		RecipientEmail:   user.Email,
		RecipientName:    recipientName,
		Timezone:         user.Timezone,
		ManageURL:        manageURL,
	}
}

func (s *Server) bookingManageURL(secretToken string) string {
	return s.cfg.HTTP.BaseURL + "/event/" + secretToken
}
