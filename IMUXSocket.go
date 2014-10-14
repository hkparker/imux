//package multiplexity
package main

import "crypto/tls"

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float32
	Recycle bool
}


func (imuxsocket IMUXSocket) Open() int {
	return 0
}

func (imuxsocket IMUXSocket) Recieve() int {	// server is provided by manager
	return 0
}

func (imuxsocket IMUXSocket) Download() int {
	// download chunks from socket to buffer (a buffer is created by the manager and passed in)
	// using channels in the buffer?
	// buffer.RecieveChannel <- new_chunk
	// buffer has a for loop reading from channel, insertion sorting, and dumping to disk
	return 0
}

func (imuxsocket IMUXSocket) Upload() int {
	// serve chunks from buffer to socket (a read queue is created by the manager and passed in)
	return 0
}

func (imuxsocket IMUXSocket) Close() int {
	return 0
}
