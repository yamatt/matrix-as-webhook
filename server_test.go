package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	config := NewDefaultConfig()
	server := NewAppServer(config)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

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
	config := NewDefaultConfig()
	server := NewAppServer(config)

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

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleRoom(t *testing.T) {
	config := NewDefaultConfig()
	server := NewAppServer(config)

	req := httptest.NewRequest("GET", "/_matrix/app/v1/rooms/%23room%3Adomain.com", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

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
	config := NewDefaultConfig()
	server := NewAppServer(config)

	req := httptest.NewRequest("GET", "/_matrix/app/v1/users/%40user%3Adomain.com", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

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
	// Create a test webhook server
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

	// Create config with test webhook
	config := &Config{
		Routes: []RouteConfig{
			{
				Pattern:    "test",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	server := NewAppServer(config)

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

	server.processEvent(event)

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

func TestProcessEventNoMatch(t *testing.T) {
	webhookCalled := false

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	config := &Config{
		Routes: []RouteConfig{
			{
				Pattern:    "alert",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	server := NewAppServer(config)

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

	server.processEvent(event)

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

	config := &Config{
		Routes: []RouteConfig{
			{
				Pattern:    "",
				WebhookURL: testServer.URL,
				Method:     "POST",
			},
		},
	}

	server := NewAppServer(config)

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

	server.processEvent(event)

	if !webhookCalled {
		t.Error("Expected webhook to be called with empty pattern (matches all)")
	}
}
