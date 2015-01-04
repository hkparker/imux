//package multiplexity
package main

import "crypto/tls"

type Session struct {
	Socket tls.Conn

	// socket to host object Peer socket
	// array of imuxmanagers
	// password?
}

func (session *Session) Process(socket tls.Conn) {
	session.Socket = socket
	commands := map[string]func(string) {
		"ls": SendFileList,
		"pwd": SendWorkingDirectory,
		"cd": ChangeWorkingDirectory,
		"mkdir": MakeDirectory,
		"rm": Remove,
		"createsession": CreateSession,
		"recievesession": RecieveSession,
		"closesession": CloseSession,
		"updatechunk": UpdateChunk,
		"updaterecycle": UpdateRecyce,
		"sendfile": SendFile,
		"recievefile": RecieveFile,
		"close": Close,
	}
	command, args := session.NextCommand()		// "["transferfile", "sourcedir destinationdir"]"
	if _, exists := commands[command]; exists {
		commands(command)(args)
	} else {
		session.Socket.Write() // Not understood
	}
}

func (session *Session) SendFileList() {
	
}

func (session *Session) SendWorkingDirectory() {
	
}

func (session *Session) ChangeDirectory() {
	
}

func (session *Session) CreateDirectory() {
	
}

func (session *Session) Remove() {
	
}

func (session *Session) CreateIMUXSession() {
	
}

func (session *Session) RecieveIMUXSession() {
	
}

func (session *Session) CloseIMUXSession() {
	
}

func (session *Session) SetChunkSize() {
	
}

func (session *Session) SetRecycling() {
	
}

func (session *Session) SendFile() {
	
}

func (session *Session) RecieveFile() {
	
}

//func (session *Session) SendDirectory() {
	
//}

//func (session *Session) RecieveDirectory() {
	
//}

func (session *Session) Close() {
	
}

//https://github.com/go-av/tls-example


