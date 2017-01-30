package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
	"sync"
	"time"
)

// A map of all TLJ servers used to read chunks back from sessions
var SessionResponsesTLJServers = make(map[string]tlj.Server)
var SRTSMux sync.Mutex

// A client socket that transports data in an imux session, autoreconnecting
type IMUXSocket struct {
	IMUXer   DataIMUX
	Redialer Redialer
}

// Dial a new connection in an imux session, creating a TLJ server for
// responses if needed.  Read data from the sockets IMUXer and write it up.
func (imux_socket *IMUXSocket) init(session_id string) {
	log.WithFields(log.Fields{
		"at": "IMUXSocket.init",
	}).Debug("starting imux socket")
	tlj_server := imuxClientSocketTLJServer(session_id)
	cooldown := 10 * time.Second
	for {
		log.WithFields(log.Fields{
			"at": "IMUXSocket.init",
		}).Debug("dialing imux socket")
		socket, err := imux_socket.Redialer()
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "IMUXSocket.init",
				"error": err.Error(),
			}).Error("error dialing imux socket, entering cooldown")
			time.Sleep(cooldown)
			continue
		}
		tlj_server.Insert(socket)
		writer, err := tlj.NewStreamWriter(socket, type_store(), reflect.TypeOf(Chunk{}))
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "IMUXSocket.init",
				"error": err.Error(),
			}).Error("error creating stream writer, entering cooldown")
			time.Sleep(cooldown)
			continue
		}

		for {
			chunk := <-imux_socket.IMUXer.Chunks
			log.WithFields(log.Fields{
				"at":          "IMUXSocket.init",
				"sequence_id": chunk.SequenceID,
				"socket_id":   chunk.SocketID,
				"session_id":  chunk.SessionID,
			}).Debug("writing chunk up transport socket")
			err := writer.Write(chunk)
			if err != nil {
				imux_socket.IMUXer.Stale <- chunk
				log.WithFields(log.Fields{
					"at":          "IMUXSocket.init",
					"error":       err.Error(),
					"sequence_id": chunk.SequenceID,
					"socket_id":   chunk.SocketID,
					"session_id":  chunk.SessionID,
				}).Error("error writing chunk up transport socket")
				break
			}
		}
		log.WithFields(log.Fields{
			"at": "IMUXSocket.init",
		}).Debug("transport socket dies, redailing after cooldown")
		time.Sleep(cooldown)
	}
}

// Create a TLJ server for a session if needed, or return the already existing server
func imuxClientSocketTLJServer(session_id string) tlj.Server {
	log.WithFields(log.Fields{
		"at":         "imuxClientSocketTLJServer",
		"session_id": session_id,
	}).Debug("checking if new TLJ server needed for session")
	SRTSMux.Lock()
	if server, exists := SessionResponsesTLJServers[session_id]; exists {
		log.WithFields(log.Fields{
			"at":         "imuxClientSocketTLJServer",
			"session_id": session_id,
		}).Debug("returning existing TLJ server for session")
		return server
	}
	SRTSMux.Unlock()
	tlj_server := tlj.Server{
		TypeStore:       type_store(),
		Tag:             tag_socket,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	go func(server tlj.Server) {
		for {
			<-server.FailedSockets
		}
	}(tlj_server)
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(Chunk); ok {
			CWQMux.Lock()
			if writer, ok := client_write_queues[chunk.SocketID]; ok {
				log.WithFields(log.Fields{
					"at":         "imuxClientSocketTLJServer",
					"session_id": session_id,
				}).Debug("accepting response chunk in transport socket TLJ server")
				writer.Write(chunk)
			} else {
				log.WithFields(log.Fields{
					"at":         "imuxClientSocketTLJServer",
					"session_id": session_id,
					"socket_id":  chunk.SocketID,
				}).Error("could not find write queue for response chunk")
			}
			CWQMux.Unlock()
		}
	})
	SRTSMux.Lock()
	SessionResponsesTLJServers[session_id] = tlj_server
	SRTSMux.Unlock()
	log.WithFields(log.Fields{
		"at":         "imuxClientSocketTLJServer",
		"session_id": session_id,
	}).Debug("created new TLJ server for session")
	return tlj_server
}
