package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Sender handles HTTP dispatch to webhook endpoints.
type Sender struct {
	client *http.Client
}

// NewSender creates a new webhook sender with a configured HTTP client.
func NewSender(timeout time.Duration) *Sender {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Sender{
		client: &http.Client{Timeout: timeout},
	}
}

// Request represents a webhook request to be sent.
type Request struct {
	URL          string
	Method       string
	Payload      map[string]interface{}
	SharedSecret string // Optional shared secret for HMAC-SHA256 signing
}

// Response represents the result of sending a webhook.
type Response struct {
	StatusCode int
	Error      error
	Body       []byte
}

// Send dispatches a webhook request and returns the response.
func (s *Sender) Send(req Request) Response {
	if req.Method == "" {
		req.Method = "POST"
	}

	log.Printf("Webhook: sending to %s with method %s", req.URL, req.Method)

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		log.Printf("Webhook: error marshaling payload: %v", err)
		return Response{Error: err}
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Webhook: error creating request: %v", err)
		return Response{Error: err}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Add signature header if shared secret is provided
	if req.SharedSecret != "" {
		signature := generateSignature(payloadBytes, req.SharedSecret)
		httpReq.Header.Set("X-Webhook-Signature", signature)
		log.Printf("Webhook: added signature header to request")
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		log.Printf("Webhook: error sending to %s: %v", req.URL, err)
		return Response{Error: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		log.Printf("Webhook: error reading response: %v", err)
		return Response{StatusCode: resp.StatusCode, Error: err}
	}

	log.Printf("Webhook: sent to %s, status: %d", req.URL, resp.StatusCode)

	if resp.StatusCode >= 400 {
		log.Printf("Webhook: error response from %s: %s", req.URL, string(respBody))
	}

	return Response{StatusCode: resp.StatusCode, Body: respBody}
}

// generateSignature creates an HMAC-SHA256 signature of the payload.
// The signature is hex-encoded and prefixed with "sha256=" for clarity.
func generateSignature(payload []byte, sharedSecret string) string {
	h := hmac.New(sha256.New, []byte(sharedSecret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return "sha256=" + signature
}
