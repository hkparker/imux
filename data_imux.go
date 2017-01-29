package imux

import (
	"io"
)

type DataIMUX struct {
	Chunks    chan Chunk
	Stale     chan Chunk
	SessionID string
}

func NewDataIMUX(session_id string) DataIMUX {
	return DataIMUX{
		Chunks:    make(chan Chunk, 10),
		Stale:     make(chan Chunk, 50),
		SessionID: session_id,
	}
}

func (data_imux *DataIMUX) ReadFrom(id string, conn io.Reader, chunk_size_mode string) {
	sequence := uint64(1)
	for {
		chunk_data := make([]byte, GetChunkSize(chunk_size_mode))
		_, err := conn.Read(chunk_data)
		if err != nil {
		}
		data_imux.Chunks <- Chunk{
			SequenceID: sequence,
			SocketID:   id,
			SessionID:  data_imux.SessionID,
			Data:       chunk_data,
		}
		sequence += 1
	}
}
