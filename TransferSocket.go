package main

import (
	"crypto/tls"
	"strconv"
	"bytes"
	"time"
)

type TransferSocket struct {
	Socket tls.Conn
	Group TransferGroup
	LastSpeed float64
	Recycle bool
	GenerateSocket func(transfer_socket *TransferSocket) error
}

func PrepareError(description string, message string) string {
	var err_msg bytes.Buffer
	err_msg.WriteString(description)
	err_msg.WriteString(message)
	return err_msg.String()
}

func (transfer_socket *TransferSocket) Download(buffer Buffer, done chan string) {
	re_open := false
	for {
		// For keeping track of the transfer speed
		start := time.Now()
		
		// Re-open the socket if it was closed by recycling after the last chunk
		if re_open {
			err = transfer_socket.GenerateSocket(&transfer_socket)
			if err != nil {
				done <- err.Error()
				break
			}
		}
		
		// Get the chunk header from the server, 32 byte array containing id and size
		header_slice = make([]byte, 32)
		_, err := transfer_socket.Socket.Read(header_slice)
		if err != nil {
			done <- PrepareError("Error reading chunk header from socket: ", err)
			delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
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
		_, err = transfer_socket.Socket.Read(chunk_data)
		if err != nil {
			done <- PrepareError("Error reading chunk data from socket: ", err)
			delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
			break
		}
		
		// Create a chunk struct and send it to the buffer
		chunk := Chunk{}
		chunk.ID = id
		chunk.Data = chunk_data
		buffer.Chunks <- chunk
		
		// Recycle the socket if needed
		if transfer_socket.Recycle {
			transfer_socket.Socket.Write([0x01])
			transfer_socket.Close()
			re_open = true
		} else {
			transfer_socket.Socket.Write([0x00])
			re_open = false
		}
		
		// Update the transfer speed
		transfer_socket.LastSpeed = size / time.Since(start)
	}
	
	// Update the speed of the socket to 0 because the transfer is over
	transfer_socket.LastSpeed = 0
}

func (transfer_socket *TransferSocket) Serve(queue ReadQueue, done chan string) {
	re_open := false
	for chunk := range queue.Chunks {
		// Keep track of transfer speed
		start := time.Now()
		
		// Re-open the socket if recycling closed it
		if re_open {
			transfer_socket.GenerateSocket(&transfer_socket)
			if err != nil {
				done <- err.Error()
				break
			}
		}
		
		// Create the chunk header containing ID and size
		header, err := chunk.GenerateHeader()
		if err != nil {
			done <- err.Error()
			delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
			break
		}
		
		// Send the chunk header
		_ , err = transfer_socket.Socket.Write(header)
		if err != nil {
			queue.StaleChunks <- chunk
			done <- PrepareError("Error writing chunk header to socket: ", err)
			delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
			break
		}
		
		// Send the chunk data
		_, err = transfer_socket.Socket.Write(chunk.Data)
		if err != nil {
			queue.StaleChunks <- chunk
			done <- PrepareError("Error writing chunk data to socket: ", err)
			delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
			break
		}
		
		// Recycle the socket if the download routine requests
		recycle_request = make([]byte, 1)
		_, err := transfer_socket.Socket.Read(recycle_request)
		if recycle_request[0] == 0x01 {
			transfer_socket.Close()
			re_open = true
		} else {
			re_open = false
		}
		
		// Update transfer speed
		transfer_socket.LastSpeed = queue.ChunkSize / time.Since(start)
	}
	
	// Write 32 bytes of 0s to indicate there are no more chunks
	transfer_socket.Socket.Write(make([]byte, 32))
	
	// Update the speed of the socket to 0 because the transfer is over
	transfer_socket.LastSpeed = 0
}

func (transfer_socket *TransferSocket) Close() error {
	delete(transfer_socket.Group.Sockets, imuxsoket.UUID)
	return transfer_socket.Socket.Close()
}



//https://github.com/go-av/tls-example

func main() {
	queue := ReadQueue{}
	queue.ChunkSize = 1024
	buffer := Buffer{}
	file, _ := os.Open("/hayden/Pictures/render.png")
	defer file.Close()
	dst_file, _ := os.Create("/hayden/Pictures/render2.png")
	defer dst_file.Close()
	go queue.Process(file)
	go buffer.Process(dst_file)
	time.Sleep(time.Second)
	for chunk := range queue.Chunks {
		buffer.Chunks <- chunk
	}
}
