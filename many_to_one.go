package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
	"sync"
)

// Chan of failures to write to destinations
var FailedSocketOuts = make(map[string]chan bool)

// WriteQueues for each outgoing socket on the server
var server_write_queues = make(map[string]*WriteQueue)
var swqMux sync.Mutex

// DataIMUX objects to read responses from each outgoing destination socket
var responders = make(map[string]DataIMUX)
var respondersMux sync.Mutex

// Tracks if goroutines have been created for each socket to read from the
// DataIMUXer for its session and write responses down
var loopers = make(map[net.Conn]bool)
var loopersMux sync.Mutex

// Create a new TLJ server to accept chunks from anywhere and order them, writing them to corresponding sockets
func ManyToOne(listener net.Listener, dial_destination Redialer) {
	tlj_server := tlj.NewServer(listener, tag_socket, type_store())
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(*Chunk); ok {
			log.WithFields(log.Fields{
				"at":          "ManyToOne",
				"sequence_id": chunk.SequenceID,
				"socket_id":   chunk.SocketID,
				"session_id":  chunk.SessionID,
			}).Debug("received chunk")
			createFailReporterIfNeeded(chunk.SocketID, chunk.SessionID)
			createResponderIMUXIfNeeded(chunk.SessionID)
			writeResponseChunksIfNeeded(context.Socket, chunk.SessionID)
			queue, err := queueForDestinationDialIfNeeded(chunk.SocketID, chunk.SessionID, dial_destination)
			if err == nil {
				queue.Chunks <- chunk
				log.WithFields(log.Fields{
					"at":          "ManyToOne",
					"sequence_id": chunk.SequenceID,
					"socket_id":   chunk.SocketID,
					"session_id":  chunk.SessionID,
				}).Debug("wrote chunk")
			} else {
				log.WithFields(log.Fields{
					"at":          "ManyToOne",
					"error":       err.Error(),
					"sequence_id": chunk.SequenceID,
					"socket_id":   chunk.SocketID,
					"session_id":  chunk.SessionID,
				}).Error("dropped chunk")
				if reporter, ok := FailedSocketOuts[chunk.SocketID]; ok {
					reporter <- true
				} else {
					log.WithFields(log.Fields{
						"at":        "ManyToOne",
						"socket_id": chunk.SocketID,
					}).Error("unable to lookup fail socket out channel")
				}
			}
		}
	})

	log.WithFields(log.Fields{
		"at": "ManyToOne",
	}).Debug("created new ManyToOne")
	err := <-tlj_server.FailedServer
	log.WithFields(log.Fields{
		"error": err.Error(),
	}).Error("TLJ server failed for ManyToOne")
}

// Create a chan if needed to pass events about this socket failing
func createFailReporterIfNeeded(socket_id, session_id string) {
	if _, present := FailedSocketOuts[socket_id]; !present {
		FailedSocketOuts[socket_id] = make(chan bool, 0)
		go func() {
			for {
				<-FailedSocketOuts[socket_id]
				responders[session_id].Chunks <- Chunk{
					SessionID:  session_id,
					SocketID:   socket_id,
					SequenceID: 1,
					Close:      true,
				}
			}
		}()
	}
}

// If it does not exist, create a DataIMUX to read data from
// outgoing destination sockets with a common session
func createResponderIMUXIfNeeded(session_id string) {
	respondersMux.Lock()
	if _, present := responders[session_id]; !present {
		responders[session_id] = NewDataIMUX(session_id)
		log.WithFields(log.Fields{
			"at":         "createResponderIMUXIfNeeded",
			"session_id": session_id,
		}).Debug("created new responder imux for session")
	}
	respondersMux.Unlock()
}

// If it is not already happening, ensure that response chunks for a specified
// session_id are written back down this socket.
func writeResponseChunksIfNeeded(socket net.Conn, session_id string) {
	loopersMux.Lock()
	if _, looping := loopers[socket]; !looping {
		log.WithFields(log.Fields{
			"at":         "writeResponseChunksIfNeeded",
			"session_id": session_id,
		}).Debug("creating write back routine for socket")
		go func() {
			writer, err := tlj.NewStreamWriter(socket, type_store(), reflect.TypeOf(Chunk{}))
			if err != nil {
				log.WithFields(log.Fields{
					"at":         "writeResponseChunksIfNeeded",
					"session_id": session_id,
					"error":      err.Error(),
				}).Error("error create return stream writer")
				return
			}
			respondersMux.Lock()
			chunk_stream, ok := responders[session_id]
			if !ok {
				log.WithFields(log.Fields{
					"at":         "writeResponseChunksIfNeeded",
					"session_id": session_id,
				}).Error("error looking up responder imux for session_id")
				return
			}
			respondersMux.Unlock()
			for {
				new_chunk := <-chunk_stream.Chunks
				err := writer.Write(new_chunk)
				if err != nil {
					responders[session_id].Stale <- new_chunk
					log.WithFields(log.Fields{
						"at":         "writeResponseChunksIfNeeded",
						"session_id": session_id,
						"data_len":   len(new_chunk.Data),
						"error":      err.Error(),
					}).Error("error writing a chunk down transport socket")
					break
				} else {
					log.WithFields(log.Fields{
						"at":         "writeResponseChunksIfNeeded",
						"session_id": session_id,
						"data_len":   len(new_chunk.Data),
					}).Debug("wrote a chunk down transport socket")
				}
			}
		}()
		loopers[socket] = true
	}
	loopersMux.Unlock()
}

// Get the queue a new chunk should go to, dialing the outgoing destination socket if this is the first time
// a socket ID has been observed.
func queueForDestinationDialIfNeeded(socket_id, session_id string, dial_destination func() (net.Conn, error)) (*WriteQueue, error) {
	swqMux.Lock()
	queue, present := server_write_queues[socket_id]
	if !present {
		log.WithFields(log.Fields{
			"at":         "queueForDestinationDialIfNeeded",
			"session_id": session_id,
			"socket_id":  socket_id,
		}).Debug("dialing destination")
		destination, err := dial_destination()
		if err != nil {
			log.WithFields(log.Fields{
				"at":         "queueForDestinationDialIfNeeded",
				"session_id": session_id,
				"socket_id":  socket_id,
				"error":      err.Error(),
			}).Error("error dialing destination")
			return queue, err
		}
		queue = NewWriteQueue(destination)
		server_write_queues[socket_id] = queue
		respondersMux.Lock()
		if imuxer, ok := responders[session_id]; ok {
			go imuxer.ReadFrom(socket_id, destination, session_id)
		} else {
			log.WithFields(log.Fields{
				"at":         "queueForDestinationDialIfNeeded",
				"session_id": session_id,
				"socket_id":  socket_id,
			}).Fatal("no responding reader exists, should not be possible")
		}
		respondersMux.Unlock()
	}
	swqMux.Unlock()
	return queue, nil
}
