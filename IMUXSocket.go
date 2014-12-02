//package multiplexity
package main

import (
	"crypto/tls"
	"encoding/binary"
	"strconv"
)

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float64
	Recycle bool
}


func (imuxsocket *IMUXSocket) Open() int {
	return 0
}

func (imuxsocket *IMUXSocket) Recieve() int {	// server is provided by manager  imuxsocket.Manager.IMUXServer?
	return 0
}

// perhaps have an options to maintain socket count?
func (imuxsocket *IMUXSocket) Download(buffer Buffer) int {
	// defer a send to a complete channel?
	for {
		// re open socket if needed
		// channel that sockets can be grabbed from?
		cmd_slice = make([]byte, 32)
		_, err := imuxsocket.Socket.Read(cmd_slice)
		if err != nil {
			break
		}
		total := 0
		for _, data := range cmd_slice {
			total += data
		}
		if total == 0 {
			break
		} else {
			colon := 0
			for iter, data := range cmd_slice {
				if data == 58 {
					colon = iter
					break
				}
			}
			id_bytes := cmd_slice[:colon]
			size_bytes := cmd_slice[colon:]
			id := binary.BigEndian.Uint64(id_bytes)
			size := binary.BigEndian.Uint64(size_bytes)
			chunk_data := make([]byte, size)
			_, err := imuxsocket.Socket.Read(chunk_data)
			if err != nil {
				break
			}
			chunk := Chunk{}
			chunk.ID = id
			chunk.Data = chunk_Data
			buffer.Chunks <- chunk
		}
	}
	return 0
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue) int {
	for chunk := range queue.Chunks {
		// re open socket if needed
		
		chunk_header := make([]byte, 32)
		chunk_size := len(chunk.Data)
		chunk_size_bytes := []byte(strconv.Itoa(chunk_size))
		len_chunk_size := len(chunk_size_bytes)
		if >= 29 {
			// overflow!
		} else {
			// merge chunk size onto the back of the slize
			chunk_header[32-len_chunk_size:] = chunk_size_bytes
			fmt.Println(string(chunk_header))
		}
		
		
		//chunk_id := make([]byte, 32)
		//binary.LittleEndian.PutUint32(chunk_id, chunk.ID)
		
		// get a byte array for the id, merge it in
		
		
		// "<32-len(id:size) zeros>id:size"
		// id and size are going to need to be in a byte array
	}
	// send a done command (all zeros)
	return 0
}

func (imuxsocket *IMUXSocket) Close() int {
	return 0
}
