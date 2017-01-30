package main

import (
	"crypto/tls"
	log "github.com/Sirupsen/logrus"
	"net"
)

// Create a new listener to accept transport imux sockets
func createServerListener(listen string) net.Listener {
	certificate := serverTLSCert(listen)
	listener, err := tls.Listen(
		"tcp",
		listen,
		&tls.Config{
			Certificates: []tls.Certificate{
				certificate,
			},
		},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "createServerListener",
			"bind":  listen,
			"error": err.Error(),
		}).Fatal("unable to start server listener")
	}
	return listener
}

// Return a function that when called dials the specified address
// and returns the new connection
func createDestinationDialer(dial string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		return net.Dial("tcp", dial)
	}
}
