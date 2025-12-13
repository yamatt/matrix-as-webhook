package main

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration
type Config struct {
	Routes []RouteConfig `json:"routes"`
}

// RouteConfig defines a routing rule for messages
type RouteConfig struct {
	// Pattern to match in message body (simple substring match)
	Pattern string `json:"pattern"`
	// WebhookURL is the destination URL to send the HTTP request
	WebhookURL string `json:"webhook_url"`
	// Method is the HTTP method to use (default: POST)
	Method string `json:"method,omitempty"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Set default HTTP method if not specified
	for i := range config.Routes {
		if config.Routes[i].Method == "" {
			config.Routes[i].Method = "POST"
		}
	}

	return &config, nil
}

// NewDefaultConfig creates a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Routes: []RouteConfig{},
	}
}
