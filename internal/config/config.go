// Package config provides configuration structures and loading for the SAML IDP.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the full application configuration.
type Config struct {
	Server           ServerConfig      `yaml:"server"`
	IDP              IDPConfig         `yaml:"idp"`
	ServiceProviders []ServiceProvider `yaml:"service_providers"`

	// baseDir is the directory containing the config file, used for resolving relative paths
	baseDir string
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

// IDPConfig contains the Identity Provider settings.
type IDPConfig struct {
	EntityID        string `yaml:"entity_id"`
	Certificate     string `yaml:"certificate"`
	CertificatePath string `yaml:"certificate_path"`
	PrivateKey      string `yaml:"private_key"`
	PrivateKeyPath  string `yaml:"private_key_path"`

	// baseDir is inherited from Config for resolving relative paths
	baseDir string
}

// ServiceProvider represents a configured SP with its users.
type ServiceProvider struct {
	EntityID     string `yaml:"entity_id"`
	ACSURL       string `yaml:"acs_url"`
	MetadataFile string `yaml:"metadata_file"`
	NameIDFormat string `yaml:"name_id_format"`
	Users        []User `yaml:"users"`

	// baseDir is inherited from Config for resolving relative paths
	baseDir string
}

// User represents a test user with attributes.
type User struct {
	Name       string                 `yaml:"name"`
	NameID     string                 `yaml:"name_id"`
	Attributes map[string]interface{} `yaml:"attributes"`
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Store the config file's directory for resolving relative paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of config file: %w", err)
	}
	cfg.baseDir = filepath.Dir(absPath)

	// Propagate baseDir to IDP config
	cfg.IDP.baseDir = cfg.baseDir

	// Propagate baseDir to service providers
	for i := range cfg.ServiceProviders {
		cfg.ServiceProviders[i].baseDir = cfg.baseDir
	}

	// Set defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	// Set default Name ID format for SPs
	for i := range cfg.ServiceProviders {
		if cfg.ServiceProviders[i].NameIDFormat == "" {
			cfg.ServiceProviders[i].NameIDFormat = "email"
		}
	}

	return &cfg, nil
}

// resolvePath resolves a path relative to the config file's directory.
// If the path is absolute, it is returned unchanged.
func resolvePath(baseDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// LoadCertificate loads the IDP certificate from config (inline or file path).
func (c *IDPConfig) LoadCertificate() (*x509.Certificate, error) {
	var pemData []byte
	var err error

	if c.Certificate != "" {
		pemData = []byte(c.Certificate)
	} else if c.CertificatePath != "" {
		certPath := resolvePath(c.baseDir, c.CertificatePath)
		pemData, err = os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no certificate provided (use certificate or certificate_path)")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block for certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// LoadPrivateKey loads the IDP private key from config (inline or file path).
func (c *IDPConfig) LoadPrivateKey() (*rsa.PrivateKey, error) {
	var pemData []byte
	var err error

	if c.PrivateKey != "" {
		pemData = []byte(c.PrivateKey)
	} else if c.PrivateKeyPath != "" {
		keyPath := resolvePath(c.baseDir, c.PrivateKeyPath)
		pemData, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no private key provided (use private_key or private_key_path)")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block for private key")
	}

	// Try PKCS1 first, then PKCS8
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an RSA key")
		}
	}

	return key, nil
}

// GetMetadataFilePath returns the resolved metadata file path.
func (sp *ServiceProvider) GetMetadataFilePath() string {
	return resolvePath(sp.baseDir, sp.MetadataFile)
}

// GetUserByName finds a user by name in a service provider's user list.
func (sp *ServiceProvider) GetUserByName(name string) *User {
	for i := range sp.Users {
		if sp.Users[i].Name == name {
			return &sp.Users[i]
		}
	}
	return nil
}
