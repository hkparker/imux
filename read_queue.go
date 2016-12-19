package imux

import (
	"net"
)

type ReadQueue struct{}

func Process(input net.Conn, destination ConnectionPool) ReadQueue {
	return ReadQueue{}
}

func (read_queue *WriteQueue) end() {}
