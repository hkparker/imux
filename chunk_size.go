package imux

import (
	log "github.com/Sirupsen/logrus"
)

var ClientChunkSize = 16384
var ObservedMaximumChunkSizes = make(map[string]int)

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
