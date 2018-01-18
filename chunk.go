package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLB"
	"gopkg.in/mgo.v2/bson"
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
	Close      bool
	Setup      bool
}

// TLB code to unpack Chunk data into an interface
func buildChunk(data []byte, _ tlb.TLBContext) interface{} {
	chunk := &Chunk{}
	err := bson.Unmarshal(data, &chunk)
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
