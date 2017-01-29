package imux

import (
	"io"
)

type WriteQueue struct {
	Destination io.Writer
	TopSequence int
	Chunks      []Chunk
}

func (write_queue *WriteQueue) Write(chunk Chunk) {}
