package idp

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/breakroom/saml-test-idp/internal/config"
	"github.com/breakroom/saml-test-idp/internal/web"
	"github.com/crewjam/saml"
)

// handleMetadata serves the IDP metadata XML.
func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	// Build metadata with supported Name ID formats
	metadata := s.idp.Metadata()

	// Add all supported Name ID formats
	for i := range metadata.IDPSSODescriptors {
		metadata.IDPSSODescriptors[i].NameIDFormats = []saml.NameIDFormat{
			saml.EmailAddressNameIDFormat,
			saml.PersistentNameIDFormat,
			saml.TransientNameIDFormat,
			saml.UnspecifiedNameIDFormat,
		}
	}

	buf, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		log.Printf("Error marshaling metadata: %v", err)
		http.Error(w, "Failed to generate metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	if _, err := w.Write(buf); err != nil {
		log.Printf("Error writing metadata: %v", err)
	}
}

// handleSSO handles SAML SSO requests.
func (s *Server) handleSSO(w http.ResponseWriter, r *http.Request) {
	// Parse the SAML request
	req, err := saml.NewIdpAuthnRequest(s.idp, r)
	if err != nil {
		log.Printf("Error parsing SAML request: %v", err)
		http.Error(w, "Invalid SAML request", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Error validating SAML request: %v", err)
		http.Error(w, "Invalid SAML request", http.StatusBadRequest)
		return
	}

	// Get SP config
	spConfig := s.spProvider.GetServiceProviderConfig(req.ServiceProviderMetadata.EntityID)
	if spConfig == nil {
		log.Printf("Unknown service provider: %s", req.ServiceProviderMetadata.EntityID)
		http.Error(w, "Unknown service provider", http.StatusBadRequest)
		return
	}

	// Store pending request and redirect to login
	// Always show login page - no session persistence for test IDP
	requestID := randomHex(16)
	s.sessionProvider.StorePendingRequest(requestID, req, spConfig)

	// Redirect to login page
	loginURL := fmt.Sprintf("/login?request_id=%s", url.QueryEscape(requestID))
	http.Redirect(w, r, loginURL, http.StatusFound)
}

// handleLogin handles the login page and user selection.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")
	if requestID == "" {
		http.Error(w, "Missing request_id", http.StatusBadRequest)
		return
	}

	// Get pending request
	pendingSession, ok := s.sessionProvider.GetPendingRequest(requestID)
	if !ok {
		http.Error(w, "Invalid or expired request", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		// Show login page
		s.showLoginPage(w, r, requestID, pendingSession)
		return
	}

	if r.Method == http.MethodPost {
		// Process login
		s.processLogin(w, r, requestID, pendingSession)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// showLoginPage renders the login page with user dropdown.
func (s *Server) showLoginPage(w http.ResponseWriter, r *http.Request, requestID string, pendingSession *SessionData) {
	tmpl, err := template.ParseFS(web.Assets, "templates/login.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := LoginPageData{
		RequestID: requestID,
		SPName:    pendingSession.SP.EntityID,
		Users:     pendingSession.SP.Users,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// LoginPageData holds data for the login template.
type LoginPageData struct {
	RequestID string
	SPName    string
	Users     []config.User
}

// processLogin handles user selection and creates SAML response.
func (s *Server) processLogin(w http.ResponseWriter, r *http.Request, requestID string, pendingSession *SessionData) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userName := r.FormValue("user")
	if userName == "" {
		http.Error(w, "No user selected", http.StatusBadRequest)
		return
	}

	// Find user
	user := pendingSession.SP.GetUserByName(userName)
	if user == nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	// Build SAML session for response (no persistent session - always show login)
	sessionID := randomHex(32)
	samlSession := &saml.Session{
		ID:               sessionID,
		CreateTime:       time.Now(),
		ExpireTime:       time.Now().Add(5 * time.Minute), // Short-lived for response only
		Index:            sessionID,
		NameID:           user.NameID,
		NameIDFormat:     string(GetNameIDFormat(pendingSession.SP.NameIDFormat)),
		SubjectID:        user.NameID,
		UserName:         user.Name,
		CustomAttributes: buildCustomAttributes(user),
	}

	// Create and send SAML response
	s.createAndSendResponse(w, r, pendingSession.SAMLRequest, samlSession)

	// Clean up pending request
	s.sessionProvider.DeletePendingRequest(requestID)
}

// createAndSendResponse creates a SAML response and sends it to the SP.
func (s *Server) createAndSendResponse(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest, session *saml.Session) {
	// Use the default assertion maker to create the assertion
	assertionMaker := saml.DefaultAssertionMaker{}
	if err := assertionMaker.MakeAssertion(req, session); err != nil {
		log.Printf("Error making assertion: %v", err)
		http.Error(w, "Failed to create assertion", http.StatusInternalServerError)
		return
	}

	// Write the response using the library's built-in method
	if err := req.WriteResponse(w); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}
