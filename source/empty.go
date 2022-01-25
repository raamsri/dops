package source

import (
	"context"

	"github.com/toppr-systems/dops/endpoint"
)

// emptySource is a Source that returns no endpoints.
type emptySource struct{}

func (e *emptySource) AddEventHandler(ctx context.Context, handler func()) {
}

// Endpoints collects endpoints of all nested Sources and returns them in a single slice.
func (e *emptySource) Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error) {
	return []*endpoint.Endpoint{}, nil
}

// NewEmptySource creates a new emptySource.
func NewEmptySource() Source {
	return &emptySource{}
}
