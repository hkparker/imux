package main

import (
	"os"
)

type Buffer struct {
	Chunks chan Chunk
}

func (buffer Buffer) Insert(chunk Chunk) {
	// insertion sort chunk.ID into slice
}

func (buffer Buffer) Dump(file *File) {
	// dump all consecutive IDs to file
}

func (buffer Buffer) Process(file *File) {
	for chunk := range buffer.Chunks {
		buffer.Insert(chunk)
		buffer.Dump(file)
	}
}

func main() {
	// IMUXManager is going to call these when doing a transfer
	buffer := Buffer{}
	buffer.Chunks = make(chan Chunk, 10)
	// try to open a file handler on the destination
	file := os.File{}
	buffer.Process(file)
}
