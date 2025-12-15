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
