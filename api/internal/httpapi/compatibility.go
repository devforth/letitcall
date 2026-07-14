package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/store"
)

const (
	compatibilityAPIPath = "/api/v1"
	defaultPageCount     = 20
)

type compatibilityPagination struct {
	Count             int     `json:"count"`
	NextPage          *string `json:"next_page"`
	PreviousPage      *string `json:"previous_page"`
	NextPageToken     *string `json:"next_page_token"`
	PreviousPageToken *string `json:"previous_page_token"`
}

func stableCompatibilityID(value string) string {
	digest := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func (s *Server) compatibilityBaseURL() string {
	return s.cfg.HTTP.BaseURL + compatibilityAPIPath
}

func (s *Server) organizationURI() string {
	return s.compatibilityBaseURL() + "/organizations/default"
}

func (s *Server) userURI(email string) string {
	return s.compatibilityBaseURL() + "/users/" + stableCompatibilityID(strings.ToLower(email))
}

func (s *Server) eventTypeURI(slug string) string {
	return s.compatibilityBaseURL() + "/event_types/" + url.PathEscape(slug)
}

func (s *Server) scheduledEventID(eventSlug string, start, end time.Time) string {
	return stableCompatibilityID(eventSlug + "\x00" + start.Format(time.RFC3339) + "\x00" + end.Format(time.RFC3339))
}

func (s *Server) scheduledEventURI(eventSlug string, start, end time.Time) string {
	return s.compatibilityBaseURL() + "/scheduled_events/" + s.scheduledEventID(eventSlug, start, end)
}

func (s *Server) inviteeURI(eventSlug string, start, end time.Time, bookingID string) string {
	return s.scheduledEventURI(eventSlug, start, end) + "/invitees/" + url.PathEscape(bookingID)
}

func (s *Server) webhookSubscriptionURI(id string) string {
	return s.compatibilityBaseURL() + "/webhook_subscriptions/" + url.PathEscape(id)
}

func (s *Server) userForURI(uri string) (model.User, error) {
	users, err := s.store.ListUsers()
	if err != nil {
		return model.User{}, err
	}
	for _, user := range users {
		if s.userURI(user.Email) == uri {
			return user, nil
		}
	}
	return model.User{}, store.ErrNotFound
}

func (s *Server) eventSlugForURI(uri string) (string, error) {
	prefix := s.compatibilityBaseURL() + "/event_types/"
	if !strings.HasPrefix(uri, prefix) {
		return "", store.ErrNotFound
	}
	slug, err := url.PathUnescape(strings.TrimPrefix(uri, prefix))
	if err != nil || slug == "" || strings.Contains(slug, "/") {
		return "", store.ErrNotFound
	}
	return slug, nil
}

func parseCompatibilityInstant(value, name string) (time.Time, error) {
	if strings.Contains(value, ".") || !strings.HasSuffix(value, "Z") {
		return time.Time{}, fmt.Errorf("%s must use UTC RFC3339 seconds", name)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.UTC().Format(time.RFC3339) != value {
		return time.Time{}, fmt.Errorf("%s must use UTC RFC3339 seconds", name)
	}
	return parsed.UTC(), nil
}

func parseOptionalCompatibilityInstant(values url.Values, name string) (*time.Time, error) {
	value := values.Get(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseCompatibilityInstant(value, name)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func paginationRequest(values url.Values) (int, int, error) {
	count := defaultPageCount
	if value := values.Get("count"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, 0, errors.New("count must be between 1 and 100")
		}
		count = parsed
	}
	offset := 0
	if token := values.Get("page_token"); token != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			return 0, 0, errors.New("page_token is invalid")
		}
		parsed, err := strconv.Atoi(string(decoded))
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("page_token is invalid")
		}
		offset = parsed
	}
	return count, offset, nil
}

func pageToken(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func compatibilityPage[T any](r *http.Request, items []T, count, offset int, baseURL string) ([]T, compatibilityPagination) {
	if offset > len(items) {
		offset = len(items)
	}
	end := min(len(items), offset+count)
	page := items[offset:end]
	pagination := compatibilityPagination{Count: len(page)}
	if end < len(items) {
		token := pageToken(end)
		link := compatibilityPageURL(r, baseURL, token)
		pagination.NextPageToken = &token
		pagination.NextPage = &link
	}
	if offset > 0 {
		previousOffset := max(0, offset-count)
		token := pageToken(previousOffset)
		link := compatibilityPageURL(r, baseURL, token)
		pagination.PreviousPageToken = &token
		pagination.PreviousPage = &link
	}
	return page, pagination
}

func compatibilityPageURL(r *http.Request, baseURL, token string) string {
	values := r.URL.Query()
	values.Set("page_token", token)
	return baseURL + strings.TrimPrefix(r.URL.Path, compatibilityAPIPath) + "?" + values.Encode()
}

func containsEmail(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
