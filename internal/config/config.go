package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

const defaultHTTPMethod = "POST"

// Config represents the application configuration.
type Config struct {
	Routes []RouteConfig `toml:"routes"`
}

// RouteConfig defines a routing rule for messages.
type RouteConfig struct {
	// Optional human-friendly route name.
	Name string `toml:"name"`
	// Selector is a CEL expression evaluated against the event JSON as `event`.
	// Should return a boolean indicating whether this route matches.
	Selector string `toml:"selector"`
	// WebhookURL is the destination URL to send the HTTP request
	WebhookURL string `toml:"webhook_url"`
	// Method is the HTTP method to use (default: POST)
	Method string `toml:"method,omitempty"`
	// StopOnMatch prevents further routes from being evaluated if this route matches (default: false)
	StopOnMatch bool `toml:"stop_on_match,omitempty"`
	// SendBody controls whether the message body is included in the webhook payload (default: true)
	SendBody *bool `toml:"send_body,omitempty"`
}

// Load reads configuration from a TOML file and applies defaults.
func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if _, err := toml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	ApplyDefaults(&cfg)

	return &cfg, nil
}

// ApplyDefaults ensures missing values are populated with sensible defaults.
func ApplyDefaults(cfg *Config) {
	if cfg == nil {
		return
	}

	for i := range cfg.Routes {
		r := &cfg.Routes[i]
		if r.Method == "" {
			r.Method = defaultHTTPMethod
		}
		if r.Selector == "" {
			r.Selector = "true" // default catch-all
		}
		if r.Name == "" {
			r.Name = r.WebhookURL
		}
		// Default SendBody to true if not specified
		if r.SendBody == nil {
			v := true
			r.SendBody = &v
		}
	}
}

// NewDefault creates a default configuration with defaults applied.
func NewDefault() *Config {
	cfg := &Config{Routes: []RouteConfig{}}
	ApplyDefaults(cfg)
	return cfg
}
