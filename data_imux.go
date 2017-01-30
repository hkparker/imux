package imux

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

type DataIMUX struct {
	Chunks    chan Chunk
	Stale     chan Chunk
	SessionID string
}

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

func (data_imux *DataIMUX) ReadFrom(id string, conn io.Reader, session_id string, chunk_size_mode string) {
	log.WithFields(log.Fields{
		"at":              "DataIMUX.ReadFrom",
		"socket_id":       id,
		"chunk_size_mode": chunk_size_mode,
	}).Debug("reading from new data source")
	sequence := uint64(1)
	for {
		chunk_data := make([]byte, GetChunkSize(chunk_size_mode, session_id))
		read, err := conn.Read(chunk_data)
		log.WithFields(log.Fields{
			"at":        "DataIMUX.ReadFrom",
			"socket_id": id,
			"size":      read,
		}).Debug("read data from data source")
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
				}).Error("error reading data from imux data source")
			}
			return
		}
		data_imux.Chunks <- Chunk{
			SequenceID: sequence,
			SocketID:   id,
			SessionID:  data_imux.SessionID,
			Data:       chunk_data,
		}
		log.WithFields(log.Fields{
			"at":        "DataIMUX.ReadFrom",
			"socket_id": id,
			"size":      read,
		}).Debug("write chunk from data source")
		sequence += 1
	}
}
