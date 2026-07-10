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
	apiKey string
	domain string
	from   string
	client *http.Client
}

func New(mailgun config.Mailgun) Sender {
	if !mailgun.Enabled() {
		return Disabled{}
	}
	return &Mailgun{
		apiKey: mailgun.APIKey,
		domain: mailgun.Domain,
		from:   mailgun.From,
		client: http.DefaultClient,
	}
}

func (m *Mailgun) SendBooking(ctx context.Context, booking Booking) error {
	location, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return fmt.Errorf("load recipient timezone: %w", err)
	}
	start := booking.Start.In(location)
	end := booking.End.In(location)
	values := url.Values{
		"from":    {m.from},
		"to":      {booking.RecipientEmail},
		"subject": {"New booking: " + booking.EventName},
		"text": {fmt.Sprintf(
			"%s booked %s from %s to %s (%s).",
			booking.AttendeeEmail,
			booking.EventName,
			start.Format("Monday, 2 January 2006 at 15:04"),
			end.Format("15:04"),
			booking.Timezone,
		)},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.mailgun.net/v3/"+url.PathEscape(m.domain)+"/messages",
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
