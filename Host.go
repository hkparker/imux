//package multiplexity
package main

import (
	"crypto/tls"
	"io"
)

type Host struct {
	IP string
	Port int
	Session tls.Conn
}

func CreateHost(ip string, port int, username, password string) Host {
	host := Host{
		IP: ip,
		Port: port,
	}
}

func (host *Host) Close() int {
	// tell the session process to exit
}

func (host *Host) GetWorkingDirectory() int {
	return 0
}

func (host *Host) ChangeDirectory() int {
	return 0
}

func (host *Host) List() int {
	io.WriteString(host.Session, "ls")
	// read response from server, parse
}

func (host *Host) MakeDirectory() int {
	return 0
}

func (host *Host) Remove() int {
	return 0
}

func (host *Host) SendFile() int {
	return 0
}

func (host *Host) RecieveFile() int {
	return 0
}

func (host *Host) CreateTransferGroup() int {
	return 0
}

func (host *Host) RecieveTransferGroup() int {
	return 0
}

func (host *Host) IncreaseTransferSockets() int {
	return 0
}

func (host *Host) CloseTransferGroup() int {
	return 0
}


// create/recieve/edit?/close imux
