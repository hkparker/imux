package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLB"
	"net"
	"reflect"
)

// Tag all TLB sockets as "all"
func tag_socket(socket net.Conn, server *tlb.Server) {
	log.WithFields(log.Fields{
		"at": "tag_socket",
	}).Debug("accepted new socket")
	server.TagSocket(socket, "all")
}

// Create a TLB type store for only chunks
func type_store() tlb.TypeStore {
	type_store := tlb.NewTypeStore()
	type_store.AddType(
		reflect.TypeOf(Chunk{}),
		reflect.TypeOf(&Chunk{}),
		buildChunk,
	)
	return type_store
}
