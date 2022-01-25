package source

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/toppr-systems/dops/endpoint"
)

// Provider-specific annotations
const (
	// The annotation to determine whether traffic will go through Cloudflare
	CloudflareProxiedKey = "dops/cloudflare-proxied"

	SetIdentifierKey = "dops/set-identifier"
)

const (
	ttlMinimum = 1
	ttlMaximum = math.MaxInt32
)

// Source defines the interface Endpoint sources should implement.
type Source interface {
	Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error)
	// AddEventHandler adds an event handler that should be triggered if something in source changes
	AddEventHandler(context.Context, func())
}

// parseTTL parses TTL from string, returning duration in seconds.
// parseTTL supports both integers like "600" and durations based
// on Go Duration like "10m", hence "600" and "10m" represent the same value.
//
// Note: for durations like "1.5s" the fraction is omitted (resulting in 1 second
// for the example).
func parseTTL(s string) (ttlSeconds int64, err error) {
	ttlDuration, err := time.ParseDuration(s)
	if err != nil {
		return strconv.ParseInt(s, 10, 64)
	}

	return int64(ttlDuration.Seconds()), nil
}

func parseTemplate(fqdnTemplate string) (tmpl *template.Template, err error) {
	if fqdnTemplate == "" {
		return nil, nil
	}
	funcs := template.FuncMap{
		"trimPrefix": strings.TrimPrefix,
	}
	return template.New("endpoint").Funcs(funcs).Parse(fqdnTemplate)
}

// suitableType returns the DNS resource record type suitable for the target.
// In this case type A for IPs and type CNAME for everything else.
func suitableType(target string) string {
	if net.ParseIP(target) != nil {
		return endpoint.RecordTypeA
	}
	return endpoint.RecordTypeCNAME
}

// endpointsForHostname returns the endpoint objects for each host-target combination.
func endpointsForHostname(hostname string, targets endpoint.Targets, ttl endpoint.TTL, providerSpecific endpoint.ProviderSpecific, setIdentifier string) []*endpoint.Endpoint {
	var endpoints []*endpoint.Endpoint

	var aTargets endpoint.Targets
	var cnameTargets endpoint.Targets

	for _, t := range targets {
		switch suitableType(t) {
		case endpoint.RecordTypeA:
			aTargets = append(aTargets, t)
		default:
			cnameTargets = append(cnameTargets, t)
		}
	}

	if len(aTargets) > 0 {
		epA := &endpoint.Endpoint{
			DNSName:          strings.TrimSuffix(hostname, "."),
			Targets:          aTargets,
			RecordTTL:        ttl,
			RecordType:       endpoint.RecordTypeA,
			Labels:           endpoint.NewLabels(),
			ProviderSpecific: providerSpecific,
			SetIdentifier:    setIdentifier,
		}
		endpoints = append(endpoints, epA)
	}

	if len(cnameTargets) > 0 {
		epCNAME := &endpoint.Endpoint{
			DNSName:          strings.TrimSuffix(hostname, "."),
			Targets:          cnameTargets,
			RecordTTL:        ttl,
			RecordType:       endpoint.RecordTypeCNAME,
			Labels:           endpoint.NewLabels(),
			ProviderSpecific: providerSpecific,
			SetIdentifier:    setIdentifier,
		}
		endpoints = append(endpoints, epCNAME)
	}

	return endpoints
}

type eventHandlerFunc func()

func (fn eventHandlerFunc) OnAdd(obj interface{})               { fn() }
func (fn eventHandlerFunc) OnUpdate(oldObj, newObj interface{}) { fn() }
func (fn eventHandlerFunc) OnDelete(obj interface{})            { fn() }
