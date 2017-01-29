package imux

import (
	"io"
)

type DataIMUX struct {
	Sources   []io.Reader
	Chunks    chan Chunk
	Stale     chan Chunk
	SessionID string
}

func NewDataIMUX(session_id string) DataIMUX {
	data_imux := DataIMUX{
		Sources:   make([]io.Reader, 0),
		Chunks:    make(chan Chunk, 0),
		Stale:     make(chan Chunk, 100),
		SessionID: session_id,
	}

	go data_imux.process()

	return data_imux
}

func (data_imux *DataIMUX) process() {

}

func (data_imux *DataIMUX) ReadFrom(conn io.Reader) {}
