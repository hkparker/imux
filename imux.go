package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/hkparker/imux/lib/client"
	"github.com/hkparker/imux/lib/common"
	"github.com/hkparker/imux/lib/server"
	"github.com/hkparker/imux/lib/session"
	"log"
	"net"
	"os/user"
)

// Client flags
var current_user, _ = user.Current()
var username = flag.String("user", current_user.Username, "username")
var hostname = flag.String("host", "", "imux server to start a session with")
var port = flag.Int("port", 443, "port")
var network_config = flag.String("networks", "0.0.0.0:200", "socket configuration string: <bind ip>:<count>;")
var chunk_size = flag.Int("chunksize", 5*1024*1024, "size of each file chink in byte")

// Server flags
var listen = flag.String("bind", "0.0.0.0", "address to bind an imux server on")
var daemon = flag.Bool("daemon", false, "run in the background, only used in server mode")
var cert = flag.String("cert", "cert.pem", "pem file with certificate to present when in server mode") // ~/.imux/cert.pem
var key = flag.String("key", "key.pem", "pem file with key for certificate presented in server mode")  // ~/.imux/key.pem

//Session flags
var ipc_file = flag.String("session", "", "unix socket used for IPC, set by server for user-context processes")

func runClient() {
	networks := common.ParseNetworks(*network_config)
	tlj_client, err := common.CreateClient(*hostname, *port) // client
	if err != nil {
		log.Fatal(err)
	}

	nonce := common.ClientLogin(*username, tlj_client) // client
	if err != nil {
		log.Fatal(err)
	}

	streamers, err := common.ConnectWorkers(*hostname, *port, networks, nonce) // client
	if err != nil {
		log.Fatal(err)
	}

	go client.CommandLoop(tlj_client, streamers, *chunk_size)
	err = <-tlj_client.Dead
	fmt.Println("control connection closed:", err)
}

func runSever() {
	if current_user, _ := user.Current(); current_user.Uid != "0" {
		log.Fatal("Server must run as root.")
	}

	log.Println("starting imux server")

	config := common.PrepareTLSConfig(*cert, *key) // server
	address := fmt.Sprintf("%s:%d", *listen, *port)
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		log.Fatal("error starting server:", err)
	}

	server := server.NewTLJServer(listener)
	if *daemon {
		// Daemonize
	}
	err = <-server.FailedServer
	log.Println("server closed: %s", err)
}

func runSession() {
	control_socket, err := net.Dial("unix", *ipc_file)
	if err != nil {
		fmt.Println(err)
		return
	}

	discard_listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return
	}
	server := session.NewTLJServer(discard_listener)
	server.Insert(control_socket)
	server.TagSocket(control_socket, "peer")
	err = <-server.FailedServer
	fmt.Println(err)
}

func main() {
	//flag.Parse()
	// if session
	// 	runSession()
	// else if bind specified
	//	runServer()
	// else if host specified
	//	runClient()
}
