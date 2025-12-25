package args

import "testing"

func TestParseDefaults(t *testing.T) {
	parsed, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.ConfigPath != "config.toml" {
		t.Errorf("expected default config path 'config.toml', got %q", parsed.ConfigPath)
	}

	if parsed.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", parsed.Port)
	}

	if parsed.GenerateRegistration != "" {
		t.Errorf("expected empty GenerateRegistration, got %q", parsed.GenerateRegistration)
	}

	if parsed.Server != "http://localhost:8080" {
		t.Errorf("expected default server 'http://localhost:8080', got %q", parsed.Server)
	}

	if parsed.AsToken != "" {
		t.Errorf("expected empty AsToken, got %q", parsed.AsToken)
	}
}

func TestParseOverride(t *testing.T) {
	parsed, err := Parse([]string{"-config", "custom.json", "-port", "9000"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.ConfigPath != "custom.json" {
		t.Errorf("expected config path 'custom.json', got %q", parsed.ConfigPath)
	}

	if parsed.Port != 9000 {
		t.Errorf("expected port 9000, got %d", parsed.Port)
	}
}

func TestParseGenerateRegistration(t *testing.T) {
	parsed, err := Parse([]string{
		"-generate-registration", "registration.yaml",
		"-server", "https://webhook.example.com",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.GenerateRegistration != "registration.yaml" {
		t.Errorf("expected GenerateRegistration 'registration.yaml', got %q", parsed.GenerateRegistration)
	}

	if parsed.Server != "https://webhook.example.com" {
		t.Errorf("expected server 'https://webhook.example.com', got %q", parsed.Server)
	}
}

// Namespace flag removed; no test is needed for it.

func TestParseGenerateRegistrationWithCustomToken(t *testing.T) {
	parsed, err := Parse([]string{
		"-generate-registration", "registration.yaml",
		"-server", "http://localhost:8080",
		"-as-token", "my-custom-token",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.AsToken != "my-custom-token" {
		t.Errorf("expected AsToken 'my-custom-token', got %q", parsed.AsToken)
	}
}

func TestParseAllGenerationFlags(t *testing.T) {
	parsed, err := Parse([]string{
		"-generate-registration", "reg.yaml",
		"-server", "https://app.example.com",
		"-as-token", "token123",
		"-port", "9000",
		"-config", "config.toml",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parsed.GenerateRegistration != "reg.yaml" {
		t.Errorf("expected GenerateRegistration 'reg.yaml', got %q", parsed.GenerateRegistration)
	}

	if parsed.Server != "https://app.example.com" {
		t.Errorf("expected server 'https://app.example.com', got %q", parsed.Server)
	}

	if parsed.AsToken != "token123" {
		t.Errorf("expected AsToken 'token123', got %q", parsed.AsToken)
	}

	if parsed.Port != 9000 {
		t.Errorf("expected port 9000, got %d", parsed.Port)
	}

	if parsed.ConfigPath != "config.toml" {
		t.Errorf("expected ConfigPath 'config.toml', got %q", parsed.ConfigPath)
	}
}
