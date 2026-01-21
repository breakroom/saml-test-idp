package idp

import (
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/url"

	"github.com/breakroom/saml-test-idp/internal/config"
	"github.com/crewjam/saml"
)

// Server represents the SAML Identity Provider server.
type Server struct {
	config          *config.Config
	certificate     *x509.Certificate
	privateKey      *rsa.PrivateKey
	idp             *saml.IdentityProvider
	spProvider      *ServiceProviderProvider
	sessionProvider *SessionProvider
}

// New creates a new IDP server from configuration.
func New(cfg *config.Config) (*Server, error) {
	// Load certificate
	cert, err := cfg.IDP.LoadCertificate()
	if err != nil {
		return nil, err
	}

	// Load private key
	key, err := cfg.IDP.LoadPrivateKey()
	if err != nil {
		return nil, err
	}

	// Create SP provider
	spProvider, err := NewServiceProviderProvider(cfg.ServiceProviders)
	if err != nil {
		return nil, err
	}

	// Parse base URL
	baseURL, err := url.Parse(cfg.Server.BaseURL)
	if err != nil {
		return nil, err
	}

	server := &Server{
		config:      cfg,
		certificate: cert,
		privateKey:  key,
		spProvider:  spProvider,
	}

	// Create session provider (manages pending requests only, no persistent sessions)
	server.sessionProvider = NewSessionProvider()

	// Create SAML IDP
	server.idp = &saml.IdentityProvider{
		Key:         key,
		Certificate: cert,
		MetadataURL: url.URL{
			Scheme: baseURL.Scheme,
			Host:   baseURL.Host,
			Path:   "/metadata",
		},
		SSOURL: url.URL{
			Scheme: baseURL.Scheme,
			Host:   baseURL.Host,
			Path:   "/sso",
		},
		ServiceProviderProvider: spProvider,
		SessionProvider:         server.sessionProvider,
	}

	return server, nil
}

// RegisterRoutes registers HTTP routes for the IDP.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metadata", s.handleMetadata)
	mux.HandleFunc("/sso", s.handleSSO)
	mux.HandleFunc("/login", s.handleLogin)
}

// GetIDP returns the underlying SAML IDP.
func (s *Server) GetIDP() *saml.IdentityProvider {
	return s.idp
}

// GetSPProvider returns the SP provider.
func (s *Server) GetSPProvider() *ServiceProviderProvider {
	return s.spProvider
}

// GetSessionProvider returns the session provider.
func (s *Server) GetSessionProvider() *SessionProvider {
	return s.sessionProvider
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *config.Config {
	return s.config
}
