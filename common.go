package main

import (
	"fmt"
	"github.com/hkparker/TLJ"
	"net"
	"os"
)

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func ParseFileList(items []string) ([]string, int) {
	// for each string which should be a full path
	// if it is a directory
	all_files := items //make([]string, 0)
	//visitor := func() {
	// if the item is a readable file, add it to the list
	//}
	//walk
	return all_files, 100 // also return the sinze in bytes of all files
}

func CreatePooledChunkChan(files []string, chunk_size int) (chan TransferChunk, chan string) {
	all_chunks := make(chan TransferChunk, 0)
	file_done := make(chan string, 0)
	for _, file := range files {
		queue := ReadQueue{
			ChunkSize: chunk_size,
		}
		fh, err := os.Open(file)
		if err == nil {
			// assign destination name as
			go queue.Process(fh)
			go func() {
				for chunk := range queue.Chunks {
					all_chunks <- chunk
				}
				file_done <- file
			}()
		} else {
			fmt.Println(err)
		}
	}
	return all_chunks, file_done
}

func StreamChunksToPut(worker tlj.Client, chunks chan TransferChunk, speed_update, total_update chan int) {
	for chunk := range chunks {
		// set start time
		err := worker.Message(chunk) // instead do an action and have it also be responder.Respond
		if err != nil {
			// make this chunk stale for the read queue?
		}
		// if this was the last chunk for a file, send back that the update was success
		// take note of elapsed time and chunk size, update my speed, update amount moved
	}
}
