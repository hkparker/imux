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

func ParseFileList(items []string) []string {
	// for each string which should be a full path
	// if it is a directory
	all_files := items //make([]string, 0)
	//visitor := func() {
	// if the item is a readable file, add it to the list
	//}
	//walk
	return all_files
}

func CreatePooledChunkChan(files []string, chunk_size int) chan TransferChunk {
	all_chunks := make(chan TransferChunk, 0)
	for _, file := range files {
		queue := ReadQueue{
			ChunkSize: chunk_size,
		}
		fh, err := os.Open(file)
		if err == nil {
			// assign destination name as
			go queue.Process(fh)
			go func() {
				for {
					all_chunks <- <-queue.Chunks
				}
			}()
		} else {
			fmt.Println(err)
		}
	}
	return all_chunks
}

func StreamChunksToPut(worker tlj.Client, chunks chan TransferChunk) {
	for chunk := range chunks {
		fmt.Println("messaged a chunk")
		// set start time
		err := worker.Message(chunk) // instead do an action and have it also be responder.Respond
		if err != nil {
			// make this chunk stale for the read queue?
		}
		// if this was the last chunk for a file, send back that the update was success
		// take note of elapsed time and chunk size, update my speed
	}
}
