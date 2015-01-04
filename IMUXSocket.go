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

func (imuxsocket *IMUXSocket) Download(buffer Buffer, done chan string) {
	// still need to keep track of speed
	// still need to add recycling
	for {
		// re open socket if needed, channel that sockets can be grabbed from?
		
		// Get the chunk header from the server
		header_slice = make([]byte, 32)
		_, err := imuxsocket.Socket.Read(header_slice)
		if err != nil {
			var err_msg bytes.Buffer
			err_msg.WriteString("Error reading chunk header from socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		// Check if the header was all 0s
		total := 0
		for _, data := range header_slice {
			total += data
		}
		if total == 0 {
			done <- "0"
			break
		}
		
		// Parse chunk information and read data
		header := strings.Fields(string(header_slice))
		id := header[0]
		size := header[1]
		chunk_data := done <- make([]byte, size)
		_, err := imuxsocket.Socket.Read(chunk_data)
		if err != nil {
			var err_msg bytes.Buffer
			err_msg.WriteString("Error reading chunk data from socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
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

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue, done chan string) {
	// still need to keep track of speed
	// still need to add recycling
	for chunk := range queue.Chunks {
		// Get a new socket if recycling is on
		
		// Create the chunk header containing ID and size
		header, err := chunk.GenerateHeader()
		if err != nil {
			done <- err
			break
		}
		
		// Send the chunk header
		_ , err := imuxsocket.Socket.Write(header)
		if err != nil {
			queue.StaleChunks <- chunk
			var err_msg bytes.Buffer
			err_msg.WriteString("Error writing chunk header to socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		// Send the chunk data
		_, err := imuxsocket.Socket.Write(chunk.Data)
		if err != nil {
			queue.StaleChunks <- chunk
			var err_msg bytes.Buffer
			err_msg.WriteString("Error writing chunk data to socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		// Recycle the socket if needed
		
	}
	
	// Write 32 bytes of 0s to indicate there are no more chunks
	imuxsocket.Socket.Write(make([]byte, 32))
}

func (imuxsocket *IMUXSocket) Close() error {
	return imuxsocket.Socket.Close()
}
