package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSend_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:    testServer.URL,
		Method: "POST",
		Payload: map[string]interface{}{
			"test": "data",
		},
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestSend_DefaultMethod(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected default method POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:     testServer.URL,
		Payload: map[string]interface{}{"test": "data"},
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestSend_WithSignature(t *testing.T) {
	sharedSecret := "my-secret-key"
	var receivedSignature string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Webhook-Signature")
		if receivedSignature == "" {
			t.Error("Expected X-Webhook-Signature header to be present")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:          testServer.URL,
		Method:       "POST",
		Payload:      map[string]interface{}{"test": "data"},
		SharedSecret: sharedSecret,
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if receivedSignature[:7] != "sha256=" {
		t.Errorf("Expected signature to start with 'sha256=', got %s", receivedSignature[:7])
	}
}

func TestGenerateSignature(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "test-secret"

	sig := generateSignature(payload, secret)

	if sig[:7] != "sha256=" {
		t.Errorf("Expected signature to start with 'sha256=', got %s", sig[:7])
	}

	sig2 := generateSignature(payload, secret)
	if sig != sig2 {
		t.Errorf("Signature should be deterministic, got different results: %s vs %s", sig, sig2)
	}

	expectedHash := hmac.New(sha256.New, []byte(secret))
	expectedHash.Write(payload)
	expectedSig := "sha256=" + hex.EncodeToString(expectedHash.Sum(nil))
	if sig != expectedSig {
		t.Errorf("Signature mismatch: got %s, expected %s", sig, expectedSig)
	}
}

func TestGenerateSignature_DifferentSecrets(t *testing.T) {
	payload := []byte(`{"test":"data"}`)

	sig1 := generateSignature(payload, "secret1")
	sig2 := generateSignature(payload, "secret2")

	if sig1 == sig2 {
		t.Error("Different secrets should produce different signatures")
	}
}

func TestSend_NoSignatureWithoutSecret(t *testing.T) {
	var receivedSignature string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Webhook-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:     testServer.URL,
		Method:  "POST",
		Payload: map[string]interface{}{"test": "data"},
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if receivedSignature != "" {
		t.Errorf("Expected no signature header when SharedSecret is empty, got %s", receivedSignature)
	}
}

func TestSend_ErrorResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:     testServer.URL,
		Method:  "POST",
		Payload: map[string]interface{}{"test": "data"},
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
	if string(resp.Body) != `{"error": "server error"}` {
		t.Errorf("Expected response body, got %s", string(resp.Body))
	}
}

func TestSend_InvalidURL(t *testing.T) {
	sender := NewSender(5 * time.Second)
	req := Request{
		URL:     "http://invalid-nonexistent-domain.test:99999",
		Method:  "POST",
		Payload: map[string]interface{}{"test": "data"},
	}

	resp := sender.Send(req)
	if resp.Error == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestSend_CustomHTTPMethod(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := NewSender(5 * time.Second)
	req := Request{
		URL:     testServer.URL,
		Method:  "PUT",
		Payload: map[string]interface{}{"test": "data"},
	}

	resp := sender.Send(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}
