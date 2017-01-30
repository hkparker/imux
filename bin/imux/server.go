package main

import (
	"net"
)

func createServerListener(listen string) net.Listener {
	return nil
}

func createDestinationDialer(dial string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		return nil, nil
	}
}
