package source

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/toppr-systems/dops/endpoint"
)

// dummySource is an implementation of Source that provides dummy endpoints for
// testing/dry-running dns providers.
type dummySource struct {
	dnsNames []string
}

const (
	defaultFQDNTemplate = "example.com"
	hostPrefix          = "dummy-"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDummySource(fqdnTemplate string) (Source, error) {
	var names []string
	if fqdnTemplate == "" {
		names = append(names, defaultFQDNTemplate)
	} else {
		names = strings.Split(strings.ReplaceAll(fqdnTemplate, " ", ""), ",")
	}

	return &dummySource{
		dnsNames: names,
	}, nil
}

func (sc *dummySource) AddEventHandler(ctx context.Context, handler func()) {
}

func (sc *dummySource) Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error) {
	endpoints := make([]*endpoint.Endpoint, 0)

	for _, name := range sc.dnsNames {
		for i := 0; i < 5; i++ {
			ep, _ := generateEndpoint(name)
			endpoints = append(endpoints, ep)
		}
	}

	return endpoints, nil
}

func generateEndpoint(dnsName string) (*endpoint.Endpoint, error) {
	ep := endpoint.NewEndpoint(
		generateDNSName(4, dnsName),
		endpoint.RecordTypeA,
		generateIPAddress(),
	)

	return ep, nil
}

func generateIPAddress() string {
	// 192.0.2.[1-255] is reserved by RFC 5737 for documentation and examples
	return net.IPv4(
		byte(192),
		byte(0),
		byte(2),
		byte(rand.Intn(253)+1),
	).String()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func generateDNSName(prefixLength int, dnsName string) string {
	prefixBytes := make([]rune, prefixLength)

	for i := range prefixBytes {
		prefixBytes[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	prefixStr := string(prefixBytes)

	return fmt.Sprintf("%s%s.%s", hostPrefix, prefixStr, dnsName)
}
