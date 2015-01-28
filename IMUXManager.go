package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"crypto/x509"
	"io/ioutil"
	"github.com/twinj/uuid"
	)

type IMUXManager struct {
	PeerAddr string
	Sockets chan tls.Conn
	Workers = map[string]IMUXSocket
}

func (manager *IMUXManager) CreateSockets() {
	// This function is ran in a goroutine and populates manager.Sockets by creating sockets
	
	// perhaps a goroutine for each bind_ip (dialer struct) is created.  When more sockets are required the corresponding channel is used
	
	manager.Sockets = make(chan tls.Conn)
	for {
		tls.DialWithDialer()// create a socket to PeerAddr
			// need to know which IP to bind to
		manager.Sockets <- socket
	}
}

//		RecieveSockets setup code:

//	// Parse the CA and create certificate pool
//	ca_file, err := ioutil.ReadFile("ca.pem")
//	if err != nil {
//	 	// report error
//	}
//	ca, err := x509.ParseCertificate(ca_file)
//	if err != nil {
//	 	// report error
//	}
//	pool := x509.NewCertPool()
//	pool.AddCert(ca)
//
//	// Parse certificate and create structs
//	priv_file, err := ioutil.ReadFile("ca.key")
//	if err != nil {
//	 	// report error
//	}
//	priv, err := x509.ParsePKCS1PrivateKey(priv_file)
//	if err != nil {
//	 	// report error
//	}
//	cert := tls.Certificate{
//		Certificate: [][]byte{ ca_file },
//		PrivateKey: priv,
//	}
//	config := tls.Config{
//		ClientAuth: tls.RequireAndVerifyClientCert,
//		Certificates: []tls.Certificate{cert},
//		ClientCAs: pool,
//	}
//	config.Rand = rand.Reader
//
//	// Create the server and send sockets to the channel
//	service := "0.0.0.0:443"
//	listener, err := tls.Listen("tcp", service, &config)
//	if err != nil {
//		// report error
//	}

func (manager *IMUXManager) RecieveSockets(server *net.Listener, max_errors int) {
	// This function is ran in a goroutine and populates manager.Sockets by recieving sockets
	manager.Sockets = make(chan tls.Conn)
	errors := 0
	for {
		if errors >= max_errors {
			break
		}
		socket, err := server.Accept()
		if err == nil {
			manager.Sockets <- socket	// any need to use uuid to check correct imux session?  Or use specific port?
		} else {
			errors += 1
		}
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
	for _, imuxsocket := range manager.Workers {
		imuxsocket.Upload(read_queue, done)
	}
	// wait for done from each imuxsocket
	// report errors as they occur? (they are the done channel)
}

func (manager *IMUXManager) DownloadFile() {
	// Create a buffer for the file
	for _, imuxsocket := range manager.Workers {
		imuxsocket.Download(buffer, done)
	}
	// wait for done from each imuxsocket
	// report errors as they occur (they are the done channel)
}

func (manager *IMUXManager) UpdateRecycling(state bool) {
	for _, imuxsocket := range manager.Workers {
		imuxsocket.Recycle = state
	}
}

func (manager *IMUXManager) CurrentSpeed() float64 {
	var speed float64 = 0.0
	for _, imuxsocket := range manager.Workers {
		speed += imuxsocket.LastSpeed
	}
	return speed
}

func (manager *IMUXManager) Close() int {
	err_count := 0
	for _, imuxsocket := range manager.Workers {
		err := imuxsocket.Close()
		if err != nil {
			err_count += 1
		}
	}
	return err_count
}
