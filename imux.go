package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"io"
	"net"
	"reflect"
)

var WriteQueues = make(map[string]WriteQueue)
var ClientChunkSize = 16384

type Redailer func() (net.Conn, error)
type IMUXConfig struct {
	Transport string
	Binds     map[string]int
}
type DataIMUX struct {
	Chunks    chan Chunk
	Stale     chan Chunk
	SessionID string
}

func NewDataIMUX(session_id string) DataIMUX {
	return DataIMUX{
		Chunks:    make(chan Chunk, 10),
		Stale:     make(chan Chunk, 50),
		SessionID: session_id,
	}
}

func (data_imux *DataIMUX) ReadFrom(id string, conn io.Reader) {
	// determine the chunk size
	var uint64 sequence = 1
	for {
		chunk_data := []byte(ClientChunkSize)
		_, err := conn.Read(&chunk_data)
		if err != nil {
		}
		data_imux.Chunks <- Chunk{
			SequenceID: sequence,
			SocketID:   id,
			SessionID:  data_imux.SessionID,
			Data:       chunk_data,
		}
		equence += 1
	}
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
		// take return chunks and write them into this socket
		WriteQueues[socket_id] = WriteQueue{
			Destination: io.Writer(socket),
		}

		go imuxer.ReadFrom(socket_id, socket)
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
					writer := tlj.NewStreaWwriter(context.Socket, type_store(), reflect.TypeOf(Chunk{}))
					for {
						new_chunk := <-responders[chunk.SessionID].Chunks
						err := writer.write(new_chunk)
						if err != nil {
							responders[chunk.SessionID].Stale <- new_chunk
							break
						}
					}
				}()
			}
			queue, present := WriteQueues[chunk.SocketID]
			if !present {
				destination, err := dial_destiation()
				if err != nil {
				}
				queue = WriteQueue{
					Destination: io.Writer(destination),
				}
				WriteQueues[chunk.SocketID] = queue
				go responders[chunk.SessionID].ReadFrom(chunk.SocketID, destination)
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

func tag_socket(socket net.Conn, server *tlj.Server) {
	server.TagSocket(socket, "all")
}

func type_store() tlj.TypeStore {
	type_store := tlj.NewTypeStore()
	type_store.AddType(
		reflect.TypeOf(Chunk{}),
		reflect.TypeOf(&Chunk{}),
		BuildChunk,
	)
	return type_store
}
