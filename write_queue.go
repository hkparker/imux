package imux

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

// A WriteQueue will receive chunks and order them, writing
// their data out to the Destination in the correct order
type WriteQueue struct {
	Destination io.Writer
	LastDump    int
	Queue       []*Chunk
}

// Accept a new chunk and dump any possible data
func (write_queue *WriteQueue) Write(chunk *Chunk) {
	log.WithFields(log.Fields{
		"at":       "WriteQueue.Write",
		"sequence": chunk.SequenceID,
		"socket":   chunk.SocketID,
		"session":  chunk.SessionID,
	}).Debug("accepting new chunk")
	write_queue.Insert(chunk)
	write_queue.Dump()
}

// Place a chunk in the correct location in the queue
func (write_queue *WriteQueue) Insert(chunk *Chunk) {
	smaller := 0
	for _, item := range write_queue.Queue {
		if item.SequenceID < chunk.SequenceID {
			smaller++
		}
	}
	smaller_chunks := write_queue.Queue[:smaller]
	larger_chunks := write_queue.Queue[smaller:]
	write_queue.Queue = append(smaller_chunks, append([]*Chunk{chunk}, larger_chunks...)...)
}

// Dump as much chunk data out the Destination as available in order
func (write_queue *WriteQueue) Dump() {
	for {
		if len(write_queue.Queue) == 0 {
			break
		}
		chunk := write_queue.Queue[0]
		if chunk.SequenceID == uint64(write_queue.LastDump+1) {
			log.WithFields(log.Fields{
				"at":       "WriteQueue.Dump",
				"sequence": chunk.SequenceID,
				"socket":   chunk.SocketID,
				"session":  chunk.SessionID,
			}).Debug("writing out chunk data")
			write_queue.Queue = write_queue.Queue[1:]
			write_queue.Destination.Write(chunk.Data)
			write_queue.LastDump = write_queue.LastDump + 1
		} else {
			break
		}
	}
}
