package idp

import (
	"testing"

	"github.com/breakroom/saml-test-idp/internal/config"
)

func TestBuildCustomAttributes(t *testing.T) {
	user := &config.User{
		Name:   "Test User",
		NameID: "test@example.com",
		Attributes: map[string]interface{}{
			"email":     "test@example.com",
			"firstName": "Test",
			"lastName":  "User",
			"groups":    []interface{}{"admin", "users"},
			"active":    true,
		},
	}

	attrs := buildCustomAttributes(user)
	if len(attrs) != 5 {
		t.Errorf("Expected 5 attributes, got %d", len(attrs))
	}

	// Check that attributes are present
	attrMap := make(map[string]bool)
	for _, attr := range attrs {
		attrMap[attr.Name] = true
	}

	expectedAttrs := []string{"email", "firstName", "lastName", "groups", "active"}
	for _, name := range expectedAttrs {
		if !attrMap[name] {
			t.Errorf("Expected attribute '%s' not found", name)
		}
	}
}

func TestBuildCustomAttributesNil(t *testing.T) {
	attrs := buildCustomAttributes(nil)
	if attrs != nil {
		t.Error("Expected nil for nil user")
	}

	user := &config.User{Name: "No Attrs"}
	attrs = buildCustomAttributes(user)
	if attrs != nil {
		t.Error("Expected nil for user with no attributes")
	}
}

func TestAttributeValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int // number of values
	}{
		{"string", "test", 1},
		{"string slice interface", []interface{}{"a", "b", "c"}, 3},
		{"string slice", []string{"a", "b"}, 2},
		{"bool true", true, 1},
		{"bool false", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := attributeValues(tt.input)
			if len(values) != tt.expected {
				t.Errorf("attributeValues(%v) returned %d values, want %d", tt.input, len(values), tt.expected)
			}
		})
	}
}

func TestSessionProviderPendingRequests(t *testing.T) {
	sp := NewSessionProvider()

	spConfig := &config.ServiceProvider{
		EntityID: "https://sp.example.com",
	}

	requestID := "test-request-123"

	// Store pending request (with nil SAMLRequest for testing)
	sp.StorePendingRequest(requestID, nil, spConfig)

	// Note: GetPendingRequest returns false if SAMLRequest is nil
	// This is intentional - we're testing the storage mechanism
	if len(sp.pendingRequests) != 1 {
		t.Errorf("Expected 1 pending request, got %d", len(sp.pendingRequests))
	}

	// Test deletion
	sp.DeletePendingRequest(requestID)
	if len(sp.pendingRequests) != 0 {
		t.Error("Expected pending request to be deleted")
	}
}

func TestRandomHex(t *testing.T) {
	hex := randomHex(16)
	if len(hex) != 16 {
		t.Errorf("Expected length 16, got %d", len(hex))
	}

	// Verify all characters are hex
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Invalid hex character: %c", c)
		}
	}

	// Test uniqueness
	hex2 := randomHex(16)
	if hex == hex2 {
		t.Error("randomHex should generate unique values")
	}
}

func TestGetSessionAlwaysReturnsNil(t *testing.T) {
	sp := NewSessionProvider()

	// GetSession should always return nil (no persistent sessions in test IDP)
	session := sp.GetSession(nil, nil, nil)
	if session != nil {
		t.Error("GetSession should always return nil for test IDP")
	}
}
