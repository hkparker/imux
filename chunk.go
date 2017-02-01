package imux

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
)

// A chunk represents a piece of information exchanged between a
// socket on the client side and a socket on the server size.  A
// session ID defines sockets that are part of one imux session,
// while the socket ID specifies which socket a chunk should queue
// into, ordered by the Sequence ID.
type Chunk struct {
	SessionID  string
	SocketID   string
	SequenceID uint64
	Data       []byte
}

// TLJ code to unpack Chunk data into an interface
func BuildChunk(data []byte, _ tlj.TLJContext) interface{} {
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
