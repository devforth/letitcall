package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
)

const compatibilityTestPath = "/api/v1"

type testTokenResponse struct {
	APIToken struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	} `json:"apiToken"`
	Token string `json:"token"`
}

func (f *fixture) bearerRequest(method, path, token string, body any) *http.Response {
	f.t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			f.t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	requestPath := strings.TrimPrefix(path, f.basePath)
	if !strings.HasPrefix(requestPath, compatibilityTestPath+"/") {
		requestPath = compatibilityTestPath + requestPath
	}
	request, err := http.NewRequest(method, f.server.URL+f.basePath+requestPath, reader)
	if err != nil {
		f.t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return mustDo(f.t, f.client, request)
}

func createTestToken(t *testing.T, f *fixture, name string) testTokenResponse {
	t.Helper()
	body := expectStatus(t, f.request(http.MethodPost, "/api/integration/tokens", map[string]string{"name": name}), http.StatusCreated)
	var response testTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Token == "" || response.APIToken.ID == "" {
		t.Fatalf("token creation response is incomplete: %s", body)
	}
	return response
}

func decodeObject(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func nestedObject(t *testing.T, value map[string]any, key string) map[string]any {
	t.Helper()
	object, ok := value[key].(map[string]any)
	if !ok {
		t.Fatalf("%s is not an object in %#v", key, value)
	}
	return object
}

func objectCollection(t *testing.T, value map[string]any) []map[string]any {
	t.Helper()
	raw, ok := value["collection"].([]any)
	if !ok {
		t.Fatalf("collection is not an array in %#v", value)
	}
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		object, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("collection item is not an object: %#v", item)
		}
		items = append(items, object)
	}
	return items
}

func requireErrorObject(t *testing.T, body []byte) {
	t.Helper()
	value := decodeObject(t, body)
	if len(value) != 1 || value["error"] == "" {
		t.Fatalf("error response must contain only an error string: %s", body)
	}
}

func requireUTCSecond(t *testing.T, raw any) {
	t.Helper()
	value, ok := raw.(string)
	if !ok || strings.Contains(value, ".") || !strings.HasSuffix(value, "Z") {
		t.Fatalf("instant is not strict UTC RFC3339 seconds: %#v", raw)
	}
	if parsed, err := time.Parse(time.RFC3339, value); err != nil || parsed.Format(time.RFC3339) != value {
		t.Fatalf("instant is not strict UTC RFC3339 seconds: %q", value)
	}
}

func TestAPIIntegrationTokensAreOneTimeOwnedAndRevocable(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/team")
	requireErrorObject(t, expectStatus(t, f.request(http.MethodGet, "/api/integration", nil), http.StatusUnauthorized))
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	unicodeToken := createTestToken(t, f, strings.Repeat("é", 100))
	expectStatus(t, f.request(http.MethodDelete, "/api/integration/tokens/"+unicodeToken.APIToken.ID, nil), http.StatusNoContent)
	requireErrorObject(t, expectStatus(t, f.request(http.MethodPost, "/api/integration/tokens", map[string]string{"name": strings.Repeat("é", 101)}), http.StatusBadRequest))

	adminToken := createTestToken(t, f, "  Lead connector  ")
	if adminToken.APIToken.Name != "Lead connector" || !strings.HasPrefix(adminToken.Token, "lic_") {
		t.Fatalf("unexpected generated token: %#v", adminToken)
	}
	digest := security.TokenDigest(adminToken.Token)
	if adminToken.APIToken.ID != digest {
		t.Fatalf("token metadata ID = %q, want digest %q", adminToken.APIToken.ID, digest)
	}
	stored, err := f.store.GetAPIToken(digest)
	if err != nil || stored.UserEmail != adminEmail {
		t.Fatalf("token digest was not persisted: token=%#v err=%v", stored, err)
	}
	encodedStored, _ := json.Marshal(stored)
	if bytes.Contains(encodedStored, []byte(adminToken.Token)) {
		t.Fatal("stored token metadata contains the secret")
	}

	integrationBody := expectStatus(t, f.request(http.MethodGet, "/api/integration", nil), http.StatusOK)
	if !bytes.Contains(integrationBody, []byte(`"baseURL":"http://example.test/team/api/v1"`)) ||
		!bytes.Contains(integrationBody, []byte(`"swaggerURL":"http://example.test/team/api/v1/swagger/"`)) ||
		bytes.Contains(integrationBody, []byte(adminToken.Token)) {
		t.Fatalf("unexpected integration response: %s", integrationBody)
	}
	userBody := expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", adminToken.Token, nil), http.StatusOK)
	userResource := nestedObject(t, decodeObject(t, userBody), "resource")
	if userResource["email"] != adminEmail || userResource["current_organization"] != "http://example.test/team/api/v1/organizations/default" {
		t.Fatalf("unexpected current user mapping: %#v", userResource)
	}
	if !strings.HasPrefix(userResource["uri"].(string), "http://example.test/team/api/v1/users/") {
		t.Fatalf("user URI does not use configured base URL: %#v", userResource["uri"])
	}
	requireUTCSecond(t, userResource["created_at"])
	if _, exists := userResource["slug"]; exists {
		t.Fatal("unsupported user slug was serialized")
	}

	expectStatus(t, f.request(http.MethodPost, "/api/users", map[string]string{
		"email": "member@example.com", "password": "MemberPassword123!", "timezone": "UTC",
	}), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	expectStatus(t, f.login("member@example.com", "MemberPassword123!"), http.StatusOK)
	memberToken := createTestToken(t, f, "Member connector")
	requireErrorObject(t, expectStatus(t, f.request(http.MethodDelete, "/api/integration/tokens/"+adminToken.APIToken.ID, nil), http.StatusNotFound))

	expectStatus(t, f.request(http.MethodPost, "/api/auth/logout", nil), http.StatusNoContent)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodDelete, "/api/users/member@example.com", nil), http.StatusNoContent)
	requireErrorObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", memberToken.Token, nil), http.StatusUnauthorized))
	expectStatus(t, f.request(http.MethodDelete, "/api/integration/tokens/"+adminToken.APIToken.ID, nil), http.StatusNoContent)
	requireErrorObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", adminToken.Token, nil), http.StatusUnauthorized))
}

func TestCompatibilityEventTypesAvailabilityPaginationAndUTC(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/calendar")
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPatch, "/api/users/"+adminEmail, map[string]string{"fullName": "Jane Host"}), http.StatusOK)

	availabilityEvent := eventTypeBody("Availability Demo", []string{adminEmail}, 3)
	for _, day := range availabilityEvent["schedule"].([]map[string]any) {
		day["breaks"] = []map[string]string{{"start": "03:00", "end": "04:00"}}
	}
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", availabilityEvent), http.StatusCreated)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Beta Demo", []string{adminEmail}, 1)), http.StatusCreated)
	token := createTestToken(t, f, "Compatibility tests")

	current := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", token.Token, nil), http.StatusOK)), "resource")
	organization := current["current_organization"].(string)
	query := url.Values{"organization": {organization}, "sort": {"name:asc"}, "count": {"1"}}
	firstPage := decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/event_types?"+query.Encode(), token.Token, nil), http.StatusOK))
	firstItems := objectCollection(t, firstPage)
	if len(firstItems) != 1 || firstItems[0]["name"] != "Availability Demo" {
		t.Fatalf("event type sorting/count failed: %#v", firstPage)
	}
	pagination := nestedObject(t, firstPage, "pagination")
	pageToken, ok := pagination["next_page_token"].(string)
	if !ok || pageToken == "" || !strings.HasPrefix(pagination["next_page"].(string), "http://example.test/calendar/api/v1/event_types?") {
		t.Fatalf("event type pagination did not use configured base URL: %#v", pagination)
	}
	query.Set("page_token", pageToken)
	secondItems := objectCollection(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/event_types?"+query.Encode(), token.Token, nil), http.StatusOK)))
	if len(secondItems) != 1 || secondItems[0]["name"] != "Beta Demo" {
		t.Fatalf("event type page token failed: %#v", secondItems)
	}

	eventType := firstItems[0]
	if eventType["booking_method"] != "instant" || eventType["kind"] != "group" || eventType["scheduling_url"] != "http://example.test/calendar/book/availability-demo" {
		t.Fatalf("unexpected event type mapping: %#v", eventType)
	}
	if _, exists := eventType["location"]; exists {
		t.Fatal("unsupported event type location was serialized")
	}
	requireUTCSecond(t, eventType["created_at"])
	detail := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/event_types/availability-demo", token.Token, nil), http.StatusOK)), "resource")
	if detail["uri"] != eventType["uri"] {
		t.Fatalf("event type detail URI changed: list=%#v detail=%#v", eventType["uri"], detail["uri"])
	}

	date := time.Now().UTC().AddDate(0, 0, 2)
	rangeStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	busyStart := rangeStart.Add(time.Hour)
	if err := f.store.PutGoogleBusy(adminEmail, model.GoogleBusyCache{
		Periods:  []model.GoogleBusyPeriod{{EventID: "external", Start: busyStart, End: busyStart.Add(30 * time.Minute)}},
		SyncedAt: time.Now().UTC().Truncate(time.Second),
	}); err != nil {
		t.Fatal(err)
	}
	bookedStart := rangeStart.Add(2 * time.Hour)
	expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": "availability-demo", "time": bookedStart.Format(time.RFC3339), "attendeeName": "First Lead",
		"attendeeEmail": "first@example.com", "attendeeTimezone": "UTC", "guestEmails": []string{},
	}), http.StatusCreated)
	availabilityQuery := url.Values{
		"event_type": {eventType["uri"].(string)},
		"start_time": {rangeStart.Format(time.RFC3339)},
		"end_time":   {rangeStart.Add(24 * time.Hour).Format(time.RFC3339)},
	}
	availability := objectCollection(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/event_type_available_times?"+availabilityQuery.Encode(), token.Token, nil), http.StatusOK)))
	foundPartiallyBooked := false
	for _, slot := range availability {
		requireUTCSecond(t, slot["start_time"])
		if slot["scheduling_url"] != "http://example.test/calendar/book/availability-demo" {
			t.Fatalf("unexpected scheduling URL: %#v", slot)
		}
		slotTime := slot["start_time"].(string)
		if slotTime == busyStart.Format(time.RFC3339) || slotTime == rangeStart.Add(3*time.Hour).Format(time.RFC3339) {
			t.Fatalf("busy or break slot was returned: %#v", slot)
		}
		if slotTime == bookedStart.Format(time.RFC3339) {
			foundPartiallyBooked = slot["invitees_remaining"] == float64(2)
		}
	}
	if !foundPartiallyBooked {
		t.Fatal("partially booked group slot did not expose remaining capacity")
	}
	availabilityQuery.Set("start_time", strings.TrimSuffix(rangeStart.Format(time.RFC3339), "Z")+".000Z")
	requireErrorObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/event_type_available_times?"+availabilityQuery.Encode(), token.Token, nil), http.StatusBadRequest))
}

func TestCompatibilityScheduledEventsAreGroupedAndInviteesMapped(t *testing.T) {
	f := newFixture(t, false)
	expectStatus(t, f.login(adminEmail, adminPassword), http.StatusOK)
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", eventTypeBody("Group Demo", []string{adminEmail}, 4)), http.StatusCreated)
	soloEvent := eventTypeBody("Solo Demo", []string{adminEmail}, 1)
	soloEvent["inviteeLimit"] = nil
	expectStatus(t, f.request(http.MethodPost, "/api/event-types", soloEvent), http.StatusCreated)
	token := createTestToken(t, f, "Lead import")
	current := nestedObject(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/users/me", token.Token, nil), http.StatusOK)), "resource")

	date := time.Now().UTC().AddDate(0, 0, 3)
	groupTime := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	first := createPublicBooking(t, f, "group-demo", groupTime, "Alex Lead", "alex@example.com", []string{"guest@example.com"}, "Discuss launch")
	second := createPublicBooking(t, f, "group-demo", groupTime, "Sam Buyer", "sam@example.com", nil, "")
	createPublicBooking(t, f, "group-demo", groupTime.Add(time.Hour), "Later Lead", "later@example.com", nil, "")
	createPublicBooking(t, f, "solo-demo", groupTime.Add(2*time.Hour), "Solo Lead", "solo@example.com", nil, "")

	query := url.Values{"organization": {current["current_organization"].(string)}, "sort": {"start_time:desc"}, "count": {"1"}}
	page := decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/scheduled_events?"+query.Encode(), token.Token, nil), http.StatusOK))
	items := objectCollection(t, page)
	if len(items) != 1 || items[0]["start_time"] != groupTime.Add(2*time.Hour).Format(time.RFC3339) {
		t.Fatalf("scheduled event sorting/count failed: %#v", page)
	}
	if nestedObject(t, page, "pagination")["next_page_token"] == nil {
		t.Fatalf("scheduled event pagination token missing: %#v", page)
	}

	query.Set("sort", "start_time:asc")
	query.Set("count", "20")
	query.Del("page_token")
	items = objectCollection(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/scheduled_events?"+query.Encode(), token.Token, nil), http.StatusOK)))
	var grouped map[string]any
	for _, item := range items {
		if item["start_time"] == groupTime.Format(time.RFC3339) {
			grouped = item
		}
	}
	if grouped == nil || grouped["status"] != "active" {
		t.Fatalf("grouped scheduled event missing: %#v", items)
	}
	counter := nestedObject(t, grouped, "invitees_counter")
	if counter["total"] != float64(2) || counter["active"] != float64(2) || counter["limit"] != float64(4) {
		t.Fatalf("unexpected grouped counter: %#v", counter)
	}
	if len(grouped["event_guests"].([]any)) != 1 {
		t.Fatalf("event guests were not mapped: %#v", grouped)
	}
	for _, key := range []string{"location", "calendar_event", "meeting_notes_plain"} {
		if _, exists := grouped[key]; exists {
			t.Fatalf("unsupported scheduled event field %q was serialized", key)
		}
	}
	requireUTCSecond(t, grouped["start_time"])
	foundSolo := false
	for _, item := range items {
		if item["start_time"] == groupTime.Add(2*time.Hour).Format(time.RFC3339) {
			foundSolo = true
			if nestedObject(t, item, "invitees_counter")["limit"] != float64(1) {
				t.Fatalf("one-to-one scheduled event did not map capacity one: %#v", item)
			}
		}
	}
	if !foundSolo {
		t.Fatal("one-to-one scheduled event was not returned")
	}

	groupURI, err := url.Parse(grouped["uri"].(string))
	if err != nil {
		t.Fatal(err)
	}
	inviteePage := decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, groupURI.Path+"/invitees?count=1&sort=created_at:asc", token.Token, nil), http.StatusOK))
	invitees := objectCollection(t, inviteePage)
	if len(invitees) != 1 || nestedObject(t, inviteePage, "pagination")["next_page_token"] == nil {
		t.Fatalf("invitee pagination failed: %#v", inviteePage)
	}
	invitee := invitees[0]
	if invitee["first_name"] != "Alex" || invitee["last_name"] != "Lead" || len(invitee["questions_and_answers"].([]any)) != 1 {
		t.Fatalf("invitee contact mapping failed: %#v", invitee)
	}
	for _, key := range []string{"tracking", "reschedule_url", "routing_form_submission", "payment", "no_show", "reconfirmation"} {
		if _, exists := invitee[key]; exists {
			t.Fatalf("unsupported invitee field %q was serialized", key)
		}
	}

	cancelBooking(t, f, first.ManageURL, "No longer available")
	activeGroup := scheduledEventByStart(t, f, token.Token, current["current_organization"].(string), groupTime)
	if activeGroup["status"] != "active" || nestedObject(t, activeGroup, "invitees_counter")["active"] != float64(1) {
		t.Fatalf("group should remain active after one cancellation: %#v", activeGroup)
	}
	cancelBooking(t, f, second.ManageURL, "Plans changed")
	canceledGroup := scheduledEventByStart(t, f, token.Token, current["current_organization"].(string), groupTime)
	if canceledGroup["status"] != "canceled" || canceledGroup["cancellation"] == nil {
		t.Fatalf("group was not canceled after all primary invitees canceled: %#v", canceledGroup)
	}
	if canceledGroup["uri"] != grouped["uri"] {
		t.Fatal("stable scheduled event URI changed after cancellation")
	}
}

type publicBookingResult struct {
	Booking   model.Booking `json:"booking"`
	ManageURL string        `json:"manageURL"`
}

func createPublicBooking(t *testing.T, f *fixture, slug string, start time.Time, name, email string, guests []string, notes string) publicBookingResult {
	t.Helper()
	body := expectStatus(t, f.request(http.MethodPost, "/api/bookings", map[string]any{
		"eventSlug": slug, "time": start.Format(time.RFC3339), "attendeeName": name,
		"attendeeEmail": email, "attendeeTimezone": "UTC", "guestEmails": guests, "notes": notes,
	}), http.StatusCreated)
	var result publicBookingResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func cancelBooking(t *testing.T, f *fixture, manageURL, reason string) {
	t.Helper()
	parsed, err := url.Parse(manageURL)
	if err != nil {
		t.Fatal(err)
	}
	secret := strings.TrimPrefix(parsed.Path, "/event/")
	expectStatus(t, f.request(http.MethodPost, "/api/events/"+secret+"/cancel", map[string]string{"reason": reason}), http.StatusOK)
}

func scheduledEventByStart(t *testing.T, f *fixture, token, organization string, start time.Time) map[string]any {
	t.Helper()
	query := url.Values{"organization": {organization}, "min_start_time": {start.Format(time.RFC3339)}, "max_start_time": {start.Add(time.Second).Format(time.RFC3339)}}
	items := objectCollection(t, decodeObject(t, expectStatus(t, f.bearerRequest(http.MethodGet, "/scheduled_events?"+query.Encode(), token, nil), http.StatusOK)))
	if len(items) != 1 {
		t.Fatalf("expected one scheduled event at %s, got %#v", start, items)
	}
	return items[0]
}

func TestPublicOpenAPIAndSelfHostedSwagger(t *testing.T) {
	f := newFixtureAtBasePath(t, false, "/team")
	body := expectStatus(t, f.request(http.MethodGet, compatibilityTestPath+"/openapi.json", nil), http.StatusOK)
	if bytes.Contains(body, []byte("__BASE_URL__")) || !bytes.Contains(body, []byte(`"url": "http://example.test/team/api/v1"`)) {
		t.Fatalf("OpenAPI server URL was not configured: %s", body)
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatalf("OpenAPI response is invalid JSON: %v", err)
	}
	paths := nestedObject(t, document, "paths")
	operations := 0
	for _, raw := range paths {
		for method := range raw.(map[string]any) {
			if method == "get" || method == "post" {
				operations++
			}
		}
	}
	if len(paths) != 7 || operations != 8 {
		t.Fatalf("OpenAPI documents %d paths and %d operations, want 7 and 8", len(paths), operations)
	}
	components := nestedObject(t, document, "components")
	schemas := nestedObject(t, components, "schemas")
	userSchema := nestedObject(t, schemas, "User")
	userProperties := nestedObject(t, userSchema, "properties")
	if nestedObject(t, userProperties, "slug")["x-not-implemented"] != true {
		t.Fatal("OpenAPI unsupported properties are not annotated")
	}
	errorSchema := nestedObject(t, schemas, "Error")
	if errorSchema["additionalProperties"] != false {
		t.Fatal("OpenAPI error schema permits fields outside the error string")
	}
	securitySchemes := nestedObject(t, components, "securitySchemes")
	if nestedObject(t, securitySchemes, "bearerAuth")["scheme"] != "bearer" {
		t.Fatal("OpenAPI bearer security scheme is missing")
	}

	swagger := expectStatus(t, f.request(http.MethodGet, compatibilityTestPath+"/swagger/", nil), http.StatusOK)
	if bytes.Contains(bytes.ToLower(swagger), []byte("cdn")) || !bytes.Contains(swagger, []byte("/team/api/v1/swagger/swagger-ui.css")) {
		t.Fatalf("Swagger page does not reference same-origin embedded assets: %s", swagger)
	}
	asset := f.request(http.MethodGet, compatibilityTestPath+"/swagger/swagger-ui.css", nil)
	expectStatus(t, asset, http.StatusOK)
	if !strings.Contains(asset.Header.Get("Content-Type"), "text/css") {
		t.Fatalf("Swagger stylesheet has unexpected content type %q", asset.Header.Get("Content-Type"))
	}
}
