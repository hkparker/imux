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
func (imuxsocket *IMUXSocket) Download(buffer Buffer, done chan int) int {
	defer done <- 0
	// defer a send to a complete channel?
	for {
		// re open socket if needed
		// channel that sockets can be grabbed from?
		header_slice = make([]byte, 32)
		_, err := imuxsocket.Socket.Read(header_slice)
		if err != nil {
			break
			// Signal the worker was killed by socket error?
		}
		
		
		total := 0
		for _, data := range header_slice {
			total += data
		}

		
		if total == 0 {
			break
			// signal that there was no more data to download
		} else {
			header := strings.Fields(string(header_slice))
			id := header[0]
			size := header[1]
			chunk_data := make([]byte, size)
			_, err := imuxsocket.Socket.Read(chunk_data)
			if err != nil {
				break
				// signal there was an error transferring data
			}
			chunk := Chunk{}
			chunk.ID = id
			chunk.Data = chunk_data
			buffer.Chunks <- chunk
		}
	}
	return 0
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue) int {
	for chunk := range queue.Chunks {
		// re open socket if needed
		var header bytes.Buffer
		
		chunk_id := strconv.Itoa(chunk.ID)
		chunk_size := strconv.Itoa(len(chunk.Data))
		
		header.WriteString(chunk_id)
		header.WriteString(" ")
		header.WriteString(chunk_size)
		space := 32-len(header)
		
		// merge chunk size onto the back of the slize
		//chunk_header[32-len_chunk_size:] = chunk_size_bytes
		fmt.Println(string(chunk_header))
		
		
		
		
		header.String()
		
		
		// "id size<32-len(id:size) spaces>"
	}
	// send a done command (all zeros)
	return 0
}

func (imuxsocket *IMUXSocket) Close() int {
	return 0
}
