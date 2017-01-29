package imux

import (
	"github.com/hkparker/TLJ"
	"reflect"
	"time"
)

type IMUXSocket struct {
	IMUXer   DataIMUX
	Redialer Redialer
}

func (imux_socket *IMUXSocket) init() {
	for {
		socket, err := imux_socket.Redialer()
		if err != nil {
		}
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
