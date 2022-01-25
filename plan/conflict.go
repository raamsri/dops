package plan

import (
	"sort"

	"github.com/toppr-systems/dops/endpoint"
)

// ConflictResolver is used to make a decision in case of two or more resources
// are trying to acquire same DNS name
type ConflictResolver interface {
	ResolveCreate(candidates []*endpoint.Endpoint) *endpoint.Endpoint
	ResolveUpdate(current *endpoint.Endpoint, candidates []*endpoint.Endpoint) *endpoint.Endpoint
}

// PerResource allows only one resource to own a given dns name
type PerResource struct{}

// ResolveCreate is invoked when dns name is not owned by any resource
// ResolveCreate takes "minimal" (string comparison of Target) endpoint to acquire the DNS record
func (s PerResource) ResolveCreate(candidates []*endpoint.Endpoint) *endpoint.Endpoint {
	var min *endpoint.Endpoint
	for _, ep := range candidates {
		if min == nil || s.less(ep, min) {
			min = ep
		}
	}
	return min
}

// ResolveUpdate is invoked when dns name is already owned by "current" endpoint
// ResolveUpdate uses "current" record as base and updates it accordingly with new version of same resource
// if it doesn't exist then pick min
func (s PerResource) ResolveUpdate(current *endpoint.Endpoint, candidates []*endpoint.Endpoint) *endpoint.Endpoint {
	currentResource := current.Labels[endpoint.ResourceLabelKey] // resource which has already acquired the DNS
	// TODO: sort candidates only needed because we can still have two endpoints from same resource here. We sort for consistency
	// TODO: remove once single endpoint can have multiple targets
	sort.SliceStable(candidates, func(i, j int) bool {
		return s.less(candidates[i], candidates[j])
	})
	for _, ep := range candidates {
		if ep.Labels[endpoint.ResourceLabelKey] == currentResource {
			return ep
		}
	}
	return s.ResolveCreate(candidates)
}

// less returns true if endpoint x is less than y
func (s PerResource) less(x, y *endpoint.Endpoint) bool {
	return x.Targets.IsLess(y.Targets)
}

// TODO: with cross-resource/cross-cluster setup alternative variations of ConflictResolver can be used
