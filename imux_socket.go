package imux

import (
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
	"sync"
	"time"
)

type IMUXSocket struct {
	IMUXer   DataIMUX
	Redialer Redialer
}

func (imux_socket *IMUXSocket) init() {
	tlj_server := tlj.Server{
		TypeStore:       type_store(),
		Tag:             tag_socket,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	tlj_server.Accept("all", reflect.TypeOf(Chunk{}), func(iface interface{}, context tlj.TLJContext) {
		if chunk, ok := iface.(Chunk); ok {
			if writer, ok := client_write_queues[chunk.SocketID]; ok {
				writer.Write(chunk)
			} else {
			}
		}
	})

	for {
		socket, err := imux_socket.Redialer()
		if err != nil {
		}
		tlj_server.Insert(socket)
		writer, err := tlj.NewStreamWriter(socket, type_store(), reflect.TypeOf(Chunk{}))
		if err != nil {
		}

		for {
			chunk := <-imux_socket.IMUXer.Chunks
			err := writer.Write(chunk)
			if err != nil {
				imux_socket.IMUXer.Stale <- chunk
				break
			}
		}
		time.Sleep(15 * time.Second)
	}
}
