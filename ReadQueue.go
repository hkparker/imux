package main

import (
	"io"
	"os"
)

type ReadQueue struct {
	Chunks      chan Chunk
	StaleChunks chan Chunk
	FileHead    int64
	ChunkSize   int
	ChunkID     int
}

func (queue *ReadQueue) Process(file *os.File) {
	queue.Chunks = make(chan Chunk)
	queue.StaleChunks = make(chan Chunk, 10)
	queue.ChunkID = 1
	file.Seek(queue.FileHead, 0)
	for {
		for len(queue.StaleChunks) > 0 {
			queue.Chunks <- (<-queue.StaleChunks)
		}
		new_chunk := Chunk{}
		new_chunk.ID = queue.ChunkID
		queue.ChunkID++
		new_chunk_data := make([]byte, queue.ChunkSize)
		bytes_read, err := file.Read(new_chunk_data)
		if err == io.EOF {
			break
		}
		if bytes_read < queue.ChunkSize {
			new_chunk_data = new_chunk_data[:bytes_read]
		}
		new_chunk.Data = new_chunk_data
		queue.Chunks <- new_chunk
	}
	close(queue.Chunks)
}
