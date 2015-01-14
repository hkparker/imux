package main

import (
	"crypto/tls"
	"strconv"
	"bytes"
	"time"
)

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float64
	Recycle bool
}

func (imuxsocket *IMUXSocket) Download(buffer Buffer, done chan string) {
	for {
		// Keep track of transfer speed
		start := time.Now()
		
		//// if reopen is true
		////	get new socket from manager
		////	if theres an error, break with error
		////	set reopen false
		
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
		id, _ := strconv.Atoi(header[0])
		size, _  := strconv.Atoi(header[1])
		chunk_data := make([]byte, size)
		_, err = imuxsocket.Socket.Read(chunk_data)
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
		
		//// if recycle
		////	send the recycling signal
		////	close the socket
		////	set reopen to true
		
		// Update transfer speed
		imuxsocket.LastSpeed = time.Since(start)
	}
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue, done chan string) {
	for chunk := range queue.Chunks {
		// Keep track of transfer speed
		start := time.Now()
		
		//// if the socket if marked as closed
		////	get a new socket from the manager
		////	mark it as opened
		
		// Create the chunk header containing ID and size
		header, err := chunk.GenerateHeader()
		if err != nil {
			done <- err
			break
		}
		
		// Send the chunk header
		_ , err = imuxsocket.Socket.Write(header)
		if err != nil {
			queue.StaleChunks <- chunk
			var err_msg bytes.Buffer
			err_msg.WriteString("Error writing chunk header to socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		// Send the chunk data
		_, err = imuxsocket.Socket.Write(chunk.Data)
		if err != nil {
			queue.StaleChunks <- chunk
			var err_msg bytes.Buffer
			err_msg.WriteString("Error writing chunk data to socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		//// if request for recycle recieved
		////	close the socket
		////	mark the socket as closed
		
		// Update transfer speed
		imuxsocket.LastSpeed = time.Since(start)
	}
	
	// Write 32 bytes of 0s to indicate there are no more chunks
	imuxsocket.Socket.Write(make([]byte, 32))
}

func (imuxsocket *IMUXSocket) Close() error {
	return imuxsocket.Socket.Close()
}



//https://github.com/go-av/tls-example

//func main() {
	//queue := ReadQueue{}
	//queue.ChunkSize = 1024
	//buffer := Buffer{}
	//file, _ := os.Open("/hayden/Pictures/render.png")
	//defer file.Close()
	//dst_file, _ := os.Create("/hayden/Pictures/render2.png")
	//defer dst_file.Close()
	//go queue.Process(file)
	//go buffer.Process(dst_file)
	//time.Sleep(time.Second)
	//for chunk := range queue.Chunks {
		//buffer.Chunks <- chunk
	//}
//}
