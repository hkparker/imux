package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"io"
	"net"
	"reflect"
)

var WriteQueues = make(map[string]WriteQueue)

type Redailer func() (net.Conn, error)
type IMUXConfig struct {
	Transport string
	Binds     map[string]int
}

// Provide a net.Listener, for which any accepted sockets will have their data
// inverse multiplexed to
func OneToMany(listener net.Listener, config IMUXConfig) error {
	// Create a new SessionID shared by all sockets accepted by this OneToMany
	session_id := "" //uuid4

	// Create a new DataIMUX to read data from accepted connections and
	// chunk all data.  Create IMUXSockets to read chunks from the DataIMUX
	// and write them to connections to the server.
	imuxer := NewDataIMUX(session_id)
	for _, count := range config.Binds {
		for i := 0; i < count; i++ {
			imux_socket := IMUXSocket{
				IMUXer:   imuxer,
				Redailer: new_redailer(),
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
		// take return chunks and write them into thi socket
		WriteQueues[socket_id] = WriteQueue{
			Destination: io.Writer(socket),
		}

		imuxer.ReadFrom(socket)
	}
}

// Create a new TLJ server to accept chunks from anywhere
// and order them, writing them to corresponding sockets.
func ManyToOne(listener net.Listener, destination string) {
	tlj_server := tlj.NewServer(listener, tag_socket, type_store())
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(Chunk); ok {
			// if a response imuxer doesn't exist for this session ID
			//   create a new response imuxer
			// if the socket isn't looping chunks back
			//   forever read from response imuxer for chunk's session and write data back
			//   if errors writing data, give back to response imuser as stale
			queue, present := WriteQueues[chunk.SocketID]
			if !present {
				// dial new outgoing socket
				// response_imuxer for session id .ReadFrom(destination_socket)
				// Create new WriteQueue addressed by SocketID
			}
			queue.Write(chunk)
		}
	})

	err := <-tlj_server.FailedServer
	log.WithFields(log.Fields{
		"error": err.Error(),
	}).Error()
}

func new_redailer() Redailer {
	return func() (net.Conn, error) {
		// Dial new conn
		// Create a TLJ server for this new conn that
		// accepts return chunks, validates session, writes
		// chunk into write queue for correct socket
		return nil, nil
	}
}
