package imux

import (
	"io"
)

type IMUXSocket struct {
	Out      io.Writer
	IMUXer   DataIMUX
	Redailer Redailer
}

func (imux_socket *IMUXSocket) init() {
	//forever
	// get a new socket from redailer
	// forever
	//  read chunk from imuxer
	//  write data to out, break if errors
	// sleep
}
