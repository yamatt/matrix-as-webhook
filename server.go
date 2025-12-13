package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// AppServer represents the Matrix Application Server
type AppServer struct {
	config     *Config
	httpClient *http.Client
}

// NewAppServer creates a new application server instance
func NewAppServer(config *Config) *AppServer {
	return &AppServer{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Router sets up the HTTP routes for the application server
func (s *AppServer) Router() http.Handler {
	r := mux.NewRouter()

	// Matrix Application Server Protocol endpoints
	r.HandleFunc("/_matrix/app/v1/transactions/{txnId}", s.handleTransaction).Methods("PUT")
	r.HandleFunc("/_matrix/app/v1/rooms/{roomAlias}", s.handleRoom).Methods("GET")
	r.HandleFunc("/_matrix/app/v1/users/{userId}", s.handleUser).Methods("GET")

	// Health check endpoint
	r.HandleFunc("/health", s.handleHealth).Methods("GET")

	return r
}

// MatrixEvent represents a Matrix event
type MatrixEvent struct {
	Type      string                 `json:"type"`
	EventID   string                 `json:"event_id"`
	RoomID    string                 `json:"room_id"`
	Sender    string                 `json:"sender"`
	Timestamp int64                  `json:"origin_server_ts"`
	Content   map[string]interface{} `json:"content"`
}

// Transaction represents a Matrix transaction
type Transaction struct {
	Events []MatrixEvent `json:"events"`
}

// handleTransaction handles incoming Matrix transactions
func (s *AppServer) handleTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txnID := vars["txnId"]

	log.Printf("Received transaction: %s", txnID)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var transaction Transaction
	if err := json.Unmarshal(body, &transaction); err != nil {
		log.Printf("Error parsing transaction: %v", err)
		http.Error(w, "Error parsing transaction", http.StatusBadRequest)
		return
	}

	log.Printf("Processing %d events", len(transaction.Events))

	// Process each event in the transaction
	for _, event := range transaction.Events {
		s.processEvent(event)
	}

	// Return empty JSON object as success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// processEvent processes a single Matrix event and routes it to webhooks
func (s *AppServer) processEvent(event MatrixEvent) {
	log.Printf("Processing event: type=%s, room=%s, sender=%s", event.Type, event.RoomID, event.Sender)

	// Only process message events
	if event.Type != "m.room.message" {
		log.Printf("Skipping non-message event type: %s", event.Type)
		return
	}

	// Extract message body
	body, ok := event.Content["body"].(string)
	if !ok {
		log.Printf("Event has no body content")
		return
	}

	log.Printf("Message body: %s", body)

	// Find matching routes and send webhooks
	for _, route := range s.config.Routes {
		if route.Pattern == "" || strings.Contains(body, route.Pattern) {
			log.Printf("Matched pattern '%s', sending to webhook: %s", route.Pattern, route.WebhookURL)
			s.sendWebhook(route, event, body)
		}
	}
}

// sendWebhook sends an HTTP request to the configured webhook URL
func (s *AppServer) sendWebhook(route RouteConfig, event MatrixEvent, messageBody string) {
	// Prepare webhook payload
	payload := map[string]interface{}{
		"event_id":   event.EventID,
		"room_id":    event.RoomID,
		"sender":     event.Sender,
		"timestamp":  event.Timestamp,
		"message":    messageBody,
		"content":    event.Content,
		"event_type": event.Type,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling webhook payload: %v", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest(route.Method, route.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending webhook to %s: %v", route.WebhookURL, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook sent to %s, status: %d", route.WebhookURL, resp.StatusCode)

	// Log response body if error status
	if resp.StatusCode >= 400 {
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
		if err != nil {
			log.Printf("Error reading webhook response: %v", err)
		} else {
			log.Printf("Webhook error response: %s", string(respBody))
		}
	}
}

// handleRoom handles room alias queries
func (s *AppServer) handleRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomAlias := vars["roomAlias"]

	log.Printf("Room query for: %s", roomAlias)

	// Return 404 - we don't manage rooms
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"errcode": "M_NOT_FOUND",
		"error":   fmt.Sprintf("Room alias %s not found", roomAlias),
	})
}

// handleUser handles user queries
func (s *AppServer) handleUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	log.Printf("User query for: %s", userID)

	// Return 404 - we don't manage users
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"errcode": "M_NOT_FOUND",
		"error":   fmt.Sprintf("User %s not found", userID),
	})
}

// handleHealth handles health check requests
func (s *AppServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
