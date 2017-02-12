package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
)

// Tag all TLJ sockets as "all"
func tag_socket(socket net.Conn, server *tlj.Server) {
	log.WithFields(log.Fields{
		"at": "tag_socket",
	}).Debug("accepted new socket")
	server.TagSocket(socket, "all")
}

// Create a TLJ type store for only chunks
func type_store() tlj.TypeStore {
	type_store := tlj.NewTypeStore()
	type_store.AddType(
		reflect.TypeOf(Chunk{}),
		reflect.TypeOf(&Chunk{}),
		buildChunk,
	)
	return type_store
}

func remoteClose(socket_id, session_id string) { // instant or after chunk n
	log.WithFields(log.Fields{
		"at":         "remoteClose",
		"socket_id":  socket_id,
		"session_id": session_id,
	}).Debug("issuing remote close")
	// Create a new SocketCloser, or chunk with Close set
	// Determine how to send this struct to the correct Session
	// remove all references to socket and write queue for socket
	// ensure future chunks are blackholed
}
