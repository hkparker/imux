//package multiplexity
package main

import "crypto/tls"

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float64
	Recycle bool
}


func (imuxsocket IMUXSocket) Open() int {
	return 0
}

func (imuxsocket IMUXSocket) Recieve() int {	// server is provided by manager  imuxsocket.Manager.IMUXServer?
	return 0
}

func (imuxsocket IMUXSocket) Download(buffer Buffer) int {
	// download chunks from imuxsocket.Socket to buffer (a buffer is created by the manager and passed in)
	// using channels in the buffer?
	// buffer.RecieveChannel <- new_chunk
	// buffer has a for loop reading from channel, insertion sorting, and dumping to disk
	// yeah, manager creates buffer, which contains chan.  Buffer reads from chan until it is closed.  manager closes chan after download.
	
	
	return 0
}

func (imuxsocket IMUXSocket) Upload(queue ChunkQueue) int {
	// serve chunks from queue to imuxsocket.Socket (a read queue is created by the manager and passed in)
	return 0
}

func (imuxsocket IMUXSocket) Close() int {
	return 0
}
