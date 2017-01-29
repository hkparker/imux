package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"io"
	"net"
	"reflect"
)

var WriteQueues = make(map[string]WriteQueue)
var responders = make(map[string]DataIMUX)
var loopers = make(map[net.Conn]bool)

// Provide a net.Listener, for which any accepted sockets will have their data
// inverse multiplexed to
func OneToMany(listener net.Listener, config IMUXConfig, redialer RedialerGenerator) error {
	// Create a new SessionID shared by all sockets accepted by this OneToMany
	session_id := "" //uuid4

	// Create a new DataIMUX to read data from accepted connections and
	// chunk all data.  Create IMUXSockets to read chunks from the DataIMUX
	// and write them to connections to the server.
	imuxer := NewDataIMUX(session_id)
	for bind, count := range config.Binds {
		for i := 0; i < count; i++ {
			imux_socket := IMUXSocket{
				IMUXer:   imuxer,
				Redialer: redialer(bind),
			}
			imux_socket.init()
		}
	}

	// In an infinite loop, accept new connections to this listener
	// and read data into the session DataIMUX.  Create a WriteQueue
	// to write out return chunks.
	for {
		socket, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{}).Error(err)
			return err
		}
		socket_id := "" //uuid4

		// Create a new WriteQueue addressed by the socket ID to
		// take return chunks and write them into this socket
		WriteQueues[socket_id] = WriteQueue{
			Destination: io.Writer(socket),
		}

		go imuxer.ReadFrom(socket_id, socket, "client")
	}
}

// Create a new TLJ server to accept chunks from anywhere
// and order them, writing them to corresponding sockets.
func ManyToOne(listener net.Listener, dial_destination func() (net.Conn, error)) {
	tlj_server := tlj.NewServer(listener, tag_socket, type_store())
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(Chunk); ok {
			if _, present := responders[chunk.SessionID]; !present {
				responders[chunk.SessionID] = NewDataIMUX(chunk.SessionID)
			}
			if _, looping := loopers[context.Socket]; !looping {
				go func() {
					writer, err := tlj.NewStreamWriter(context.Socket, type_store(), reflect.TypeOf(Chunk{}))
					if err != nil {
					}
					for {
						new_chunk := <-responders[chunk.SessionID].Chunks
						err := writer.Write(new_chunk)
						if err != nil {
							responders[chunk.SessionID].Stale <- new_chunk
							break
						}
					}
				}()
				loopers[context.Socket] = true
			}
			queue, present := WriteQueues[chunk.SocketID]
			if !present {
				destination, err := dial_destination()
				if err != nil {
				}
				queue = WriteQueue{
					Destination: io.Writer(destination),
				}
				WriteQueues[chunk.SocketID] = queue
				if imuxer, ok := responders[chunk.SessionID]; ok {
					imuxer.ReadFrom(chunk.SocketID, destination, "server")
				}
			}
			queue.Write(chunk)
		}
	})

	err := <-tlj_server.FailedServer
	log.WithFields(log.Fields{
		"error": err.Error(),
	}).Error()
}
