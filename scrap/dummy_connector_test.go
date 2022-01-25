package scrap

import (
	"context"
	"encoding/gob"
	"net"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/toppr-systems/dops/source"
)

func TestDummyConnector(t *testing.T) {
	// Listen on TCP port 9876 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", ":9876")
	if err != nil {
		log.Fatal(err)
	}
	log.Infoln("listening...")
	defer l.Close()

	dnsName := "dops2.toppr.systems"
	dummy, err := source.NewDummySource(dnsName)
	if err != nil {
		log.Errorf("error creating a dummy source: %v", err)
	}
	log.Infoln("dummy Source created.")
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		encoder := gob.NewEncoder(conn)
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn, enc *gob.Encoder) {
			endpoints, err := dummy.Endpoints(context.TODO())
			if err != nil {
				log.Errorf("error fetching endpoints: %v", err)
			}
			log.Infof("endpoints: %v", endpoints)
			err = enc.Encode(endpoints)
			if err != nil {
				log.Errorf("error encoding endpoints: %v", err)
			}
			log.Infoln("endpoints encoded")

			// Shut down the connection.
			c.Close()
		}(conn, encoder)
	}
}
