package endpoint

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// RecordType enum values
	RecordTypeA     = "A"
	RecordTypeCNAME = "CNAME"
	RecordTypeTXT   = "TXT"
	RecordTypeSRV   = "SRV"
	RecordTypeNS    = "NS"
	RecordTypePTR   = "PTR"
)

type TTL int64

func (ttl TTL) IsConfigured() bool {
	return ttl > 0
}

// Targets is a representation of a list of targets for an endpoint.
type Targets []string

func NewTargets(target ...string) Targets {
	t := make(Targets, 0, len(target))
	t = append(t, target...)
	return t
}

func (t Targets) String() string {
	return strings.Join(t, ";")
}

func (t Targets) Len() int {
	return len(t)
}

func (t Targets) Less(i, j int) bool {
	return t[i] < t[j]
}

func (t Targets) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t Targets) Same(o Targets) bool {
	if len(t) != len(o) {
		return false
	}
	sort.Stable(t)
	sort.Stable(o)

	for i, e := range t {
		if !strings.EqualFold(e, o[i]) {
			return false
		}
	}
	return true
}

// IsLess compares two targets and choose the 'lesser' one -
// 'less' is the shorter list of targets or where the first entry is less
func (t Targets) IsLess(o Targets) bool {
	if len(t) < len(o) {
		return true
	}
	if len(t) > len(o) {
		return false
	}

	sort.Sort(t)
	sort.Sort(o)

	for i, e := range t {
		if e != o[i] {
			return e < o[i]
		}
	}
	return false
}

// ProviderSpecificProperty holds the name and value of a configuration which is specific to individual DNS providers
type ProviderSpecificProperty struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ProviderSpecific holds configuration which is specific to individual DNS providers
type ProviderSpecific []ProviderSpecificProperty

// Endpoint is a high-level way of a connection between a service and an IP
type Endpoint struct {
	// hostname of the DNS record
	DNSName string  `json:"dnsName,omitempty"`
	Targets Targets `json:"targets,omitempty"`
	// RecordType type of record, e.g. CNAME, A, SRV, TXT etc
	RecordType string `json:"recordType,omitempty"`
	// Identifier to distinguish multiple records with the same name and type (e.g. Route53 records with routing policies other than 'simple')
	SetIdentifier    string           `json:"setIdentifier,omitempty"`
	RecordTTL        TTL              `json:"recordTTL,omitempty"`
	Labels           Labels           `json:"labels,omitempty"`
	ProviderSpecific ProviderSpecific `json:"providerSpecific,omitempty"`
}

// NewEndpoint initialization method to be used to create an endpoint
func NewEndpoint(dnsName, recordType string, targets ...string) *Endpoint {
	return NewEndpointWithTTL(dnsName, recordType, TTL(0), targets...)
}

// NewEndpointWithTTL initialization method to be used to create an endpoint with a TTL struct
func NewEndpointWithTTL(dnsName, recordType string, ttl TTL, targets ...string) *Endpoint {
	cleanTargets := make([]string, len(targets))
	for idx, target := range targets {
		cleanTargets[idx] = strings.TrimSuffix(target, ".")
	}

	return &Endpoint{
		DNSName:    strings.TrimSuffix(dnsName, "."),
		Targets:    cleanTargets,
		RecordType: recordType,
		Labels:     NewLabels(),
		RecordTTL:  ttl,
	}
}

// WithSetIdentifier applies the given set identifier to the endpoint.
func (e *Endpoint) WithSetIdentifier(setIdentifier string) *Endpoint {
	e.SetIdentifier = setIdentifier
	return e
}

// WithProviderSpecific attaches a key/value pair to the Endpoint and returns the Endpoint.
// This can be used to pass additional data through the stages of DNSOps' Endpoint processing.
// The assumption is that most of the time this will be provider specific metadata that doesn't
// warrant its own field on the Endpoint object itself. It differs from Labels in the fact that it's
// not persisted in the Registry but only kept in memory during a single record synchronization.
func (e *Endpoint) WithProviderSpecific(key, value string) *Endpoint {
	if e.ProviderSpecific == nil {
		e.ProviderSpecific = ProviderSpecific{}
	}

	e.ProviderSpecific = append(e.ProviderSpecific, ProviderSpecificProperty{Name: key, Value: value})
	return e
}

func (e *Endpoint) GetProviderSpecificProperty(key string) (ProviderSpecificProperty, bool) {
	for _, providerSpecific := range e.ProviderSpecific {
		if providerSpecific.Name == key {
			return providerSpecific, true
		}
	}
	return ProviderSpecificProperty{}, false
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s %d IN %s %s %s %s", e.DNSName, e.RecordTTL, e.RecordType, e.SetIdentifier, e.Targets, e.ProviderSpecific)
}

// DNSEndpointSpec defines the desired state of DNSEndpoint
type DNSEndpointSpec struct {
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
}

// DNSEndpointStatus defines the observed state of DNSEndpoint
type DNSEndpointStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type DNSEndpoint struct {
	Spec   DNSEndpointSpec   `json:"spec,omitempty"`
	Status DNSEndpointStatus `json:"status,omitempty"`
}

type DNSEndpointList struct {
	Items []DNSEndpoint `json:"items"`
}
