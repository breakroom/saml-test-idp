package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: "0.0.0.0"
  port: 9090
  base_url: "https://idp.example.com"

idp:
  entity_id: "https://idp.example.com/metadata"
  certificate_path: "test.crt"
  private_key_path: "test.key"

service_providers:
  - entity_id: "https://sp.example.com"
    acs_url: "https://sp.example.com/acs"
    name_id_format: "persistent"
    users:
      - name: "Test User"
        name_id: "test@example.com"
        attributes:
          email: "test@example.com"
          role: "admin"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test server config
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.BaseURL != "https://idp.example.com" {
		t.Errorf("Expected base_url 'https://idp.example.com', got '%s'", cfg.Server.BaseURL)
	}

	// Test IDP config
	if cfg.IDP.EntityID != "https://idp.example.com/metadata" {
		t.Errorf("Expected entity_id 'https://idp.example.com/metadata', got '%s'", cfg.IDP.EntityID)
	}

	// Test service providers
	if len(cfg.ServiceProviders) != 1 {
		t.Fatalf("Expected 1 service provider, got %d", len(cfg.ServiceProviders))
	}

	sp := cfg.ServiceProviders[0]
	if sp.EntityID != "https://sp.example.com" {
		t.Errorf("Expected SP entity_id 'https://sp.example.com', got '%s'", sp.EntityID)
	}
	if sp.ACSURL != "https://sp.example.com/acs" {
		t.Errorf("Expected SP acs_url 'https://sp.example.com/acs', got '%s'", sp.ACSURL)
	}
	if sp.NameIDFormat != "persistent" {
		t.Errorf("Expected SP name_id_format 'persistent', got '%s'", sp.NameIDFormat)
	}

	// Test users
	if len(sp.Users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(sp.Users))
	}

	user := sp.Users[0]
	if user.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got '%s'", user.Name)
	}
	if user.NameID != "test@example.com" {
		t.Errorf("Expected user name_id 'test@example.com', got '%s'", user.NameID)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Minimal config without defaults
	configContent := `
idp:
  certificate_path: "test.crt"
  private_key_path: "test.key"

service_providers:
  - entity_id: "https://sp.example.com"
    acs_url: "https://sp.example.com/acs"
    users:
      - name: "Test User"
        name_id: "test@example.com"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test defaults
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	// Test default name_id_format
	if cfg.ServiceProviders[0].NameIDFormat != "email" {
		t.Errorf("Expected default name_id_format 'email', got '%s'", cfg.ServiceProviders[0].NameIDFormat)
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestServiceProviderGetUserByName(t *testing.T) {
	sp := &ServiceProvider{
		Users: []User{
			{Name: "Alice", NameID: "alice@example.com"},
			{Name: "Bob", NameID: "bob@example.com"},
		},
	}

	// Test finding existing user
	user := sp.GetUserByName("Alice")
	if user == nil {
		t.Fatal("Expected to find user Alice")
	}
	if user.NameID != "alice@example.com" {
		t.Errorf("Expected NameID 'alice@example.com', got '%s'", user.NameID)
	}

	// Test finding non-existing user
	user = sp.GetUserByName("Charlie")
	if user != nil {
		t.Error("Expected nil for non-existing user")
	}
}

func TestLoadCertificate(t *testing.T) {
	// Use the certificate from testdata
	certPath := "../../testdata/test.crt"

	// Test loading from path
	idpCfg := &IDPConfig{CertificatePath: certPath}
	cert, err := idpCfg.LoadCertificate()
	if err != nil {
		t.Fatalf("LoadCertificate failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Expected certificate, got nil")
	}

	// Test loading inline
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate file: %v", err)
	}
	idpCfg = &IDPConfig{Certificate: string(certPEM)}
	cert, err = idpCfg.LoadCertificate()
	if err != nil {
		t.Fatalf("LoadCertificate (inline) failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Expected certificate, got nil")
	}

	// Test missing certificate
	idpCfg = &IDPConfig{}
	_, err = idpCfg.LoadCertificate()
	if err == nil {
		t.Error("Expected error for missing certificate")
	}
}

func TestLoadPrivateKey(t *testing.T) {
	// Use the private key from testdata
	keyPath := "../../testdata/test.key"

	// Test loading from path
	idpCfg := &IDPConfig{PrivateKeyPath: keyPath}
	key, err := idpCfg.LoadPrivateKey()
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}
	if key == nil {
		t.Fatal("Expected private key, got nil")
	}

	// Test loading inline
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read private key file: %v", err)
	}
	idpCfg = &IDPConfig{PrivateKey: string(keyPEM)}
	key, err = idpCfg.LoadPrivateKey()
	if err != nil {
		t.Fatalf("LoadPrivateKey (inline) failed: %v", err)
	}
	if key == nil {
		t.Fatal("Expected private key, got nil")
	}

	// Test missing private key
	idpCfg = &IDPConfig{}
	_, err = idpCfg.LoadPrivateKey()
	if err == nil {
		t.Error("Expected error for missing private key")
	}
}
