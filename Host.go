//package multiplexity
package main

import "crypto/tls"

type Host struct {
	IP string
	Port int
	Hostname string
	// socket to session object Peer socket
	Messages chan string
	Session tls.Conn
}

func (host Host) Open() int {
	return 0
}

func (host Host) Close() int {
	return 0
}

func (host Host) GetWorkingDirectory() int {
	return 0
}

func (host Host) ChangeDirectory() int {
	return 0
}

func (host Host) List() int {
	return 0
}

func (host Host) MakeDirectory() int {
	return 0
}

func (host Host) Remove() int {
	return 0
}

func (host Host) SetChunkSize() int {
	return 0
}

func (host Host) SetRecycling() int {
	return 0
}

func (host Host) SendFile() int {
	return 0
}

func (host Host) RecieveFile() int {
	return 0
}

func (host Host) InitiateIMUXSession() int {
	return 0
}

func (host Host) RecieveIMUXSession() int {
	return 0
}

func (host Host) IncreaseIMUXSockets() int {
	return 0
}

func (host Host) DecreaseIMUXSockets() int {
	return 0
}

func (host Host) CloseIMUXSession() int {
	return 0
}


// create/recieve/edit?/close imux
