//package multiplexity
package main

import "crypto/tls"

type IMUXSocket struct {
	Socket tls.Conn
	Manager IMUXManager
	LastSpeed float64
	Recycle bool
}


func (imuxsocket *IMUXSocket) Open() int {
	return 0
}

func (imuxsocket *IMUXSocket) Recieve() int {	// server is provided by manager  imuxsocket.Manager.IMUXServer?
	return 0
}

func (imuxsocket *IMUXSocket) Download(buffer Buffer) int {
	// defer a send to a complete channel?
	for {
		cmd_slice = make([]byte, 32)
		imuxsocket.Socket.Read(cmd_slice)	// break if err
		total := 0
		for _, data := range cmd_slice {	// iter?
			total += data
		}
		done := false
		if data == 0 {
			break
		} else {
			colon := 0
			for iter, data := range cmd_slice {
				if data == 58 {
					colon = iter
					break
				}
			}
			id_bytes = cmd_slice[:colon]
			size_bytes = cmd_slice[colon:]
			// assign id as first int, size as second int
			// recieve size as data
			// create chunk, send it to the buffer
		}
	}
	return 0
}

func (imuxsocket *IMUXSocket) Upload(queue ReadQueue) int {
	// serve chunks from queue to imuxsocket.Socket (a read queue is created by the manager and passed in)
	// "<32-len(id:size) zeros>id:size"
	return 0
}

func (imuxsocket *IMUXSocket) Close() int {
	return 0
}
