package idp

import (
	"testing"

	"github.com/crewjam/saml"
)

func TestGetNameIDFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected saml.NameIDFormat
	}{
		{"email", saml.EmailAddressNameIDFormat},
		{"persistent", saml.PersistentNameIDFormat},
		{"transient", saml.TransientNameIDFormat},
		{"unspecified", saml.UnspecifiedNameIDFormat},
		{"unknown", saml.EmailAddressNameIDFormat}, // Default
		{"", saml.EmailAddressNameIDFormat},        // Empty defaults to email
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetNameIDFormat(tt.input)
			if result != tt.expected {
				t.Errorf("GetNameIDFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetNameIDFormatString(t *testing.T) {
	tests := []struct {
		input    saml.NameIDFormat
		expected string
	}{
		{saml.EmailAddressNameIDFormat, "email"},
		{saml.PersistentNameIDFormat, "persistent"},
		{saml.TransientNameIDFormat, "transient"},
		{saml.UnspecifiedNameIDFormat, "unspecified"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetNameIDFormatString(tt.input)
			if result != tt.expected {
				t.Errorf("GetNameIDFormatString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNameIDFormatsCompleteness(t *testing.T) {
	// Ensure all formats can round-trip
	for name, format := range NameIDFormats {
		result := GetNameIDFormatString(format)
		if result != name {
			t.Errorf("Round-trip failed for %q: got %q", name, result)
		}
	}
}
