package endpoint

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrInvalidOrigin is returned when origin was not found, or different origin is found
	ErrInvalidOrigin = errors.New("origin is unknown or not found")
)

const (
	origin = "dops"
	// OwnerLabelKey is the name of the label that defines the owner of an Endpoint.
	OwnerLabelKey = "owner"
	// ResourceLabelKey is the name of the label that identifies resource which acquires the DNS name
	ResourceLabelKey = "resource"

	// DualstackLabelKey is the name of the label that identifies dualstack endpoints
	DualstackLabelKey = "dualstack"
)

// Labels store metadata related to the endpoint
// it is then stored in a persistent storage via serialization
type Labels map[string]string

// NewLabels returns empty Labels
func NewLabels() Labels {
	return map[string]string{}
}

// NewLabelsFromString constructs endpoints labels from a provided format string
// if origin set to another value is found then error is returned
// no origin automatically assumes is not owned by dops and returns invalidOrigin error
func NewLabelsFromString(labelText string) (Labels, error) {
	endpointLabels := map[string]string{}
	labelText = strings.Trim(labelText, "\"")
	tokens := strings.Split(labelText, ",")
	foundExternalDNSOrigin := false
	for _, token := range tokens {
		if len(strings.Split(token, "=")) != 2 {
			continue
		}
		key := strings.Split(token, "=")[0]
		val := strings.Split(token, "=")[1]
		if key == "origin" && val != origin {
			return nil, ErrInvalidOrigin
		}
		if key == "origin" {
			foundExternalDNSOrigin = true
			continue
		}
		if strings.HasPrefix(key, origin) {
			endpointLabels[strings.TrimPrefix(key, origin+"/")] = val
		}
	}

	if !foundExternalDNSOrigin {
		return nil, ErrInvalidOrigin
	}

	return endpointLabels, nil
}

// Serialize transforms endpoints labels into a dops recognizable format string
// withQuotes adds additional quotes
func (l Labels) Serialize(withQuotes bool) string {
	var tokens []string
	tokens = append(tokens, fmt.Sprintf("origin=%s", origin))
	var keys []string
	for key := range l {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		tokens = append(tokens, fmt.Sprintf("%s/%s=%s", origin, key, l[key]))
	}
	if withQuotes {
		return fmt.Sprintf("\"%s\"", strings.Join(tokens, ","))
	}
	return strings.Join(tokens, ",")
}
