package imux

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

// A WriteQueue will receive chunks and order them, writing
// their data out to the Destination in the correct order
type WriteQueue struct {
	destination io.WriteCloser
	lastDump    int
	Chunks      chan *Chunk
	queue       []*Chunk
}

func NewWriteQueue(destination io.WriteCloser) *WriteQueue {
	write_queue := WriteQueue{
		destination: destination,
		Chunks:      make(chan *Chunk, 0),
		queue:       make([]*Chunk, 0),
	}
	go write_queue.process()
	return &write_queue
}

func (write_queue *WriteQueue) process() {
	for chunk := range write_queue.Chunks {
		write_queue.insert(chunk)
		write_queue.dump()
	}
}

// Place a chunk in the correct location in the queue
func (write_queue *WriteQueue) insert(chunk *Chunk) {
	if chunk.SequenceID == 0 {
		log.WithFields(log.Fields{
			"socket":  chunk.SocketID,
			"session": chunk.SessionID,
		}).Debug("reset chunk received")
		write_queue.bail(chunk.SocketID)
		return
	}
	smaller := 0
	for _, item := range write_queue.queue {
		if item.SequenceID < chunk.SequenceID {
			smaller++
		}
	}
	smaller_chunks := write_queue.queue[:smaller]
	larger_chunks := write_queue.queue[smaller:]
	write_queue.queue = append(smaller_chunks, append([]*Chunk{chunk}, larger_chunks...)...)
}

// Dump as much chunk data out the Destination as available in order
func (write_queue *WriteQueue) dump() {
	for {
		if len(write_queue.queue) == 0 {
			break
		}
		chunk := write_queue.queue[0]
		if chunk.SequenceID == uint64(write_queue.lastDump+1) {
			log.WithFields(log.Fields{
				"at":       "WriteQueue.Dump",
				"sequence": chunk.SequenceID,
				"socket":   chunk.SocketID,
				"session":  chunk.SessionID,
			}).Debug("writing out chunk data")
			write_queue.queue = write_queue.queue[1:]
			_, err := write_queue.destination.Write(chunk.Data)
			if chunk.Close {
				log.WithFields(log.Fields{
					"sequence": chunk.SequenceID,
					"socket":   chunk.SocketID,
					"session":  chunk.SessionID,
					"data_len": len(chunk.Data),
				}).Debug("close chunk")
				write_queue.bail(chunk.SocketID)
				return
			}
			if err != nil {
				log.WithFields(log.Fields{
					"at":    "WriteQueue.Dump",
					"error": err.Error(),
				}).Warn("error writing data out")
				if reporter, ok := FailedSocketOuts[chunk.SocketID]; ok {
					reporter <- true
				} else {
					log.WithFields(log.Fields{
						"at":        "WriteQueue.dump",
						"socket_id": chunk.SocketID,
					}).Error("unable to lookup fail socket out channel")
				}
			}
			write_queue.lastDump = write_queue.lastDump + 1
		} else {
			break
		}
	}
}

func (write_queue *WriteQueue) bail(socket_id string) {
	write_queue.destination.Close()
	swqMux.Lock()
	delete(server_write_queues, socket_id)
	swqMux.Unlock()
	cwqMux.Lock()
	delete(client_write_queues, socket_id)
	cwqMux.Unlock()
}
