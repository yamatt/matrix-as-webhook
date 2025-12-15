package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configContent := `[[routes]]
name = "test"
selector = "true"
webhook_url = "http://example.com/webhook"
method = "POST"

[[routes]]
name = "alert"
selector = "true"
webhook_url = "http://example.com/alert"`

	tmpfile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(cfg.Routes))
	}

	if cfg.Routes[0].Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", cfg.Routes[0].Name)
	}

	if cfg.Routes[0].WebhookURL != "http://example.com/webhook" {
		t.Errorf("Expected webhook URL 'http://example.com/webhook', got '%s'", cfg.Routes[0].WebhookURL)
	}

	if cfg.Routes[0].Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", cfg.Routes[0].Method)
	}

	if cfg.Routes[1].Method != "POST" {
		t.Errorf("Expected default method 'POST', got '%s'", cfg.Routes[1].Method)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	if _, err := Load("/nonexistent/config.json"); err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefault()

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.Routes == nil {
		t.Error("Expected routes to be initialized")
	}

	if len(cfg.Routes) != 0 {
		t.Errorf("Expected empty routes, got %d routes", len(cfg.Routes))
	}
}
