package imux

import (
	"github.com/hkparker/tlj"
	"io"
	"reflect"
	"time"
)

type IMUXSocket struct {
	IMUXer   DataIMUX
	Redailer Redailer
}

func (imux_socket *IMUXSocket) init() {
	for {
		socket := imux_socket.Redailer()
		writer := tlj.NewStreamWriter(socket, type_store(), reflect.TypeOf(Chunk{}))
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
