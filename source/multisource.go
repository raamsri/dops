package source

import (
	"context"

	"github.com/toppr-systems/dops/endpoint"
)

// multiSource is a Source that merges the endpoints of its nested Sources.
type multiSource struct {
	children       []Source
	defaultTargets []string
}

// Endpoints collects endpoints of all nested Sources and returns them in a single slice.
func (ms *multiSource) Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error) {
	result := []*endpoint.Endpoint{}

	for _, s := range ms.children {
		endpoints, err := s.Endpoints(ctx)
		if err != nil {
			return nil, err
		}
		if len(ms.defaultTargets) > 0 {
			for i := range endpoints {
				endpoints[i].Targets = ms.defaultTargets
			}
		}
		result = append(result, endpoints...)
	}

	return result, nil
}

func (ms *multiSource) AddEventHandler(ctx context.Context, handler func()) {
	for _, s := range ms.children {
		s.AddEventHandler(ctx, handler)
	}
}

// NewMultiSource creates a new multiSource.
func NewMultiSource(children []Source, defaultTargets []string) Source {
	return &multiSource{children: children, defaultTargets: defaultTargets}
}
