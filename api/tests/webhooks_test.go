package tests

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/httpapi"
	"github.com/letitcall/letitcall/api/internal/model"
)

type receivedWebhook struct {
	body      string
	signature string
}

func createWebhook(t *testing.T, f *fixture, token string, body map[string]any, status int) map[string]any {
	t.Helper()
	response := expectStatus(t, f.bearerRequest(http.MethodPost, "/webhook_subscriptions", token, body), status)
	if status != http.StatusCreated {
		requireErrorObject(t, response)
		return nil
	}
	return nestedObject(t, decodeObject(t, response), "resource")
}

func waitWebhook(t *testing.T, received <-chan receivedWebhook) receivedWebhook {
	t.Helper()
	select {
	case webhook := <-received:
		return webhook
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for webhook delivery")
		return receivedWebhook{}
	}
}

func verifyWebhookSignature(t *testing.T, webhook receivedWebhook, signingKey string) {
	t.Helper()
	parts := strings.Split(webhook.signature, ",")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "t=") || !strings.HasPrefix(parts[1], "v1=") {
		t.Fatalf("invalid Webhook-Signature header %q", webhook.signature)
	}
	timestamp := strings.TrimPrefix(parts[0], "t=")
	if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
		t.Fatalf("invalid signature timestamp %q", timestamp)
	}
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = mac.Write([]byte(timestamp + "." + webhook.body))
	want := hex.EncodeToString(mac.Sum(nil))
	if got := strings.TrimPrefix(parts[1], "v1="); !hmac.Equal([]byte(got), []byte(want)) {
		t.Fatalf("signature digest = %q, want %q", got, want)
	}
}

func startWebhookWorker(server *httpapi.Server) (context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		server.RunWebhookDelivery(ctx)
	}()
	return cancel, done
}

func waitForDeliveryState(t *testing.T, f *fixture, predicate func([]model.WebhookDelivery) bool) []model.WebhookDelivery {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		deliveries, err := f.store.ListWebhookDeliveries()
		if err != nil {
			t.Fatal(err)
		}
		if predicate(deliveries) {
			return deliveries
		}
		if time.Now().After(deadline) {
			t.Fatalf("webhook deliveries did not reach expected state: %#v", deliveries)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWebhookSubscriptionsValidatePersistAndRespectUserScope(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC",
	}), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Scoped Demo", []string{adminEmail}, 3)), http.StatusCreated)
	adminToken := createTestToken(t, f, "Admin webhooks")
	admin := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", adminToken.Token, nil), http.StatusOK)), "resource")

	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	expectStatus(t, f.login("member@example.com", "MemberPassword123!"), http.StatusOK)
	memberToken := createTestToken(t, f, "Member webhooks")
	member := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", memberToken.Token, nil), http.StatusOK)), "resource")

	adminReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer adminReceiver.Close()
	memberReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer memberReceiver.Close()
	organization := admin["current_organization"].(string)

	createWebhook(t, f, adminToken.Token, map[string]any{
		"url": adminReceiver.URL, "events": []string{"invitee.created", "invitee.created"},
		"organization": organization, "user": admin["uri"], "scope": "user",
	}, http.StatusBadRequest)
	createWebhook(t, f, adminToken.Token, map[string]any{
		"url": adminReceiver.URL, "events": []string{"routing_form_submission.created"},
		"organization": organization, "user": admin["uri"], "scope": "user",
	}, http.StatusBadRequest)
	createWebhook(t, f, adminToken.Token, map[string]any{
		"url": "ftp://receiver.example/hook", "events": []string{"invitee.created"},
		"organization": organization, "user": admin["uri"], "scope": "user",
	}, http.StatusBadRequest)

	adminSubscription := createWebhook(t, f, adminToken.Token, map[string]any{
		"url": adminReceiver.URL, "events": []string{"invitee.created", "invitee.canceled"},
		"organization": organization, "user": admin["uri"], "scope": "user", "signing_key": "0123456789abcdef",
	}, http.StatusCreated)
	if adminSubscription["user"] != admin["uri"] || adminSubscription["creator"] != admin["uri"] {
		t.Fatalf("unexpected user-scoped subscription: %#v", adminSubscription)
	}
	if _, exists := adminSubscription["signing_key"]; exists {
		t.Fatal("webhook signing key was returned")
	}
	createWebhook(t, f, adminToken.Token, map[string]any{
		"url": adminReceiver.URL, "events": []string{"invitee.created"},
		"organization": organization, "user": admin["uri"], "scope": "user",
	}, http.StatusConflict)
	createWebhook(t, f, memberToken.Token, map[string]any{
		"url": memberReceiver.URL, "events": []string{"invitee.created"},
		"organization": organization, "user": member["uri"], "scope": "user",
	}, http.StatusCreated)

	subscriptions, err := f.store.ListWebhookSubscriptions()
	if err != nil || len(subscriptions) != 2 {
		t.Fatalf("webhook subscriptions were not persisted: subscriptions=%#v err=%v", subscriptions, err)
	}
	for _, subscription := range subscriptions {
		if subscription.CallbackURL == adminReceiver.URL {
			if subscription.EncryptedSigningKey == "" || subscription.EncryptedSigningKey == "0123456789abcdef" {
				t.Fatal("webhook signing key was not encrypted at rest")
			}
			encoded, _ := json.Marshal(subscription)
			if strings.Contains(string(encoded), "0123456789abcdef") {
				t.Fatal("stored webhook subscription contains the plaintext signing key")
			}
		}
	}

	listQuery := url.Values{"organization": {organization}, "scope": {"user"}, "user": {admin["uri"].(string)}}
	listed := objectCollection(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/webhook_subscriptions?"+listQuery.Encode(), adminToken.Token, nil), http.StatusOK)))
	if len(listed) != 1 || listed[0]["callback_url"] != adminReceiver.URL {
		t.Fatalf("webhook subscription user filtering failed: %#v", listed)
	}

	date := time.Now().UTC().AddDate(0, 0, 2)
	createPublicBooking(t, f, "scoped-demo", time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, time.UTC), "Scoped Lead", "lead@example.com", nil, "")
	deliveries, err := f.store.ListWebhookDeliveries()
	if err != nil || len(deliveries) != 1 {
		t.Fatalf("user scope queued %d deliveries, want only the hosting user: err=%v", len(deliveries), err)
	}
	deliverySubscription, err := f.store.GetWebhookSubscription(deliveries[0].SubscriptionID)
	if err != nil || deliverySubscription.CallbackURL != adminReceiver.URL {
		t.Fatalf("delivery was queued for the wrong user scope: subscription=%#v err=%v", deliverySubscription, err)
	}
}

func TestWebhookCreationCancellationRawBodySignatureAndCompletion(t *testing.T) {
	received := make(chan receivedWebhook, 2)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		received <- receivedWebhook{body: string(body), signature: r.Header.Get("Webhook-Signature")}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer receiver.Close()

	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Webhook Demo", []string{adminEmail}, 2)), http.StatusCreated)
	token := createTestToken(t, f, "Webhook receiver")
	current := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", token.Token, nil), http.StatusOK)), "resource")
	signingKey := "a-strong-signing-key"
	createdSubscription := createWebhook(t, f, token.Token, map[string]any{
		"url": receiver.URL, "events": []string{"invitee.created", "invitee.canceled", "routing_form_submission.created"},
		"organization": current["current_organization"], "scope": "organization", "signing_key": signingKey,
	}, http.StatusCreated)
	if createdSubscription["scope"] != "organization" || createdSubscription["user"] != nil {
		t.Fatalf("unexpected organization subscription: %#v", createdSubscription)
	}

	date := time.Now().UTC().AddDate(0, 0, 2)
	booking := createPublicBooking(t, f, "webhook-demo", time.Date(date.Year(), date.Month(), date.Day(), 11, 0, 0, 0, time.UTC), "Alex Lead", "alex@example.com", nil, "Qualification notes")
	deliveries, err := f.store.ListWebhookDeliveries()
	if err != nil || len(deliveries) != 1 {
		t.Fatalf("booking lifecycle did not persist its webhook before success: deliveries=%#v err=%v", deliveries, err)
	}
	storedRawBody := deliveries[0].Payload

	cancelWorker, workerDone := startWebhookWorker(f.api)
	defer func() {
		cancelWorker()
		<-workerDone
	}()
	created := waitWebhook(t, received)
	if created.body != storedRawBody {
		t.Fatalf("webhook attempt body changed from persisted raw JSON:\nwant %s\n got %s", storedRawBody, created.body)
	}
	verifyWebhookSignature(t, created, signingKey)
	createdEvent := decodeObject(t, []byte(created.body))
	if createdEvent["event"] != "invitee.created" {
		t.Fatalf("unexpected created webhook: %#v", createdEvent)
	}
	createdPayload := nestedObject(t, createdEvent, "payload")
	if createdPayload["email"] != "alex@example.com" || createdPayload["status"] != "active" || createdPayload["scheduled_event"] == nil {
		t.Fatalf("unexpected created webhook payload: %#v", createdPayload)
	}
	waitForDeliveryState(t, f, func(deliveries []model.WebhookDelivery) bool { return len(deliveries) == 0 })

	cancelBooking(t, f, booking.ManageURL, "Lead canceled")
	canceled := waitWebhook(t, received)
	verifyWebhookSignature(t, canceled, signingKey)
	canceledEvent := decodeObject(t, []byte(canceled.body))
	if canceledEvent["event"] != "invitee.canceled" {
		t.Fatalf("unexpected canceled webhook: %#v", canceledEvent)
	}
	canceledPayload := nestedObject(t, canceledEvent, "payload")
	if canceledPayload["status"] != "canceled" || canceledPayload["cancellation"] == nil {
		t.Fatalf("unexpected canceled webhook payload: %#v", canceledPayload)
	}
	waitForDeliveryState(t, f, func(deliveries []model.WebhookDelivery) bool { return len(deliveries) == 0 })
}

func TestWebhookRetryPersistsAndRecoversAfterRestart(t *testing.T) {
	var failing atomic.Bool
	failing.Store(true)
	received := make(chan receivedWebhook, 4)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		received <- receivedWebhook{body: string(body), signature: r.Header.Get("Webhook-Signature")}
		if failing.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer receiver.Close()

	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Retry Demo", []string{adminEmail}, 1)), http.StatusCreated)
	token := createTestToken(t, f, "Retry receiver")
	current := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", token.Token, nil), http.StatusOK)), "resource")
	createWebhook(t, f, token.Token, map[string]any{
		"url": receiver.URL, "events": []string{"invitee.created"},
		"organization": current["current_organization"], "scope": "organization",
	}, http.StatusCreated)

	date := time.Now().UTC().AddDate(0, 0, 2)
	createPublicBooking(t, f, "retry-demo", time.Date(date.Year(), date.Month(), date.Day(), 14, 0, 0, 0, time.UTC), "Retry Lead", "retry@example.com", nil, "")
	cancelFirstWorker, firstWorkerDone := startWebhookWorker(f.api)
	firstAttempt := waitWebhook(t, received)
	deliveries := waitForDeliveryState(t, f, func(deliveries []model.WebhookDelivery) bool {
		return len(deliveries) == 1 && deliveries[0].Attempts == 1
	})
	if deliveries[0].NextAttemptAt.Before(time.Now().UTC().Add(50 * time.Second)) {
		t.Fatalf("first retry was not delayed by about one minute: %#v", deliveries[0])
	}
	cancelFirstWorker()
	<-firstWorkerDone

	delivery := deliveries[0]
	delivery.NextAttemptAt = time.Now().UTC().Truncate(time.Second).Add(-time.Second)
	if err := f.store.PutWebhookDelivery(delivery); err != nil {
		t.Fatal(err)
	}
	failing.Store(false)
	cfg := testConfig(f.dataPath)
	cfg.HTTP.BaseURL = "http://example.test"
	restarted, err := httpapi.New(cfg, f.store)
	if err != nil {
		t.Fatal(err)
	}
	cancelRestartedWorker, restartedWorkerDone := startWebhookWorker(restarted)
	secondAttempt := waitWebhook(t, received)
	if secondAttempt.body != firstAttempt.body {
		t.Fatalf("retried webhook body changed across restart:\nfirst %s\nsecond %s", firstAttempt.body, secondAttempt.body)
	}
	waitForDeliveryState(t, f, func(deliveries []model.WebhookDelivery) bool { return len(deliveries) == 0 })
	cancelRestartedWorker()
	<-restartedWorkerDone
}
