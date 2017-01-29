package imux

import (
	"encoding/json"
	"github.com/hkparker/TLJ"
)

type Chunk struct {
	SessionID  string
	SocketID   string
	SequenceID uint64
	Data       []byte
}

func BuildChunk(data []byte, _ tlj.TLJContext) interface{} {
	chunk := &Chunk{}
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		return nil
	}
	return chunk
}
