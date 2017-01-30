package imux

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
)

type Chunk struct {
	SessionID  string
	SocketID   string
	SequenceID uint64
	Data       []byte
}

func BuildChunk(data []byte, _ tlj.TLJContext) interface{} {
	log.WithFields(log.Fields{
		"at": "BuildChunk",
	}).Debug("unmarshaling chunk data")
	chunk := &Chunk{}
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "BuildChunk",
			"error": err.Error(),
		}).Error("error unmarshaling chunk data")
		return nil
	}
	return chunk
}
