package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yamatt/go-as-webhook/internal/config"
	"github.com/yamatt/go-as-webhook/internal/router"
	"github.com/yamatt/go-as-webhook/internal/webhook"
)

// AppServer represents the Matrix Application Server.
type AppServer struct {
	config        *config.Config
	webhookSender *webhook.Sender
}

// NewAppServer creates a new application server instance.
func NewAppServer(cfg *config.Config) *AppServer {
	return &AppServer{
		config:        cfg,
		webhookSender: webhook.NewSender(30 * time.Second),
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
		s.dispatchWebhook(event, t)
	}
}

// dispatchWebhook constructs a webhook payload and sends it via the webhook module.
func (s *AppServer) dispatchWebhook(event MatrixEvent, target router.Target) {
	payload := map[string]interface{}{
		"event_id":   event.EventID,
		"room_id":    event.RoomID,
		"sender":     event.Sender,
		"timestamp":  event.Timestamp,
		"content":    event.Content,
		"event_type": event.Type,
	}

	// Conditionally include message body based on send_body flag
	if target.SendBody {
		body, ok := event.Content["body"].(string)
		if ok {
			payload["message"] = body
		}
	}

	req := webhook.Request{
		URL:          target.URL,
		Method:       target.Method,
		Payload:      payload,
		SharedSecret: target.SharedSecret,
	}

	_ = s.webhookSender.Send(req)
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
