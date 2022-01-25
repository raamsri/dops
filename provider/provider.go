package provider

import (
	"context"
	"net"
	"strings"

	"github.com/toppr-systems/dops/endpoint"
	"github.com/toppr-systems/dops/plan"
)

// Provider defines the interface DNS providers should implement.
type Provider interface {
	Records(ctx context.Context) ([]*endpoint.Endpoint, error)
	ApplyChanges(ctx context.Context, changes *plan.Changes) error
	PropertyValuesEqual(name string, previous string, current string) bool
	AdjustEndpoints(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint
	GetDomainFilter() endpoint.DomainFilterInterface
}

type BaseProvider struct {
}

func (b BaseProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint {
	return endpoints
}

func (b BaseProvider) PropertyValuesEqual(name, previous, current string) bool {
	return previous == current
}

func (b BaseProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return endpoint.DomainFilter{}
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string { return "provider context value " + k.name }

// RecordsContextKey is a context key. It can be used during ApplyChanges
// to access previously cached records. The associated value will be of
// type []*endpoint.Endpoint.
var RecordsContextKey = &contextKey{"records"}

// EnsureTrailingDot ensures that the hostname receives a trailing dot if it hasn't already.
func EnsureTrailingDot(hostname string) string {
	if net.ParseIP(hostname) != nil {
		return hostname
	}

	return strings.TrimSuffix(hostname, ".") + "."
}

// Difference tells which entries need to be respectively
// added, removed, or left untouched for "current" to be transformed to "desired"
func Difference(current, desired []string) ([]string, []string, []string) {
	add, remove, leave := []string{}, []string{}, []string{}
	index := make(map[string]struct{}, len(current))
	for _, x := range current {
		index[x] = struct{}{}
	}
	for _, x := range desired {
		if _, found := index[x]; found {
			leave = append(leave, x)
			delete(index, x)
		} else {
			add = append(add, x)
			delete(index, x)
		}
	}
	for x := range index {
		remove = append(remove, x)
	}
	return add, remove, leave
}
