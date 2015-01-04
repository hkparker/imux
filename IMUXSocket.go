//package multiplexity
package main

import (
	"crypto/tls"
	"encoding/binary"
	"strconv"
	"bytes"
)

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float64
	Recycle bool
}


func (imuxsocket *IMUXSocket) Open(dialer net.Dialer) int {
	return 0
}

func (imuxsocket *IMUXSocket) Recieve() int {	// server is provided by manager  imuxsocket.Manager.IMUXServer?
	return 0
}

// perhaps have an options to maintain socket count?
func (imuxsocket *IMUXSocket) Download(buffer Buffer, done chan int) {
	defer done <- 0
	for {
		// re open socket if needed, channel that sockets can be grabbed from?
		
		// Get the chunk header from the server
		header_slice = make([]byte, 32)
		_, err := imuxsocket.Socket.Read(header_slice)
		if err != nil {
			// socket broken while trying to read chunk header
			break
		}
		
		// Check if the header was all 0s
		total := 0
		for _, data := range header_slice {
			total += data
		}
		if total == 0 {
			break
		}
		
		// Parse chunk information and read data
		header := strings.Fields(string(header_slice))
		id := header[0]
		size := header[1]
		chunk_data := make([]byte, size)
		_, err := imuxsocket.Socket.Read(chunk_data)
		if err != nil {
			// socket broken while trying to read chunk data
			break
		}
		
		// Create chunk and send to buffer
		chunk := Chunk{}
		chunk.ID = id
		chunk.Data = chunk_data
		buffer.Chunks <- chunk
		
		// Recycle socket if needed
		
	}
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue) {
	for chunk := range queue.Chunks {
		// Get a new socket if recycling is on
		
		// Create the chunk header containing ID and size
		header, err := chunk.GenerateHeader()
		if err != nil {
			// log.Write(err)
			break
		}
		
		// Send the chunk header
		_ , err := imuxsocket.Socket.Write(header)
		if err != nil {
			// socket error when sending chunk header
			// this is a stale chunk, this worker is dead
		}
		
		// Send the chunk data
		_, err := imuxsocket.Socket.Write(chunk.Data)
		if err != nil {
			// socket error when sending chunk data
			// this is a stale chunk, this worker is dead
		}
		
		// Recycle the socket if needed
		
	}
	
	// Write 32 bytes of 0s to indicate there are no more chunks
	imuxsocket.Socket.Write(make([]byte, 32))
}

func (imuxsocket *IMUXSocket) Close() int {
	return 0
}
