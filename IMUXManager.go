package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"crypto/x509"
	"io/ioutil"
	)

type IMUXManager struct {
	PeerAddr string
	Sockets chan tls.Conn
	Workers = map[string]IMUXSocket
	// when a new worker is created give it a uuid and assign it to the hash table
}

func (manager *IMUXManager) CreateSockets() {
	// This function is ran in a goroutine and populates manager.Sockets by creating sockets

	// Cert comes from control socket (same cert for all tls sockets) so no need to download cert here
	manager.Sockets = make(chan tls.Conn)
	for {
		// create a socket to PeerAddr
			// need to know which IP to bind to
		// put it in the channel
	}
}

func (manager *IMUXManager) RecieveSockets() {
	// This function is ran in a goroutine and populates manager.Sockets by recieving sockets
	// error reporting?
	// perhaps it accepts a listening server, so any errors would be caught before
	
	// Parse the CA and create certificate pool
	ca_file, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		 // report error
	}
	ca, err := x509.ParseCertificate(ca_file)
	if err != nil {
		 // report error
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	
	// Parse certificate and create structs
	priv_file, err := ioutil.ReadFile("ca.key")
	if err != nil {
		 // report error
	}
	priv, err := x509.ParsePKCS1PrivateKey(priv_file)
	if err != nil {
		 // report error
	}
	cert := tls.Certificate{
		Certificate: [][]byte{ ca_file },
		PrivateKey: priv,
	}
	config := tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs: pool,
	}
	config.Rand = rand.Reader
	
	// Create the server and send sockets to the channel
	service := "0.0.0.0:443"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		// report error
	}
	manager.Sockets = make(chan tls.Conn)
	for {
		// recieve a socket from that server
		// put it in the chan
	}
}

func (manager *IMUXManager) IncreaseSockets(bind_ips map[string]int) int {
	return workers_created
}

func (manager *IMUXManager) DecreaseSockets() {
	
}

func (manager *IMUXManager) ServeFile(filename string, starting_position int) {
	// Create ReadQueue from file
	// set starting position
	// for each manager.Workers
	//   go serve the read queue
	// wait for done from each imuxsocket
	// report errors as they occur? (they are the done channel)
}

func (manager *IMUXManager) DownloadFile() {
	// Create a buffer for the file
	// for each manager.Workers
	//   go download to the buffer
	// wait for done from each imuxsocket
	// report errors as they occur (they are the done channel)
}

func (manager *IMUXManager) UpdateRecycling(state bool) int {
	// for each imux socket
		imuxsocket.Recycle = state
	// end for
}

func (manager *IMUXManager) UpdateChunkSize() int {	// need to interact with ReadQueue involved with transfer
	return 0
}

func (manager *IMUXManager) CurrentSpeed() float64 {
	var speed float64 = 0.0
	// for each imux socket
		speed += imuxsocket.LastSpeed
	// end for
	return speed
}

func (manager *IMUXManager) Close() int {
	return 0
}
