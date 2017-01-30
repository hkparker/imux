package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"io"
	"net"
	"reflect"
	"sync"
)

// WriteQueues for each outgoing socket on the server
var server_write_queues = make(map[string]WriteQueue)
var SWQMux sync.Mutex

// DataIMUX objects to read responses from each outgoing destination socket
var responders = make(map[string]DataIMUX)
var RespondersMux sync.Mutex

// Tracks if goroutines have been created for each socket to read from the
// DataIMUXer for its session and write responses down
var loopers = make(map[net.Conn]bool)
var LoopersMux sync.Mutex

// Create a new TLJ server to accept chunks from anywhere and order them, writing them to corresponding sockets
func ManyToOne(listener net.Listener, dial_destination func() (net.Conn, error)) {
	tlj_server := tlj.NewServer(listener, tag_socket, type_store())
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(*Chunk); ok {
			log.WithFields(log.Fields{
				"sequence_id": chunk.SequenceID,
				"socket_id":   chunk.SocketID,
				"session_id":  chunk.SessionID,
			}).Debug("received chunk")
			createResponderIMUXIfNeeded(chunk.SessionID)
			writeResponseChunksIfNeeded(context.Socket, chunk.SessionID)
			updateSessionChunkSize(chunk.SessionID, len(chunk.Data))
			queue, err := queueForDestinationDialIfNeeded(chunk.SocketID, chunk.SessionID, dial_destination)
			if err == nil {
				queue.Write(chunk)
				log.WithFields(log.Fields{
					"error":       err.Error(),
					"sequence_id": chunk.SequenceID,
					"socket_id":   chunk.SocketID,
					"session_id":  chunk.SessionID,
				}).Debug("wrote chunk")
			} else {
				log.WithFields(log.Fields{
					"error":       err.Error(),
					"sequence_id": chunk.SequenceID,
					"socket_id":   chunk.SocketID,
					"session_id":  chunk.SessionID,
				}).Error("dropped chunk")
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

// If it does not exist, create a DataIMUX to read data from
// outgoing destination sockets with a common session
func createResponderIMUXIfNeeded(session_id string) {
	log.WithFields(log.Fields{
		"at":         "createResponderIMUXIfNeeded",
		"session_id": session_id,
	}).Debug("checking if responder data imux exists")
	RespondersMux.Lock()
	if _, present := responders[session_id]; !present {
		responders[session_id] = NewDataIMUX(session_id)
		log.WithFields(log.Fields{
			"at":         "createResponderIMUXIfNeeded",
			"session_id": session_id,
		}).Debug("created new responder imux for session")
	}
	RespondersMux.Unlock()
}

// If it is not already happening, ensure that response chunks for a specified
// session_id are written back down this socket.
func writeResponseChunksIfNeeded(socket net.Conn, session_id string) {
	log.WithFields(log.Fields{
		"at":         "writeResponseChunksIfNeeded",
		"session_id": session_id,
	}).Debug("checking if socket needs chunk data written back down it")
	LoopersMux.Lock()
	if _, looping := loopers[socket]; !looping {
		log.WithFields(log.Fields{
			"at":         "writeResponseChunksIfNeeded",
			"session_id": session_id,
		}).Debug("creating goroutine to write data back down transport socket")
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
			RespondersMux.Lock()
			chunk_stream := responders[session_id]
			RespondersMux.Unlock()
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
	LoopersMux.Unlock()
}

// Increase the server side chunk size for a session if a new largest chunk has been seen
func updateSessionChunkSize(session_id string, data_len int) {
	log.WithFields(log.Fields{
		"at":         "updateSessionChunkSize",
		"session_id": session_id,
		"data_len":   data_len,
	}).Debug("checking if chunk size is new session max")
	OBCSMux.Lock()
	if size, present := ObservedMaximumChunkSizes[session_id]; present {
		if data_len > size {
			log.WithFields(log.Fields{
				"at":         "updateSessionChunkSize",
				"session_id": session_id,
				"data_len":   data_len,
			}).Debug("updating chunk size")
			ObservedMaximumChunkSizes[session_id] = data_len
		}
	} else {
		log.WithFields(log.Fields{
			"at":         "updateSessionChunkSize",
			"session_id": session_id,
			"data_len":   data_len,
		}).Debug("setting chunk size for first time")
		ObservedMaximumChunkSizes[session_id] = data_len
	}
	OBCSMux.Unlock()
}

// Get the queue a new chunk should go to, dialing the outgoing destination socket if this is the first time
// a socket ID has been observed.
func queueForDestinationDialIfNeeded(socket_id, session_id string, dial_destination func() (net.Conn, error)) (WriteQueue, error) {
	log.WithFields(log.Fields{
		"at":         "queueForDestinationDialIfNeeded",
		"session_id": session_id,
		"socket_id":  socket_id,
	}).Debug("checking if destination needs to be dialed and write queue created")
	SWQMux.Lock()
	queue, present := server_write_queues[socket_id]
	SWQMux.Unlock()
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
		}
		queue = WriteQueue{
			Destination: io.Writer(destination),
		}
		SWQMux.Lock()
		server_write_queues[socket_id] = queue
		SWQMux.Unlock()
		RespondersMux.Lock()
		if imuxer, ok := responders[session_id]; ok {
			imuxer.ReadFrom(socket_id, destination, session_id, "server")
		} else {
			log.WithFields(log.Fields{
				"at":         "queueForDestinationDialIfNeeded",
				"session_id": session_id,
				"socket_id":  socket_id,
			}).Fatal("no responding reader exists, should not be possible")
		}
		RespondersMux.Unlock()
	}
	return queue, nil
}
