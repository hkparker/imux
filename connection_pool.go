package imux

import (
	"net"
)

type ConnectionPool struct {
	Destination string
	Binds       map[string]int
	ChunkSize   int
	chunks      chan Chunk
}

func NewConnectionPool(destination string, socket_count int, chunk_size int) ConnectionPool {
	// for each socket
	go func() {
		// read some chunk from the channel
		// send it down a stream writer?
	}()
	return ConnectionPool{}
}

func (connection_pool *ConnectionPool) chunksBackTo(socket net.Conn, socket_id string) {

}

func (connection_pool *ConnectionPool) end(socket_id string) {

}
