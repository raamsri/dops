package registry

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/endpoint"
	"github.com/toppr-systems/dops/plan"
)

// Registry is an interface which should enables ownership concept in dops
// Records() returns ALL records registered with DNS provider
// each entry includes owner information
// ApplyChanges(changes *plan.Changes) propagates the changes to the DNS Provider API and correspondingly updates ownership depending on type of registry being used
type Registry interface {
	Records(ctx context.Context) ([]*endpoint.Endpoint, error)
	ApplyChanges(ctx context.Context, changes *plan.Changes) error
	PropertyValuesEqual(attribute string, previous string, current string) bool
	AdjustEndpoints(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint
	GetDomainFilter() endpoint.DomainFilterInterface
}

//TODO(ideahitme): consider moving this to Plan
func filterOwnedRecords(ownerID string, eps []*endpoint.Endpoint) []*endpoint.Endpoint {
	filtered := []*endpoint.Endpoint{}
	for _, ep := range eps {
		if endpointOwner, ok := ep.Labels[endpoint.OwnerLabelKey]; !ok || endpointOwner != ownerID {
			log.Debugf(`Skipping endpoint %v because owner id does not match, found: "%s", required: "%s"`, ep, endpointOwner, ownerID)
			continue
		}
		filtered = append(filtered, ep)
	}
	return filtered
}
