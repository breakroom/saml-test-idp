package idp

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/breakroom/saml-test-idp/internal/config"
	"github.com/crewjam/saml"
)

// ServiceProviderProvider implements saml.ServiceProviderProvider.
type ServiceProviderProvider struct {
	mu  sync.RWMutex
	sps map[string]*ServiceProviderEntry
}

// ServiceProviderEntry holds SP metadata and user configuration.
type ServiceProviderEntry struct {
	Metadata *saml.EntityDescriptor
	Config   *config.ServiceProvider
}

// NewServiceProviderProvider creates a new SP provider from config.
func NewServiceProviderProvider(sps []config.ServiceProvider) (*ServiceProviderProvider, error) {
	provider := &ServiceProviderProvider{
		sps: make(map[string]*ServiceProviderEntry),
	}

	for i := range sps {
		sp := &sps[i]
		entry, err := provider.createEntry(sp)
		if err != nil {
			return nil, fmt.Errorf("failed to create SP entry for %s: %w", sp.EntityID, err)
		}
		provider.sps[sp.EntityID] = entry
	}

	return provider, nil
}

func (p *ServiceProviderProvider) createEntry(sp *config.ServiceProvider) (*ServiceProviderEntry, error) {
	var metadata *saml.EntityDescriptor

	if sp.MetadataFile != "" {
		// Load metadata from file (path is resolved relative to config file)
		metadataPath := sp.GetMetadataFilePath()
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata file: %w", err)
		}
		metadata = &saml.EntityDescriptor{}
		if err := xml.Unmarshal(data, metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	} else if sp.ACSURL != "" {
		// Create metadata from ACS URL
		metadata = &saml.EntityDescriptor{
			EntityID: sp.EntityID,
			SPSSODescriptors: []saml.SPSSODescriptor{
				{
					AssertionConsumerServices: []saml.IndexedEndpoint{
						{
							Binding:  saml.HTTPPostBinding,
							Location: sp.ACSURL,
							Index:    1,
						},
					},
				},
			},
		}
	} else {
		return nil, fmt.Errorf("SP must have either acs_url or metadata_file")
	}

	return &ServiceProviderEntry{
		Metadata: metadata,
		Config:   sp,
	}, nil
}

// GetServiceProvider implements saml.ServiceProviderProvider.
func (p *ServiceProviderProvider) GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.sps[serviceProviderID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return entry.Metadata, nil
}

// GetServiceProviderConfig returns the config for an SP.
func (p *ServiceProviderProvider) GetServiceProviderConfig(entityID string) *config.ServiceProvider {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.sps[entityID]
	if !ok {
		return nil
	}
	return entry.Config
}

// GetAllServiceProviders returns all configured SPs.
func (p *ServiceProviderProvider) GetAllServiceProviders() []*config.ServiceProvider {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sps := make([]*config.ServiceProvider, 0, len(p.sps))
	for _, entry := range p.sps {
		sps = append(sps, entry.Config)
	}
	return sps
}
