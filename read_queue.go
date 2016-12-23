package imux

import (
	"net"
)

type ReadQueue struct {
	Input       net.Conn
	Destination ConnectionPool
}

func (read_queue ReadQueue) process() {
}

func (read_queue *WriteQueue) end() {}
