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
	re_open := false
	for {
		// For keeping track of the transfer speed
		start := time.Now()
		
		// Re-open the socket if it was closed by recycling after the last chunk
		if re_open {
		////	get new socket from manager
		////	if theres an error, break with error
			re_open = false
		}
		
		// Get the chunk header from the server, 32 byte array containing id and size
		header_slice = make([]byte, 32)
		_, err := imuxsocket.Socket.Read(header_slice)
		if err != nil {
			var err_msg bytes.Buffer
			err_msg.WriteString("Error reading chunk header from socket: ")
			err_msg.WriteString(err)
			done <- err_msg.String()
			break
		}
		
		// Check if the header was all 0s indicating no more chunks
		total := 0
		for _, data := range header_slice {
			total += data
		}
		if total == 0 {
			done <- "0"
			break
		}
		
		// Parse chunk information and read the chunk data
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
		
		// Create a chunk struct and send it to the buffer
		chunk := Chunk{}
		chunk.ID = id
		chunk.Data = chunk_data
		buffer.Chunks <- chunk
		
		// Recycle the socket if needed
		if imuxmanager.Recycle {
		////	send the recycling signal
		////	close the socket
			re_open = true
		}
		
		// Update the transfer speed
		imuxsocket.LastSpeed = time.Since(start)
	}
	
	// Update the speed of the socket to 0 because the transfer is over
	imuxsocket.LastSpeed = 0
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue, done chan string) {
	re_open := false
	for chunk := range queue.Chunks {
		// Keep track of transfer speed
		start := time.Now()
		
		// Re-open the socket if recycling closed it
		if re_open {
			////	get a new socket from the manager
			////	mark it as opened
		}
		
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
		////	re_open = true or false
		
		// Update transfer speed
		imuxsocket.LastSpeed = time.Since(start)
	}
	
	// Write 32 bytes of 0s to indicate there are no more chunks
	imuxsocket.Socket.Write(make([]byte, 32))
	
	// Update the speed of the socket to 0 because the transfer is over
	imuxsocket.LastSpeed = 0
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
