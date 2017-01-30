package main

import (
	"net"
)

// Create a new listener to accept transport imux sockets
func createServerListener(listen string) net.Listener {
	// Get the TLS certificate to present

	// Create a TLS listener and return it
	return nil
}

// return a function that when called dials the specified address
// and returns the new connection
func createDestinationDialer(dial string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		// Parse the port and address

		// Dial the address with a TCP socket
		return nil, nil
	}
}
