package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"io"
	"net"
)

var client_write_queues = make(map[string]WriteQueue)

// Provide a net.Listener, for which any accepted sockets will have their data
// inverse multiplexed to
func OneToMany(listener net.Listener, binds map[string]int, redialer_generator RedialerGenerator) error {
	// Create a new SessionID shared by all sockets accepted by this OneToMany
	session_id := uuid.NewV4().String()
	log.WithFields(log.Fields{
		"at":         "OneToMany",
		"session_id": session_id,
		"binds":      binds,
	}).Debug("creating new OneToMany")

	// Create a new DataIMUX to read data from accepted connections and
	// chunk all data.  Create IMUXSockets to read chunks from the DataIMUX
	// and write them to connections to the server.
	imuxer := NewDataIMUX(session_id)
	for bind, count := range binds {
		for i := 0; i < count; i++ {
			log.WithFields(log.Fields{
				"at":         "OneToMany",
				"bind":       bind,
				"session_id": session_id,
			}).Debug("creating new imux socket")
			imux_socket := IMUXSocket{
				IMUXer:   imuxer,
				Redialer: redialer_generator(bind),
			}
			imux_socket.init(session_id)
		}
	}

	// In an infinite loop, accept new connections to this listener
	// and read data into the session DataIMUX.  Create a WriteQueue
	// to write out return chunks.
	for {
		socket, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"at":         "OneToMany",
				"session_id": session_id,
				"error":      err.Error(),
			}).Debug("error accepting new inbound connection to imux")
			return err
		}
		socket_id := uuid.NewV4().String()
		log.WithFields(log.Fields{
			"at":         "OneToMany",
			"session_id": session_id,
			"socket_id":  socket_id,
		}).Debug("accepted new inbound connection to imux")

		// Create a new WriteQueue addressed by the socket ID to
		// take return chunks and write them into this socket
		client_write_queues[socket_id] = WriteQueue{
			Destination: io.Writer(socket),
		}

		go imuxer.ReadFrom(socket_id, socket, session_id, "client")
	}
}