package main

import "os"

type WriteBuffer struct {
	Chunks   chan Chunk
	Queue    []Chunk
	LastDump int
}

func (buffer *WriteBuffer) Insert(chunk Chunk) {
	smaller := 0
	for _, item := range buffer.Queue {
		if item.ID < chunk.ID {
			smaller++
		}
	}
	smaller_chunks := buffer.Queue[:smaller]
	larger_chunks := buffer.Queue[smaller:]
	buffer.Queue = append(smaller_chunks, append([]Chunk{chunk}, larger_chunks...)...)
}

func (buffer *WriteBuffer) Dump(file *os.File) {
	for {
		if len(buffer.Queue) == 0 {
			break
		}
		chunk := buffer.Queue[0]
		if chunk.ID == buffer.LastDump+1 {
			buffer.Queue = buffer.Queue[1:]
			file.Write([]byte(chunk.Data))
			buffer.LastDump = buffer.LastDump + 1
		} else {
			break
		}
	}
}

func (buffer *WriteBuffer) Process(file *os.File) {
	buffer.Chunks = make(chan Chunk, 10)
	for chunk := range buffer.Chunks {
		buffer.Insert(chunk)
		buffer.Dump(file)
	}
}
