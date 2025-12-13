package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `{
		"routes": [
			{
				"pattern": "test",
				"webhook_url": "http://example.com/webhook",
				"method": "POST"
			},
			{
				"pattern": "alert",
				"webhook_url": "http://example.com/alert"
			}
		]
	}`

	tmpfile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	config, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(config.Routes))
	}

	if config.Routes[0].Pattern != "test" {
		t.Errorf("Expected pattern 'test', got '%s'", config.Routes[0].Pattern)
	}

	if config.Routes[0].WebhookURL != "http://example.com/webhook" {
		t.Errorf("Expected webhook URL 'http://example.com/webhook', got '%s'", config.Routes[0].WebhookURL)
	}

	if config.Routes[0].Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", config.Routes[0].Method)
	}

	// Second route should have default method
	if config.Routes[1].Method != "POST" {
		t.Errorf("Expected default method 'POST', got '%s'", config.Routes[1].Method)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Routes == nil {
		t.Error("Expected routes to be initialized")
	}

	if len(config.Routes) != 0 {
		t.Errorf("Expected empty routes, got %d routes", len(config.Routes))
	}
}
