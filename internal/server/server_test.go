package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configpkg "github.com/yamatt/go-as-webhook/internal/config"
)

func TestHandleHealth(t *testing.T) {
	cfg := configpkg.NewDefault()
	srv := NewAppServer(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestHandleTransaction(t *testing.T) {
	cfg := configpkg.NewDefault()
	srv := NewAppServer(cfg)

	transaction := Transaction{
		Events: []MatrixEvent{
			{
				Type:      "m.room.message",
				EventID:   "$test_event",
				RoomID:    "!room:domain.com",
				Sender:    "@user:domain.com",
				Timestamp: 1234567890,
				Content: map[string]interface{}{
					"body":    "test message",
					"msgtype": "m.text",
				},
			},
		},
	}

	body, err := json.Marshal(transaction)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	req := httptest.NewRequest("PUT", "/_matrix/app/v1/transactions/test123", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleRoom(t *testing.T) {
	cfg := configpkg.NewDefault()
	srv := NewAppServer(cfg)

	req := httptest.NewRequest("GET", "/_matrix/app/v1/rooms/%23room%3Adomain.com", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["errcode"] != "M_NOT_FOUND" {
		t.Errorf("Expected errcode 'M_NOT_FOUND', got '%s'", response["errcode"])
	}
}

func TestHandleUser(t *testing.T) {
	cfg := configpkg.NewDefault()
	srv := NewAppServer(cfg)

	req := httptest.NewRequest("GET", "/_matrix/app/v1/users/%40user%3Adomain.com", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["errcode"] != "M_NOT_FOUND" {
		t.Errorf("Expected errcode 'M_NOT_FOUND', got '%s'", response["errcode"])
	}
}

func TestProcessEventWithWebhook(t *testing.T) {
	webhookCalled := false
	var receivedPayload map[string]interface{}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode webhook payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	cfg := &configpkg.Config{
		Routes: []configpkg.RouteConfig{
			{
				Name:       "match-test",
				Selector:   "event.content.body.contains('test')",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	srv := NewAppServer(cfg)

	event := MatrixEvent{
		Type:      "m.room.message",
		EventID:   "$test_event",
		RoomID:    "!room:domain.com",
		Sender:    "@user:domain.com",
		Timestamp: 1234567890,
		Content: map[string]interface{}{
			"body":    "this is a test message",
			"msgtype": "m.text",
		},
	}

	srv.processEvent(event)

	if !webhookCalled {
		t.Error("Expected webhook to be called")
	}

	if receivedPayload == nil {
		t.Fatal("No payload received")
	}

	if receivedPayload["message"] != "this is a test message" {
		t.Errorf("Expected message 'this is a test message', got '%s'", receivedPayload["message"])
	}

	if receivedPayload["sender"] != "@user:domain.com" {
		t.Errorf("Expected sender '@user:domain.com', got '%s'", receivedPayload["sender"])
	}
}

func TestProcessEventSendBodyFalse(t *testing.T) {
	webhookCalled := false
	var receivedPayload map[string]interface{}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sendBodyFalse := false
	cfg := &configpkg.Config{
		Routes: []configpkg.RouteConfig{
			{
				Name:       "no-body",
				Selector:   "true",
				WebhookURL: testServer.URL,
				Method:     "POST",
				SendBody:   &sendBodyFalse,
			},
		},
	}

	srv := NewAppServer(cfg)

	event := MatrixEvent{
		Type:      "m.room.message",
		EventID:   "$test_event",
		RoomID:    "!room:domain.com",
		Sender:    "@user:domain.com",
		Timestamp: 1234567890,
		Content: map[string]interface{}{
			"body":    "this is a test message",
			"msgtype": "m.text",
		},
	}

	srv.processEvent(event)

	if !webhookCalled {
		t.Error("Expected webhook to be called")
	}

	if receivedPayload == nil {
		t.Fatal("No payload received")
	}

	if _, hasMessage := receivedPayload["message"]; hasMessage {
		t.Errorf("Expected message field to be absent (send_body=false), but it was present: %v", receivedPayload["message"])
	}

	if receivedPayload["event_id"] != "$test_event" {
		t.Errorf("Expected event_id '$test_event', got '%s'", receivedPayload["event_id"])
	}
}

func TestProcessEventStopOnMatch(t *testing.T) {
	firstWebhookCalled := false
	secondWebhookCalled := false

	firstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstWebhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer firstServer.Close()

	secondServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondWebhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer secondServer.Close()

	cfg := &configpkg.Config{
		Routes: []configpkg.RouteConfig{
			{
				Name:        "stop-route",
				Selector:    "true",
				WebhookURL:  firstServer.URL,
				Method:      "POST",
				StopOnMatch: true,
			},
			{
				Name:       "should-not-match",
				Selector:   "true",
				WebhookURL: secondServer.URL,
				Method:     "POST",
			},
		},
	}

	srv := NewAppServer(cfg)

	event := MatrixEvent{
		Type:      "m.room.message",
		EventID:   "$test_event",
		RoomID:    "!room:domain.com",
		Sender:    "@user:domain.com",
		Timestamp: 1234567890,
		Content: map[string]interface{}{
			"body":    "test message",
			"msgtype": "m.text",
		},
	}

	srv.processEvent(event)

	if !firstWebhookCalled {
		t.Error("Expected first webhook to be called")
	}

	if secondWebhookCalled {
		t.Error("Expected second webhook NOT to be called due to stop_on_match")
	}
}

func TestProcessEventNoMatch(t *testing.T) {
	webhookCalled := false

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	cfg := &configpkg.Config{
		Routes: []configpkg.RouteConfig{
			{
				Name:       "no-match",
				Selector:   "event.content.body.contains('alert')",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	srv := NewAppServer(cfg)

	event := MatrixEvent{
		Type:      "m.room.message",
		EventID:   "$test_event",
		RoomID:    "!room:domain.com",
		Sender:    "@user:domain.com",
		Timestamp: 1234567890,
		Content: map[string]interface{}{
			"body":    "this is a test message",
			"msgtype": "m.text",
		},
	}

	srv.processEvent(event)

	if webhookCalled {
		t.Error("Expected webhook not to be called when pattern doesn't match")
	}
}

func TestProcessEventEmptyPattern(t *testing.T) {
	webhookCalled := false

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	cfg := &configpkg.Config{
		Routes: []configpkg.RouteConfig{
			{
				Name:       "match-all",
				Selector:   "true",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	srv := NewAppServer(cfg)

	event := MatrixEvent{
		Type:      "m.room.message",
		EventID:   "$test_event",
		RoomID:    "!room:domain.com",
		Sender:    "@user:domain.com",
		Timestamp: 1234567890,
		Content: map[string]interface{}{
			"body":    "any message",
			"msgtype": "m.text",
		},
	}

	srv.processEvent(event)

	if !webhookCalled {
		t.Error("Expected webhook to be called with empty pattern (matches all)")
	}
}
