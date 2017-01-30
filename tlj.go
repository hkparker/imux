package imux

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
)

func tag_socket(socket net.Conn, server *tlj.Server) {
	log.WithFields(log.Fields{
		"at": "tag_socket",
	}).Debug("tagging new socket")
	server.TagSocket(socket, "all")
}

func type_store() tlj.TypeStore {
	type_store := tlj.NewTypeStore()
	type_store.AddType(
		reflect.TypeOf(Chunk{}),
		reflect.TypeOf(&Chunk{}),
		BuildChunk,
	)
	return type_store
}
