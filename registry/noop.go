package registry

import (
	"context"

	"github.com/toppr-systems/dops/endpoint"
	"github.com/toppr-systems/dops/plan"
	"github.com/toppr-systems/dops/provider"
)

// NoopRegistry implements registry interface without ownership directly propagating changes to dns provider
type NoopRegistry struct {
	provider provider.Provider
}

// NewNoopRegistry returns new NoopRegistry object
func NewNoopRegistry(provider provider.Provider) (*NoopRegistry, error) {
	return &NoopRegistry{
		provider: provider,
	}, nil
}

func (im *NoopRegistry) GetDomainFilter() endpoint.DomainFilterInterface {
	return im.provider.GetDomainFilter()
}

// Records returns the current records from the dns provider
func (im *NoopRegistry) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	return im.provider.Records(ctx)
}

// ApplyChanges propagates changes to the dns provider
func (im *NoopRegistry) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	return im.provider.ApplyChanges(ctx, changes)
}

// PropertyValuesEqual compares two property values for equality
func (im *NoopRegistry) PropertyValuesEqual(attribute string, previous string, current string) bool {
	return im.provider.PropertyValuesEqual(attribute, previous, current)
}

// AdjustEndpoints modifies the endpoints as needed by the specific provider
func (im *NoopRegistry) AdjustEndpoints(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint {
	return im.provider.AdjustEndpoints(endpoints)
}
