package imux

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

var MaxChunkDataSize = 16384

// A DataIMUX will read data from multiple io.Readers and chunk the data
// into a chunk chan.  The Stale attribute provides a way to insert chunks
// back into the chan from external sources.
type DataIMUX struct {
	Chunks    chan Chunk
	Stale     chan Chunk
	SessionID string
}

// Create a new DataIMUX for a given session
func NewDataIMUX(session_id string) DataIMUX {
	log.WithFields(log.Fields{
		"at":         "NewDataIMUX",
		"session_id": session_id,
	}).Debug("creating data imux")
	return DataIMUX{
		Chunks:    make(chan Chunk, 10),
		Stale:     make(chan Chunk, 50),
		SessionID: session_id,
	}
}

// Read from a new data source in this DataIMUX, create chunks from it tagged with the
// provided socket ID.
func (data_imux *DataIMUX) ReadFrom(id string, conn io.Reader) {
	log.WithFields(log.Fields{
		"at":        "DataIMUX.ReadFrom",
		"socket_id": id,
	}).Debug("reading from new data source")
	data_imux.Chunks <- Chunk{
		SocketID:  id,
		SessionID: data_imux.SessionID,
		Setup:     true,
	}
	sequence := uint64(1)
	for {
		chunk_data := make([]byte, MaxChunkDataSize)
		read, err := conn.Read(chunk_data)
		log.WithFields(log.Fields{
			"at":        "DataIMUX.ReadFrom",
			"socket_id": id,
			"size":      read,
		}).Debug("read data from data source")
		chunk_data = chunk_data[:read]
		close := false
		if err != nil {
			if err == io.EOF {
				log.WithFields(log.Fields{
					"at":        "DataIMUX.ReadFrom",
					"error":     err.Error(),
					"socket_id": id,
				}).Debug("EOF from data imux source")
			} else {
				log.WithFields(log.Fields{
					"at":        "DataIMUX.ReadFrom",
					"error":     err.Error(),
					"socket_id": id,
				}).Debug("error reading data from imux data source")
			}
			close = true
		}
		data_imux.Chunks <- Chunk{
			SequenceID: sequence,
			SocketID:   id,
			SessionID:  data_imux.SessionID,
			Data:       chunk_data,
			Close:      close,
		}
		log.WithFields(log.Fields{
			"at":        "DataIMUX.ReadFrom",
			"socket_id": id,
			"size":      read,
		}).Debug("write chunk from data source")
		sequence += 1
		if close {
			return
		}
	}
}
