package idp

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/breakroom/saml-test-idp/internal/config"
)

func TestNewServiceProviderProvider(t *testing.T) {
	sps := []config.ServiceProvider{
		{
			EntityID:     "https://sp1.example.com",
			ACSURL:       "https://sp1.example.com/acs",
			NameIDFormat: "email",
			Users: []config.User{
				{Name: "Alice", NameID: "alice@example.com"},
			},
		},
		{
			EntityID:     "https://sp2.example.com",
			ACSURL:       "https://sp2.example.com/acs",
			NameIDFormat: "persistent",
			Users: []config.User{
				{Name: "Bob", NameID: "bob-uuid"},
			},
		},
	}

	provider, err := NewServiceProviderProvider(sps)
	if err != nil {
		t.Fatalf("NewServiceProviderProvider failed: %v", err)
	}

	// Test GetServiceProvider
	req := httptest.NewRequest("GET", "/sso", nil)

	metadata, err := provider.GetServiceProvider(req, "https://sp1.example.com")
	if err != nil {
		t.Fatalf("GetServiceProvider failed: %v", err)
	}
	if metadata.EntityID != "https://sp1.example.com" {
		t.Errorf("Expected EntityID 'https://sp1.example.com', got '%s'", metadata.EntityID)
	}

	// Test GetServiceProviderConfig
	spConfig := provider.GetServiceProviderConfig("https://sp1.example.com")
	if spConfig == nil {
		t.Fatal("Expected SP config, got nil")
	}
	if spConfig.NameIDFormat != "email" {
		t.Errorf("Expected NameIDFormat 'email', got '%s'", spConfig.NameIDFormat)
	}

	// Test non-existent SP
	_, err = provider.GetServiceProvider(req, "https://nonexistent.example.com")
	if err == nil {
		t.Error("Expected error for non-existent SP")
	}

	spConfig = provider.GetServiceProviderConfig("https://nonexistent.example.com")
	if spConfig != nil {
		t.Error("Expected nil for non-existent SP config")
	}
}

func TestNewServiceProviderProviderWithMetadataFile(t *testing.T) {
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "sp-metadata.xml")

	// Simple SP metadata
	metadataXML := `<?xml version="1.0" encoding="UTF-8"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://sp-from-file.example.com">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
      Location="https://sp-from-file.example.com/acs" index="1"/>
  </SPSSODescriptor>
</EntityDescriptor>`

	if err := os.WriteFile(metadataPath, []byte(metadataXML), 0644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	sps := []config.ServiceProvider{
		{
			EntityID:     "https://sp-from-file.example.com",
			MetadataFile: metadataPath,
			NameIDFormat: "transient",
			Users: []config.User{
				{Name: "Test User", NameID: "test@example.com"},
			},
		},
	}

	provider, err := NewServiceProviderProvider(sps)
	if err != nil {
		t.Fatalf("NewServiceProviderProvider failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/sso", nil)
	metadata, err := provider.GetServiceProvider(req, "https://sp-from-file.example.com")
	if err != nil {
		t.Fatalf("GetServiceProvider failed: %v", err)
	}
	if metadata.EntityID != "https://sp-from-file.example.com" {
		t.Errorf("Expected EntityID from file, got '%s'", metadata.EntityID)
	}
}

func TestNewServiceProviderProviderInvalidConfig(t *testing.T) {
	// SP without ACS URL or metadata file
	sps := []config.ServiceProvider{
		{
			EntityID: "https://invalid.example.com",
			Users: []config.User{
				{Name: "Test", NameID: "test@example.com"},
			},
		},
	}

	_, err := NewServiceProviderProvider(sps)
	if err == nil {
		t.Error("Expected error for SP without ACS URL or metadata file")
	}
}

func TestGetAllServiceProviders(t *testing.T) {
	sps := []config.ServiceProvider{
		{EntityID: "https://sp1.example.com", ACSURL: "https://sp1.example.com/acs"},
		{EntityID: "https://sp2.example.com", ACSURL: "https://sp2.example.com/acs"},
		{EntityID: "https://sp3.example.com", ACSURL: "https://sp3.example.com/acs"},
	}

	provider, err := NewServiceProviderProvider(sps)
	if err != nil {
		t.Fatalf("NewServiceProviderProvider failed: %v", err)
	}

	allSPs := provider.GetAllServiceProviders()
	if len(allSPs) != 3 {
		t.Errorf("Expected 3 SPs, got %d", len(allSPs))
	}
}
