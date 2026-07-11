package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const googleEventsURL = "https://www.googleapis.com/calendar/v3/calendars/primary/events"

type Event struct {
	Name           string
	Description    string
	AttendeeEmails []string
	Start          time.Time
	End            time.Time
}

func AddGoogleEvent(ctx context.Context, client *http.Client, event Event) (string, error) {
	payload := struct {
		Summary     string     `json:"summary"`
		Description string     `json:"description"`
		Attendees   []attendee `json:"attendees"`
		Start       date       `json:"start"`
		End         date       `json:"end"`
	}{
		Summary:     event.Name,
		Description: event.Description,
		Attendees:   attendees(event.AttendeeEmails),
		Start:       date{DateTime: event.Start.UTC().Format(time.RFC3339)},
		End:         date{DateTime: event.End.UTC().Format(time.RFC3339)},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, eventURL(""), bytes.NewReader(encoded))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return "", fmt.Errorf("Google Calendar returned %s: %s", response.Status, body)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Google Calendar event: %w", err)
	}
	if result.ID == "" {
		return "", fmt.Errorf("Google Calendar event response did not include an ID")
	}
	return result.ID, nil
}

func UpdateGoogleEvent(ctx context.Context, client *http.Client, eventID string, event Event) error {
	payload := struct {
		Summary     string     `json:"summary"`
		Description string     `json:"description"`
		Attendees   []attendee `json:"attendees"`
	}{Summary: event.Name, Description: event.Description, Attendees: attendees(event.AttendeeEmails)}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, eventURL(eventID), bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("Google Calendar returned %s: %s", response.Status, body)
	}
	return nil
}

type date struct {
	DateTime string `json:"dateTime"`
}

type attendee struct {
	Email string `json:"email"`
}

func attendees(emails []string) []attendee {
	values := make([]attendee, len(emails))
	for index, email := range emails {
		values[index] = attendee{Email: email}
	}
	return values
}

func eventURL(eventID string) string {
	value := googleEventsURL
	if eventID != "" {
		value += "/" + url.PathEscape(eventID)
	}
	return value + "?sendUpdates=all"
}
