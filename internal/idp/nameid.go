package idp

import "github.com/crewjam/saml"

// NameIDFormats maps friendly config names to SAML NameID format URIs.
var NameIDFormats = map[string]saml.NameIDFormat{
	"email":       saml.EmailAddressNameIDFormat,
	"persistent":  saml.PersistentNameIDFormat,
	"transient":   saml.TransientNameIDFormat,
	"unspecified": saml.UnspecifiedNameIDFormat,
}

// GetNameIDFormat returns the SAML NameID format for a given config value.
// Returns EmailAddressNameIDFormat as default if not found.
func GetNameIDFormat(format string) saml.NameIDFormat {
	if f, ok := NameIDFormats[format]; ok {
		return f
	}
	return saml.EmailAddressNameIDFormat
}

// GetNameIDFormatString returns the config-friendly name for a NameID format.
func GetNameIDFormatString(format saml.NameIDFormat) string {
	for name, f := range NameIDFormats {
		if f == format {
			return name
		}
	}
	return "email"
}
