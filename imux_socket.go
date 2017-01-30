package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
	"sync"
	"time"
)

var SessionResponsesTLJServers = make(map[string]tlj.Server)

type IMUXSocket struct {
	IMUXer   DataIMUX
	Redialer Redialer
}

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

func imuxClientSocketTLJServer(session_id string) tlj.Server {
	log.WithFields(log.Fields{
		"at":         "imuxClientSocketTLJServer",
		"session_id": session_id,
	}).Debug("checking if new TLJ server needed for session")
	if server, exists := SessionResponsesTLJServers[session_id]; exists {
		log.WithFields(log.Fields{
			"at":         "imuxClientSocketTLJServer",
			"session_id": session_id,
		}).Debug("returning existing TLJ server for session")
		return server
	}
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
			if writer, ok := client_write_queues[chunk.SocketID]; ok {
				writer.Write(chunk)
			} else {
			}
		}
	})
	SessionResponsesTLJServers[session_id] = tlj_server
	log.WithFields(log.Fields{
		"at":         "imuxClientSocketTLJServer",
		"session_id": session_id,
	}).Debug("created new TLJ server for session")
	return tlj_server
}
