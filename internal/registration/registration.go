package registration

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RegistrationFile represents the Matrix Application Service registration file.
type RegistrationFile struct {
	ID          string                 `yaml:"id"`
	Url         string                 `yaml:"url"`
	AsToken     string                 `yaml:"as_token"`
	HsToken     string                 `yaml:"hs_token"`
	RateLimited bool                   `yaml:"rate_limited"`
	Namespaces  Namespaces             `yaml:"namespaces"`
	SoloUnit    bool                   `yaml:"solo_unit,omitempty"`
	Protocols   []string               `yaml:"protocols,omitempty"`
	Limits      map[string]interface{} `yaml:"limits,omitempty"`
}

// Namespaces defines the namespace configuration
type Namespaces struct {
	Users   []Namespace `yaml:"users,omitempty"`
	Aliases []Namespace `yaml:"aliases,omitempty"`
	Rooms   []Namespace `yaml:"rooms,omitempty"`
}

// Namespace represents a single namespace entry
type Namespace struct {
	Exclusive bool   `yaml:"exclusive"`
	Regex     string `yaml:"regex"`
}

// Generate creates a new registration file with the given configuration
func Generate(serverURL string, asToken string) (*RegistrationFile, error) {
	// Generate tokens if not provided
	if asToken == "" {
		token, err := generateToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate AS token: %w", err)
		}
		asToken = token
	}

	hsToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate HS token: %w", err)
	}

	reg := &RegistrationFile{
		ID:          "matrix-as-webhook",
		Url:         serverURL,
		AsToken:     asToken,
		HsToken:     hsToken,
		RateLimited: false,
		Namespaces: Namespaces{
			Users:   []Namespace{},
			Aliases: []Namespace{},
			Rooms:   []Namespace{},
		},
	}

	return reg, nil
}

// WriteToFile saves the registration to a YAML file
func (r *RegistrationFile) WriteToFile(path string) error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write registration file: %w", err)
	}

	return nil
}

// generateToken creates a random token (32 bytes as hex string)
func generateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
