package registration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateBasic(t *testing.T) {
	serverURL := "http://localhost:8080"
	reg, err := Generate(serverURL, "")

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if reg.ID != "matrix-as-webhook" {
		t.Errorf("Expected ID 'matrix-as-webhook', got '%s'", reg.ID)
	}

	if reg.Url != serverURL {
		t.Errorf("Expected URL '%s', got '%s'", serverURL, reg.Url)
	}

	if reg.AsToken == "" {
		t.Error("Expected AS token to be generated, got empty string")
	}

	if reg.HsToken == "" {
		t.Error("Expected HS token to be generated, got empty string")
	}

	if reg.RateLimited != false {
		t.Errorf("Expected RateLimited to be false, got %v", reg.RateLimited)
	}

	if len(reg.Namespaces.Users) != 0 {
		t.Errorf("Expected no user namespaces, got %d", len(reg.Namespaces.Users))
	}
}

func TestGenerateWithCustomAsToken(t *testing.T) {
	serverURL := "https://example.com"
	customToken := "my-custom-token-12345"

	reg, err := Generate(serverURL, customToken)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if reg.AsToken != customToken {
		t.Errorf("Expected AS token '%s', got '%s'", customToken, reg.AsToken)
	}

	if reg.HsToken == "" {
		t.Error("Expected HS token to be generated, got empty string")
	}

	if reg.HsToken == customToken {
		t.Error("Expected HS token to be different from AS token")
	}
}

// Namespace generation is no longer supported via CLI; namespaces remain empty by default.

func TestGenerateTokenUniqueness(t *testing.T) {
	reg1, _ := Generate("http://localhost:8080", "")
	reg2, _ := Generate("http://localhost:8080", "")

	if reg1.AsToken == reg2.AsToken {
		t.Error("Expected different AS tokens for different generations")
	}

	if reg1.HsToken == reg2.HsToken {
		t.Error("Expected different HS tokens for different generations")
	}
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "registration.yaml")

	reg, _ := Generate("http://localhost:8080", "test-token")
	err := reg.WriteToFile(filePath)

	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Registration file was not created at %s", filePath)
	}

	// Check file is readable and contains content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read registration file: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Error("Registration file is empty")
	}

	if !contains(content, "matrix-as-webhook") {
		t.Error("Registration file does not contain ID")
	}

	if !contains(content, "http://localhost:8080") {
		t.Error("Registration file does not contain server URL")
	}

	if !contains(content, "test-token") {
		t.Error("Registration file does not contain AS token")
	}

	// No namespace should be present by default
}

func TestWriteToFileCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "nested", "registration.yaml")

	reg, _ := Generate("http://localhost:8080", "")
	err := reg.WriteToFile(filePath)

	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Registration file was not created at %s", filePath)
	}
}

func TestWriteToFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "registration.yaml")

	reg, _ := Generate("http://localhost:8080", "")
	err := reg.WriteToFile(filePath)

	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Check file permissions (should be 0600 - read/write for owner only)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)

	if mode != expectedMode {
		t.Errorf("Expected file permissions %o, got %o", expectedMode, mode)
	}
}

func TestGenerateWithAllParameters(t *testing.T) {
	serverURL := "https://webhook.example.com"
	asToken := "custom-as-token"

	reg, err := Generate(serverURL, asToken)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if reg.Url != serverURL {
		t.Errorf("Expected URL '%s', got '%s'", serverURL, reg.Url)
	}

	if reg.AsToken != asToken {
		t.Errorf("Expected AS token '%s', got '%s'", asToken, reg.AsToken)
	}

	if len(reg.Namespaces.Users) != 0 {
		t.Fatalf("Expected 0 user namespaces by default, got %d", len(reg.Namespaces.Users))
	}
}

// Helper function
func contains(str, substr string) bool {
	for i := 0; i < len(str)-len(substr)+1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
