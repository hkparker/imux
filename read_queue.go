package imux

import (
	"os"
)

type ReadQueue struct {
	Chunks      chan TransferChunk
	StaleChunks chan TransferChunk
	FileHead    int64
	ChunkSize   int
	ChunkID     int
	Destination string
}

func (queue *ReadQueue) Process(file *os.File) {
	queue.StaleChunks = make(chan TransferChunk)
	queue.ChunkID = 1
	file.Seek(queue.FileHead, 0)
	for {
		for len(queue.StaleChunks) > 0 {
			queue.Chunks <- (<-queue.StaleChunks)
		}
		new_chunk := TransferChunk{}
		new_chunk.ID = queue.ChunkID
		new_chunk.Destination = queue.Destination
		queue.ChunkID++
		new_chunk_data := make([]byte, queue.ChunkSize)
		bytes_read, err := file.Read(new_chunk_data)
		if err != nil {
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
