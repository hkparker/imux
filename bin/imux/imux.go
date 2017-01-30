package main

import (
	"encoding/json"
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/imux"
)

var client bool
var binds string
var server bool
var listen string
var dial string
var chunk_size int
var debug bool

func main() {
	flag.BoolVar(&client, "client", false, "create an imux client")
	flag.StringVar(&binds, "binds", "{\"0.0.0.0\": 10}", "JSON encoding of map from bind address strings to int counts")
	flag.BoolVar(&server, "server", false, "create an imux server")
	flag.StringVar(&listen, "listen", "0.0.0.0:443", "listener address and port for clients to imux out and servers to imux in")
	flag.StringVar(&dial, "dial", "127.0.0.1:443", "dial address and port for clients to dial servers and servers to dial out")
	flag.IntVar(&chunk_size, "chunk-size", 16384, "maximum number of bytes per chunk")
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.Parse()
	validateFlags()
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if server {
		imux.ManyToOne(
			createServerListener(listen),
			createDestinationDialer(dial),
		)
	} else if client {
		imux.ClientChunkSize = chunk_size
		bind_map := make(map[string]int)
		err := json.Unmarshal([]byte(binds), &bind_map)
		if err != nil {
		}
		TOFU(dial)
		imux.OneToMany(
			createClientListener(listen),
			bind_map,
			createRedailerGenerator(dial),
		)
	}
}

func validateFlags() {

}
