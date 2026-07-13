package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

const (
	webhookInviteeCreated  = "invitee.created"
	webhookInviteeCanceled = "invitee.canceled"
	webhookRoutingCreated  = "routing_form_submission.created"
)

// TODO: Emit routing form submission webhooks when routing forms are implemented.

type createWebhookSubscriptionRequest struct {
	URL          string   `json:"url"`
	Events       []string `json:"events"`
	Organization string   `json:"organization"`
	User         string   `json:"user"`
	Scope        string   `json:"scope"`
	SigningKey   string   `json:"signing_key"`
}

type compatibilityWebhookSubscription struct {
	URI          string    `json:"uri"`
	CallbackURL  string    `json:"callback_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	State        string    `json:"state"`
	Events       []string  `json:"events"`
	Scope        string    `json:"scope"`
	Organization string    `json:"organization"`
	User         *string   `json:"user"`
	Creator      string    `json:"creator"`
	// TODO: Add retry_started_at when retry state is exposed as subscription metadata.
}

func (s *Server) createWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	var request createWebhookSubscriptionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return
	}
	callback, err := url.Parse(strings.TrimSpace(request.URL))
	if err != nil || (callback.Scheme != "http" && callback.Scheme != "https") || callback.Host == "" {
		writeError(w, http.StatusBadRequest, "url must be a full HTTP or HTTPS URL")
		return
	}
	if request.Organization != s.organizationURI() {
		writeError(w, http.StatusNotFound, "organization not found")
		return
	}
	if request.Scope != "organization" && request.Scope != "user" {
		writeError(w, http.StatusBadRequest, "scope must be organization or user")
		return
	}
	var userEmail string
	if request.Scope == "user" {
		user, err := s.userForURI(request.User)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			internalError(w, err, "load webhook user")
			return
		}
		userEmail = user.Email
	} else if request.User != "" {
		writeError(w, http.StatusBadRequest, "user must be omitted for organization scope")
		return
	}
	events, err := normalizeWebhookEvents(request.Events, request.Scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var encryptedSigningKey string
	if request.SigningKey != "" {
		if utf8.RuneCountInString(request.SigningKey) < 16 {
			writeError(w, http.StatusBadRequest, "signing_key must be at least 16 characters")
			return
		}
		encryptedSigningKey, err = s.tokenCipher.Encrypt([]byte(request.SigningKey))
		if err != nil {
			internalError(w, err, "encrypt webhook signing key")
			return
		}
	}
	id, err := security.RandomToken(12)
	if err != nil {
		internalError(w, err, "generate webhook subscription ID")
		return
	}
	now := s.now().UTC().Truncate(time.Second)
	subscription := model.WebhookSubscription{
		ID: id, CallbackURL: callback.String(), Events: events, Scope: request.Scope,
		UserEmail: userEmail, CreatorEmail: userFromRequest(r).Email,
		EncryptedSigningKey: encryptedSigningKey, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreateWebhookSubscription(subscription); errors.Is(err, store.ErrExists) {
		writeError(w, http.StatusConflict, "a webhook subscription with this URL already exists")
		return
	} else if err != nil {
		internalError(w, err, "create webhook subscription")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"resource": s.compatibilityWebhookSubscription(subscription)})
}

func (s *Server) listWebhookSubscriptions(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	if values.Get("organization") != s.organizationURI() {
		writeError(w, http.StatusNotFound, "organization not found")
		return
	}
	scope := values.Get("scope")
	if scope != "organization" && scope != "user" {
		writeError(w, http.StatusBadRequest, "scope must be organization or user")
		return
	}
	var userEmail string
	if scope == "user" {
		user, err := s.userForURI(values.Get("user"))
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			internalError(w, err, "load webhook list user")
			return
		}
		userEmail = user.Email
	}
	subscriptions, err := s.store.ListWebhookSubscriptions()
	if err != nil {
		internalError(w, err, "list webhook subscriptions")
		return
	}
	items := make([]compatibilityWebhookSubscription, 0)
	for _, subscription := range subscriptions {
		if subscription.Scope != scope || (scope == "user" && !strings.EqualFold(subscription.UserEmail, userEmail)) {
			continue
		}
		items = append(items, s.compatibilityWebhookSubscription(subscription))
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

func normalizeWebhookEvents(values []string, scope string) ([]string, error) {
	if len(values) == 0 {
		return nil, errors.New("events must contain at least one event")
	}
	seen := make(map[string]bool)
	events := make([]string, 0, len(values))
	for _, value := range values {
		if value != webhookInviteeCreated && value != webhookInviteeCanceled && value != webhookRoutingCreated {
			return nil, errors.New("events contains an unsupported event")
		}
		if value == webhookRoutingCreated && scope != "organization" {
			return nil, errors.New("routing_form_submission.created requires organization scope")
		}
		if seen[value] {
			return nil, errors.New("events must not contain duplicates")
		}
		seen[value] = true
		events = append(events, value)
	}
	return events, nil
}

func (s *Server) compatibilityWebhookSubscription(subscription model.WebhookSubscription) compatibilityWebhookSubscription {
	var userURI *string
	if subscription.UserEmail != "" {
		value := s.userURI(subscription.UserEmail)
		userURI = &value
	}
	return compatibilityWebhookSubscription{
		URI: s.webhookSubscriptionURI(subscription.ID), CallbackURL: subscription.CallbackURL,
		CreatedAt: subscription.CreatedAt, UpdatedAt: subscription.UpdatedAt, State: "active",
		Events: subscription.Events, Scope: subscription.Scope, Organization: s.organizationURI(),
		User: userURI, Creator: s.userURI(subscription.CreatorEmail),
	}
}

type webhookInviteePayload struct {
	compatibilityInvitee
	ScheduledEvent compatibilityScheduledEvent `json:"scheduled_event"`
}

type webhookEvent struct {
	Event     string                `json:"event"`
	CreatedAt time.Time             `json:"created_at"`
	CreatedBy string                `json:"created_by"`
	Payload   webhookInviteePayload `json:"payload"`
}

func (s *Server) queueWebhookEvent(eventName string, eventType model.EventType, booking model.Booking) error {
	subscriptions, err := s.store.ListWebhookSubscriptions()
	if err != nil {
		return err
	}
	matching := make([]model.WebhookSubscription, 0)
	for _, subscription := range subscriptions {
		if !containsString(subscription.Events, eventName) {
			continue
		}
		if subscription.Scope == "user" && !containsEmail(booking.RecipientEmails, subscription.UserEmail) {
			continue
		}
		matching = append(matching, subscription)
	}
	if len(matching) == 0 {
		return nil
	}
	groups, err := s.scheduledEventGroups()
	if err != nil {
		return err
	}
	var group scheduledEventGroup
	for _, candidate := range groups {
		if candidate.EventSlug == booking.EventSlug && candidate.Start.Equal(booking.Time) && candidate.End.Equal(booking.EndTime) {
			group = candidate
			break
		}
	}
	invitee, err := s.compatibilityInvitee(group, booking)
	if err != nil {
		return err
	}
	scheduledEvent, err := s.compatibilityScheduledEvent(group)
	if err != nil {
		return err
	}
	createdAt := booking.CreatedAt
	if booking.CanceledAt != nil {
		createdAt = *booking.CanceledAt
	}
	payload, err := json.Marshal(webhookEvent{
		Event: eventName, CreatedAt: createdAt, CreatedBy: s.userURI(eventType.CreatedBy),
		Payload: webhookInviteePayload{compatibilityInvitee: invitee, ScheduledEvent: scheduledEvent},
	})
	if err != nil {
		return err
	}
	now := s.now().UTC().Truncate(time.Second)
	deliveries := make([]model.WebhookDelivery, 0, len(matching))
	for _, subscription := range matching {
		id, err := security.RandomToken(18)
		if err != nil {
			return err
		}
		deliveries = append(deliveries, model.WebhookDelivery{
			ID: id, SubscriptionID: subscription.ID, Payload: string(payload),
			NextAttemptAt: now, CreatedAt: now,
		})
	}
	if err := s.store.CreateWebhookDeliveries(deliveries); err != nil {
		return err
	}
	select {
	case s.webhookWake <- struct{}{}:
	default:
	}
	return nil
}

func (s *Server) RunWebhookDelivery(ctx context.Context) {
	for {
		if err := s.deliverDueWebhooks(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("deliver webhooks", "error", err)
		}
		timer := time.NewTimer(time.Minute)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.webhookWake:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (s *Server) deliverDueWebhooks(ctx context.Context) error {
	deliveries, err := s.store.ListDueWebhookDeliveries(s.now().UTC())
	if err != nil {
		return err
	}
	var wait sync.WaitGroup
	errorsChannel := make(chan error, len(deliveries))
	for _, delivery := range deliveries {
		delivery := delivery
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsChannel <- s.deliverWebhook(ctx, delivery)
		}()
	}
	wait.Wait()
	close(errorsChannel)
	var deliveryErrors []error
	for err := range errorsChannel {
		if err != nil {
			deliveryErrors = append(deliveryErrors, err)
		}
	}
	return errors.Join(deliveryErrors...)
}

func (s *Server) deliverWebhook(ctx context.Context, delivery model.WebhookDelivery) error {
	subscription, err := s.store.GetWebhookSubscription(delivery.SubscriptionID)
	if errors.Is(err, store.ErrNotFound) {
		return s.store.DeleteWebhookDelivery(delivery.ID)
	}
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, subscription.CallbackURL, bytes.NewBufferString(delivery.Payload))
	if err != nil {
		return s.retryWebhook(delivery, err)
	}
	request.Header.Set("Content-Type", "application/json")
	if subscription.EncryptedSigningKey != "" {
		key, err := s.tokenCipher.Decrypt(subscription.EncryptedSigningKey)
		if err != nil {
			return s.retryWebhook(delivery, err)
		}
		timestamp := strconv.FormatInt(s.now().Unix(), 10)
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write([]byte(timestamp + "." + delivery.Payload))
		request.Header.Set("Webhook-Signature", "t="+timestamp+",v1="+hex.EncodeToString(mac.Sum(nil)))
	}
	response, err := s.webhookHTTP.Do(request)
	if err != nil {
		return s.retryWebhook(delivery, err)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return s.store.DeleteWebhookDelivery(delivery.ID)
	}
	return s.retryWebhook(delivery, fmt.Errorf("receiver returned status %d", response.StatusCode))
}

func (s *Server) retryWebhook(delivery model.WebhookDelivery, deliveryErr error) error {
	delivery.Attempts++
	delivery.NextAttemptAt = s.now().UTC().Truncate(time.Second).Add(webhookRetryDelay(delivery.Attempts))
	if err := s.store.PutWebhookDelivery(delivery); err != nil {
		return errors.Join(deliveryErr, err)
	}
	slog.Warn("webhook delivery scheduled for retry", "delivery", delivery.ID, "attempts", delivery.Attempts, "error", deliveryErr)
	return nil
}

func webhookRetryDelay(attempt int) time.Duration {
	delay := time.Minute
	for index := 1; index < attempt && delay < 24*time.Hour; index++ {
		delay = min(delay*2, 24*time.Hour)
	}
	return delay
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
