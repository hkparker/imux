package main

import (
	"flag"
)

var client bool
var binds string
var server bool
var listen string
var dial string
var chunk_size int

func main() {
	flag.BoolVar(&client, "client", false, "create an imux client")
	flag.StringVar(&binds, "binds", "{\"0.0.0.0\": 10}", "JSON encoding of map from bind address strings to int counts")
	flag.BoolVar(&server, "server", false, "create an imux server")
	flag.StringVar(&listen, "listen", "0.0.0.0:443", "listener address and port for clients to imux out and servers to imux in")
	flag.StringVar(&dial, "dial", "127.0.0.1:443", "dial address and port for clients to dial servers and servers to dial out")
	flag.IntVar(&chunk_size, "chunk-size", 16384, "maximum number of bytes per chunk")
	flag.Parse()
	validateFlags()

	if server {
		// create a ManyToOne
	} else if client {
		// create a OneToMany
	}
}

func validateFlags() {

}
