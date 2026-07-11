package mailing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/config"
)

type Mailgun struct {
	apiKey    string
	baseURL   string
	domain    string
	from      string
	client    *http.Client
	renderer  *Renderer
	brandName string
}

func New(mailgun config.Mailgun, renderer *Renderer, brandName string) Sender {
	if !mailgun.Enabled() {
		return Disabled{}
	}
	return &Mailgun{
		apiKey:    mailgun.APIKey,
		baseURL:   mailgun.BaseURL,
		domain:    mailgun.Domain,
		from:      mailgun.From,
		client:    http.DefaultClient,
		renderer:  renderer,
		brandName: brandName,
	}
}

func (m *Mailgun) SendBooking(ctx context.Context, booking Booking) error {
	return m.send(ctx, booking.RecipientEmail, func(eventDateTime string) (Message, error) {
		return m.renderer.RenderNewEvent(TemplateData{
			BrandName:        m.brandName,
			Subject:          fmt.Sprintf("New Event - %s - %s - %s", booking.AttendeeName, eventDateTime, booking.EventName),
			RecipientName:    booking.RecipientName,
			EventName:        booking.EventName,
			AttendeeName:     booking.AttendeeName,
			AttendeeEmail:    booking.AttendeeEmail,
			AttendeeTimezone: booking.AttendeeTimezone,
			EventDateTime:    eventDateTime,
			Notes:            booking.Notes,
			ManageURL:        booking.ManageURL,
		})
	}, booking)
}

func (m *Mailgun) SendCancellation(ctx context.Context, cancellation Cancellation) error {
	return m.send(ctx, cancellation.RecipientEmail, func(eventDateTime string) (Message, error) {
		return m.renderer.RenderCancellation(TemplateData{
			BrandName:     m.brandName,
			Subject:       fmt.Sprintf("Canceled Event - %s - %s - %s", cancellation.AttendeeName, eventDateTime, cancellation.EventName),
			RecipientName: cancellation.RecipientName,
			EventName:     cancellation.EventName,
			AttendeeName:  cancellation.AttendeeName,
			AttendeeEmail: cancellation.AttendeeEmail,
			EventDateTime: eventDateTime,
			CanceledBy:    cancellation.CanceledBy,
			Reason:        cancellation.Reason,
			ManageURL:     cancellation.ManageURL,
		})
	}, cancellation.Booking)
}

func (m *Mailgun) send(ctx context.Context, recipient string, render func(string) (Message, error), booking Booking) error {
	location, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return fmt.Errorf("load recipient timezone: %w", err)
	}
	start := booking.Start.In(location)
	end := booking.End.In(location)
	eventDateTime := fmt.Sprintf("%s–%s, %s (%s)", start.Format("15:04"), end.Format("15:04"), start.Format("Monday, 2 January 2006"), booking.Timezone)
	message, err := render(eventDateTime)
	if err != nil {
		return err
	}
	values := url.Values{
		"from":    {m.from},
		"to":      {recipient},
		"subject": {message.Subject},
		"text":    {message.Text},
		"html":    {message.HTML},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		m.baseURL+"/v3/"+url.PathEscape(m.domain)+"/messages",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}
	request.SetBasicAuth("api", m.apiKey)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := m.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("Mailgun returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return nil
}
