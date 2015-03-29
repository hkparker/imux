package main

import (
	"net"
	"os"
	"io"
	"fmt"
	"strings"
	"log"
)

type Session struct {
	ControlSocket net.Conn
//	Groups map[string]TranferGroup
}

func (session *Session) SendFileList(args []string) {
	
}

func (session *Session) SendWorkingDirectory(_ []string) {
	//current_directory, _ := os.Getwd()
	//io.WriteString(session.ControlSocket, current_directory)
	io.WriteString(session.ControlSocket, "pwd?")
}

func (session *Session) ChangeDirectory(args []string) {
	directory := args[0]
	err := os.Chdir(directory)
	if err == nil {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Changed directory to %s", directory)))
	} else {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Error changing directory: %s", err)))
	}
}

func (session *Session) CreateDirectory(args []string) {
	
}

func (session *Session) Remove(args []string) {
	
}

func (session *Session) CreateIMUXSession(args []string) {
	
}

func (session *Session) RecieveIMUXSession(args []string) {
	
}

func (session *Session) CloseIMUXSession(args []string) {
	
}

func (session *Session) SetChunkSize(args []string) {
	
}

func (session *Session) SetRecycling(args []string) {
	
}

func (session *Session) SendFile(args []string) {
	
}

func (session *Session) RecieveFile(args []string) {
	
}

func (session *Session) Close(args []string) {
	
}


func main() {
	// get the socket name from the arg
	ipc_file := os.Args[1]
	// parse uuid here
	control_socket, err := net.Dial("unix", ipc_file)
	defer control_socket.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("creating new session with %s", ipc_file)
	
	// Commands the session can run on this system
	commands := map[string]func(*Session, []string) {
	//	"ls": (*Session).SendFileList,
		"pwd": (*Session).SendWorkingDirectory,
		"cd": (*Session).ChangeDirectory,
	//	"mkdir": (*Session).MakeDirectory,
	//	"rm": (*Session).Remove,
	//	"createsession": (*Session).CreateSession,
	//	"recievesession": (*Session).RecieveSession,
	//	"closesession": (*Session).CloseSession,
	//	"updatechunk": (*Session).UpdateChunk,
	//	"updaterecycle": (*Session).UpdateRecyce,
	//	"sendfile": (*Session).SendFile,
	//	"recievefile": (*Session).RecieveFile,
	//	"close": (*Session).Close,
	}
	
	// create a session struct
	session := Session{}
	//session.Groups = make(map[string]TranferGroup)
	session.ControlSocket = control_socket
	
	for {
		message := make([]byte, 1024)
		n, _ := session.ControlSocket.Read(message)
		command := string(message[:n])
		command_fields := strings.Fields(command)
		if function, exists := commands[command_fields[0]]; exists {
			if len(command_fields) == 1 {
				function(&session, nil)
			} else {
				function(&session, command_fields[1:])
			}
		}
	}
}
