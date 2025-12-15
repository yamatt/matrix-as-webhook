package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yamatt/go-as-webhook/internal/config"
	"github.com/yamatt/go-as-webhook/internal/router"
)

// AppServer represents the Matrix Application Server.
type AppServer struct {
	config     *config.Config
	httpClient *http.Client
}

// NewAppServer creates a new application server instance.
func NewAppServer(cfg *config.Config) *AppServer {
	return &AppServer{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Router sets up the HTTP routes for the application server.
func (s *AppServer) Router() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/_matrix/app/v1/transactions/{txnId}", s.handleTransaction).Methods("PUT")
	r.HandleFunc("/_matrix/app/v1/rooms/{roomAlias}", s.handleRoom).Methods("GET")
	r.HandleFunc("/_matrix/app/v1/users/{userId}", s.handleUser).Methods("GET")
	r.HandleFunc("/health", s.handleHealth).Methods("GET")

	return r
}

// MatrixEvent represents a Matrix event.
type MatrixEvent struct {
	Type      string                 `json:"type"`
	EventID   string                 `json:"event_id"`
	RoomID    string                 `json:"room_id"`
	Sender    string                 `json:"sender"`
	Timestamp int64                  `json:"origin_server_ts"`
	Content   map[string]interface{} `json:"content"`
}

// Transaction represents a Matrix transaction.
type Transaction struct {
	Events []MatrixEvent `json:"events"`
}

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
	for _, event := range transaction.Events {
		s.processEvent(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (s *AppServer) processEvent(event MatrixEvent) {
	log.Printf("Processing event: type=%s, room=%s, sender=%s", event.Type, event.RoomID, event.Sender)

	if event.Type != "m.room.message" {
		log.Printf("Skipping non-message event type: %s", event.Type)
		return
	}

	body, ok := event.Content["body"].(string)
	if !ok {
		log.Printf("Event has no body content")
		return
	}

	log.Printf("Message body: %s", body)

	// Use router resolver to compute targets (supports JSONata).
	res, err := router.NewResolver(s.config)
	if err != nil {
		log.Printf("Router init error: %v", err)
		return
	}
	targets, err := res.Resolve(event)
	if err != nil {
		log.Printf("Router resolve error: %v", err)
		return
	}
	if len(targets) == 0 {
		log.Printf("No routes matched for event %s in room %s", event.EventID, event.RoomID)
		return
	}
	for _, t := range targets {
		log.Printf("Forwarding event %s to route '%s' -> %s (%s)", event.EventID, t.Name, t.URL, t.Method)
		s.sendWebhook(config.RouteConfig{WebhookURL: t.URL, Method: t.Method}, event, body)
	}
}

func (s *AppServer) sendWebhook(route config.RouteConfig, event MatrixEvent, messageBody string) {
	payload := map[string]interface{}{
		"event_id":   event.EventID,
		"room_id":    event.RoomID,
		"sender":     event.Sender,
		"timestamp":  event.Timestamp,
		"message":    messageBody,
		"content":    event.Content,
		"event_type": event.Type,
	}

	log.Printf("Sending webhook to %s with method %s", route.WebhookURL, route.Method)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling webhook payload: %v", err)
		return
	}

	req, err := http.NewRequest(route.Method, route.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating webhook request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending webhook to %s: %v", route.WebhookURL, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook sent to %s, status: %d", route.WebhookURL, resp.StatusCode)

	if resp.StatusCode >= 400 {
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if err != nil {
			log.Printf("Error reading webhook response: %v", err)
		} else {
			log.Printf("Webhook error response: %s", string(respBody))
		}
	}
}

func (s *AppServer) handleRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomAlias := vars["roomAlias"]

	log.Printf("Room query for: %s", roomAlias)

	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"errcode": "M_NOT_FOUND",
		"error":   fmt.Sprintf("Room alias %s not found", roomAlias),
	})
}

func (s *AppServer) handleUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	log.Printf("User query for: %s", userID)

	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"errcode": "M_NOT_FOUND",
		"error":   fmt.Sprintf("User %s not found", userID),
	})
}

func (s *AppServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
