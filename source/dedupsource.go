package source

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/endpoint"
)

// dedupSource is a Source that removes duplicate endpoints from its wrapped source.
type dedupSource struct {
	source Source
}

// NewDedupSource creates a new dedupSource wrapping the provided Source.
func NewDedupSource(source Source) Source {
	return &dedupSource{source: source}
}

// Endpoints collects endpoints from its wrapped source and returns them without duplicates.
func (ms *dedupSource) Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error) {
	result := []*endpoint.Endpoint{}
	collected := map[string]bool{}

	endpoints, err := ms.source.Endpoints(ctx)
	if err != nil {
		return nil, err
	}

	for _, ep := range endpoints {
		identifier := ep.DNSName + " / " + ep.SetIdentifier + " / " + ep.Targets.String()

		if _, ok := collected[identifier]; ok {
			log.Debugf("Removing duplicate endpoint %s", ep)
			continue
		}

		collected[identifier] = true
		result = append(result, ep)
	}

	return result, nil
}

func (ms *dedupSource) AddEventHandler(ctx context.Context, handler func()) {
	ms.source.AddEventHandler(ctx, handler)
}
