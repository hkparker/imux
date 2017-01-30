package imux

import (
	log "github.com/Sirupsen/logrus"
)

// The global chunk size shared by all sessions on a client
var ClientChunkSize = 16384

// Servers stay aware of the largest chunk a session has sent
// and use it as the response chunk size
var ObservedMaximumChunkSizes = make(map[string]int)

// Get the correct chunk size for a given mode of operation
// (client or server) and the session ID.
func GetChunkSize(chunk_size_mode, session_id string) int {
	log.WithFields(log.Fields{
		"at":         "GetChunkSize",
		"mode":       chunk_size_mode,
		"session_id": session_id,
	}).Debug("looking up chunk size")
	if chunk_size_mode == "client" {
		return ClientChunkSize
	} else if chunk_size_mode == "server" {
		if max, ok := ObservedMaximumChunkSizes[session_id]; ok {
			return max
		}
	}
	log.WithFields(log.Fields{
		"at":         "GetChunkSize",
		"mode":       chunk_size_mode,
		"session_id": session_id,
	}).Warn("no chunk size defined for chunk")
	return 16384
}
