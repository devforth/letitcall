package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const googleEventsURL = "https://www.googleapis.com/calendar/v3/calendars/primary/events"

type Event struct {
	Name          string
	AttendeeEmail string
	Start         time.Time
	End           time.Time
}

func AddGoogleEvent(ctx context.Context, client *http.Client, event Event) error {
	payload := struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Start       date   `json:"start"`
		End         date   `json:"end"`
	}{
		Summary:     event.Name,
		Description: "Booked by " + event.AttendeeEmail,
		Start:       date{DateTime: event.Start.UTC().Format(time.RFC3339)},
		End:         date{DateTime: event.End.UTC().Format(time.RFC3339)},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, googleEventsURL, bytes.NewReader(encoded))
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
