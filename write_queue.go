package imux

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

type WriteQueue struct {
	Destination io.Writer
	LastDump    int
	Queue       []Chunk
}

func (write_queue *WriteQueue) Write(chunk Chunk) {
	log.WithFields(log.Fields{
		"at":       "WriteQueue.Write",
		"sequence": chunk.SequenceID,
		"socket":   chunk.SocketID,
		"session":  chunk.SessionID,
	}).Debug("accepting new chunk")
	write_queue.Insert(chunk)
	write_queue.Dump()
}

func (write_queue *WriteQueue) Insert(chunk Chunk) {
	smaller := 0
	for _, item := range write_queue.Queue {
		if item.SequenceID < chunk.SequenceID {
			smaller++
		}
	}
	smaller_chunks := write_queue.Queue[:smaller]
	larger_chunks := write_queue.Queue[smaller:]
	write_queue.Queue = append(smaller_chunks, append([]Chunk{chunk}, larger_chunks...)...)
}

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
