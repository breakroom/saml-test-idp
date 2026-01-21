package idp

import (
	"net/http"
	"sync"
	"time"

	"github.com/breakroom/saml-test-idp/internal/config"
	"github.com/crewjam/saml"
	"github.com/spf13/cast"
)

// SessionProvider manages pending SAML requests during the login flow.
// Note: This IDP intentionally does not persist user sessions - each SSO
// request shows the login page to allow selecting different test users.
type SessionProvider struct {
	mu              sync.RWMutex
	pendingRequests map[string]*SessionData
}

// SessionData holds pending SAML request information.
type SessionData struct {
	ID          string
	SP          *config.ServiceProvider
	CreateTime  time.Time
	ExpireTime  time.Time
	SAMLRequest *saml.IdpAuthnRequest
}

// NewSessionProvider creates a new session provider.
func NewSessionProvider() *SessionProvider {
	return &SessionProvider{
		pendingRequests: make(map[string]*SessionData),
	}
}

// GetSession implements saml.SessionProvider.
// Always returns nil to force showing the login page on every SSO request.
// This is intentional for a test IDP - we want users to select a test user each time.
func (sp *SessionProvider) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	return nil
}

// buildCustomAttributes converts user attributes to SAML attributes.
func buildCustomAttributes(user *config.User) []saml.Attribute {
	if user == nil || user.Attributes == nil {
		return nil
	}

	attrs := make([]saml.Attribute, 0, len(user.Attributes))
	for name, value := range user.Attributes {
		attr := saml.Attribute{
			FriendlyName: name,
			Name:         name,
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
			Values:       attributeValues(value),
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// attributeValues converts a value to SAML attribute values.
func attributeValues(value interface{}) []saml.AttributeValue {
	// Handle slices specially to create multiple attribute values
	if slice, err := cast.ToStringSliceE(value); err == nil && len(slice) > 0 {
		values := make([]saml.AttributeValue, 0, len(slice))
		for _, s := range slice {
			values = append(values, saml.AttributeValue{Type: "xs:string", Value: s})
		}
		return values
	}

	// For single values, convert to string
	return []saml.AttributeValue{{Type: "xs:string", Value: cast.ToString(value)}}
}

// StorePendingRequest stores a pending SAML auth request.
func (sp *SessionProvider) StorePendingRequest(requestID string, req *saml.IdpAuthnRequest, spConfig *config.ServiceProvider) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.pendingRequests[requestID] = &SessionData{
		ID:          requestID,
		SP:          spConfig,
		CreateTime:  time.Now(),
		ExpireTime:  time.Now().Add(10 * time.Minute),
		SAMLRequest: req,
	}
}

// GetPendingRequest retrieves a pending SAML auth request.
func (sp *SessionProvider) GetPendingRequest(requestID string) (*SessionData, bool) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	session, ok := sp.pendingRequests[requestID]
	if !ok || session.SAMLRequest == nil {
		return nil, false
	}
	if session.ExpireTime.Before(time.Now()) {
		return nil, false
	}
	return session, true
}

// DeletePendingRequest removes a pending request.
func (sp *SessionProvider) DeletePendingRequest(requestID string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	delete(sp.pendingRequests, requestID)
}

func randomHex(n int) string {
	const hexChars = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hexChars[pseudoRand()%16]
	}
	return string(b)
}

// Simple pseudo-random for session IDs (not cryptographic)
var randState uint64 = uint64(time.Now().UnixNano())

func pseudoRand() int {
	randState = randState*6364136223846793005 + 1
	return int(randState >> 33)
}
