package imux

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLB"
)

// A chunk represents a piece of information exchanged between a
// socket on the client side and a socket on the server size.  A
// session ID defines sockets that are part of one imux session,
// while the socket ID specifies which socket a chunk should queue
// into, ordered by the Sequence ID.
type Chunk struct {
	SessionID  string `json:"a"`
	SocketID   string `json:"b"`
	SequenceID uint64 `json:"c"`
	Data       []byte `json:"d"`
	Close      bool   `json:"e"`
}

// TLB code to unpack Chunk data into an interface
func buildChunk(data []byte, _ tlb.TLBContext) interface{} {
	chunk := &Chunk{}
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "BuildChunk",
			"error": err.Error(),
		}).Error("error unmarshaling chunk data")
		return nil
	}
	log.WithFields(log.Fields{
		"at":          "BuildChunk",
		"sequence_id": chunk.SequenceID,
		"socket_id":   chunk.SocketID,
		"session_id":  chunk.SessionID,
	}).Debug("unmarshalled chunk data")
	return chunk
}
