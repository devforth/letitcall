package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type BusyPeriod struct {
	EventID string
	Start   time.Time
	End     time.Time
}

func ListGoogleBusy(ctx context.Context, client *http.Client, start, end time.Time) ([]BusyPeriod, error) {
	periods := make([]BusyPeriod, 0)
	pageToken := ""
	for {
		endpoint, err := url.Parse(googleEventsURL)
		if err != nil {
			return nil, err
		}
		query := endpoint.Query()
		query.Set("timeMin", start.UTC().Format(time.RFC3339))
		query.Set("timeMax", end.UTC().Format(time.RFC3339))
		query.Set("singleEvents", "true")
		query.Set("maxResults", "2500")
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		endpoint.RawQuery = query.Encode()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, err
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			response.Body.Close()
			return nil, fmt.Errorf("Google Calendar returned %s: %s", response.Status, body)
		}
		var result struct {
			TimeZone      string `json:"timeZone"`
			NextPageToken string `json:"nextPageToken"`
			Items         []struct {
				ID           string          `json:"id"`
				Status       string          `json:"status"`
				Transparency string          `json:"transparency"`
				Start        googleEventDate `json:"start"`
				End          googleEventDate `json:"end"`
			} `json:"items"`
		}
		if err := json.NewDecoder(io.LimitReader(response.Body, 10<<20)).Decode(&result); err != nil {
			response.Body.Close()
			return nil, fmt.Errorf("decode Google Calendar events: %w", err)
		}
		response.Body.Close()
		location, err := calendarLocation(result.TimeZone)
		if err != nil {
			return nil, err
		}
		for _, event := range result.Items {
			if event.Status == "cancelled" || event.Transparency == "transparent" {
				continue
			}
			eventStart, err := event.Start.instant(location)
			if err != nil {
				return nil, err
			}
			eventEnd, err := event.End.instant(location)
			if err != nil {
				return nil, err
			}
			periods = append(periods, BusyPeriod{EventID: event.ID, Start: eventStart, End: eventEnd})
		}
		if result.NextPageToken == "" {
			return periods, nil
		}
		pageToken = result.NextPageToken
	}
}

type googleEventDate struct {
	Date     string `json:"date"`
	DateTime string `json:"dateTime"`
}

func (d googleEventDate) instant(location *time.Location) (time.Time, error) {
	if d.DateTime != "" {
		value, err := time.Parse(time.RFC3339, d.DateTime)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse Google Calendar dateTime: %w", err)
		}
		return value.UTC().Truncate(time.Second), nil
	}
	value, err := time.ParseInLocation(time.DateOnly, d.Date, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse Google Calendar date: %w", err)
	}
	return value.UTC().Truncate(time.Second), nil
}

func calendarLocation(name string) (*time.Location, error) {
	if name == "" {
		return time.UTC, nil
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("load Google Calendar timezone: %w", err)
	}
	return location, nil
}
