package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLB"
	"net"
	"reflect"
	"sync"
	"time"
)

// A function that can be called by an IMUXSocket to reconnect after an error
type Redialer func() (net.Conn, error)

// A function that generates Redialers for specific bind addresses
type RedialerGenerator func(string) Redialer

// A map of all TLB servers used to read chunks back from sessions
var sessionResponsesTLBServers = make(map[string]tlb.Server)
var srtsMux sync.Mutex

// A client socket that transports data in an imux session, autoreconnecting
type IMUXSocket struct {
	IMUXer   DataIMUX
	Redialer Redialer
}

// Dial a new connection in an imux session, creating a TLB server for
// responses if needed.  Read data from the sockets IMUXer and write it up.
func (imux_socket *IMUXSocket) init(session_id string) {
	log.WithFields(log.Fields{
		"at": "IMUXSocket.init",
	}).Debug("starting imux socket")
	tlb_server := imuxClientSocketTLBServer(session_id)
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
		tlb_server.Insert(socket)
		writer, err := tlb.NewStreamWriter(socket, type_store(), reflect.TypeOf(Chunk{}))
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

// Create a TLB server for a session if needed, or return the already existing server
func imuxClientSocketTLBServer(session_id string) tlb.Server {
	srtsMux.Lock()
	defer srtsMux.Unlock()
	if server, exists := sessionResponsesTLBServers[session_id]; exists {
		return server
	}
	tlb_server := tlb.Server{
		TypeStore:       type_store(),
		Tag:             tag_socket,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, tlb.TLBContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, tlb.TLBContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	go func(server tlb.Server) {
		for {
			<-server.FailedSockets
		}
	}(tlb_server)
	tlb_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlb.TLBContext) {
		if chunk, ok := iface.(*Chunk); ok {
			cwqMux.Lock()
			if writer, ok := client_write_queues[chunk.SocketID]; ok {
				log.WithFields(log.Fields{
					"at":         "imuxClientSocketTLBServer",
					"session_id": session_id,
				}).Debug("accepting response chunk in transport socket TLB server")
				writer.Chunks <- chunk
			} else {
				log.WithFields(log.Fields{
					"at":         "imuxClientSocketTLBServer",
					"session_id": session_id,
					"socket_id":  chunk.SocketID,
				}).Error("could not find write queue for response chunk")
			}
			cwqMux.Unlock()
		}
	})
	sessionResponsesTLBServers[session_id] = tlb_server
	log.WithFields(log.Fields{
		"at":         "imuxClientSocketTLBServer",
		"session_id": session_id,
	}).Debug("created new TLB server for session")
	return tlb_server
}
