package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"io"
	"net"
	"reflect"
)

var server_write_queues = make(map[string]WriteQueue)
var responders = make(map[string]DataIMUX)
var loopers = make(map[net.Conn]bool)

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
			queue, present := server_write_queues[chunk.SocketID]
			if !present {
				destination, err := dial_destination()
				if err != nil {
				}
				queue = WriteQueue{
					Destination: io.Writer(destination),
				}
				server_write_queues[chunk.SocketID] = queue
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
