package source

import (
	"context"
	"encoding/gob"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/endpoint"
)

const (
	dialTimeout = 30 * time.Second
)

// connectorSource is an implementation of Source that provides endpoints by connecting
// to a remote tcp server. The encoding/decoding is done using encoder/gob package.
type connectorSource struct {
	remoteServer string
}

// NewConnectorSource creates a new connectorSource with the given config.
func NewConnectorSource(remoteServer string) (Source, error) {
	return &connectorSource{
		remoteServer: remoteServer,
	}, nil
}

// Endpoints returns endpoint objects.
func (cs *connectorSource) Endpoints(ctx context.Context) ([]*endpoint.Endpoint, error) {
	endpoints := []*endpoint.Endpoint{}

	conn, err := net.DialTimeout("tcp", cs.remoteServer, dialTimeout)
	if err != nil {
		log.Errorf("Connection error: %v", err)
		return nil, err
	}
	defer conn.Close()

	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&endpoints); err != nil {
		log.Errorf("Decode error: %v", err)
		return nil, err
	}

	log.Debugf("Received endpoints: %#v", endpoints)

	return endpoints, nil
}

func (cs *connectorSource) AddEventHandler(ctx context.Context, handler func()) {
}
