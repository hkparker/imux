package imux

type WriteQueue struct{}

func OpenWriteQueue(destination string) WriteQueue {
	return WriteQueue{}
}

func (write_queue *WriteQueue) WriteChunk(sequence_id int, data []byte) {}
