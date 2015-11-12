package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

type Session struct {
	ControlSocket net.Conn
	//	Groups map[string]TranferGroup
}

func (session *Session) WorkingDirectory(_ []string) {
	current_directory, err := os.Getwd()
	if err != nil {
		session.ControlSocket.Write([]byte(err.Error()))
	} else {
		session.ControlSocket.Write([]byte(fmt.Sprintf("%s", current_directory)))
	}
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
	directory := args[0]
	err := os.MkdirAll(directory, 0644)
	if err == nil {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Created directory %s", directory)))
	} else {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Error creating directory: %s", err)))
	}
}

func (session *Session) Remove(args []string) {
	item := args[0]
	err := os.RemoveAll(item)
	if err == nil {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Removed %s", item)))
	} else {
		session.ControlSocket.Write([]byte(fmt.Sprintf("Error removing: %s", err)))
	}
}

func (session *Session) List(args []string) {
	directory := args[0]
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		session.ControlSocket.Write([]byte(err.Error()))
		return
	}
	items := make([]Entry, 0)
	for _, f := range files {
		item := Entry{Name: f.Name(),
			Size:  f.Size(),
			Perms: f.Mode().String(),
			Mod:   f.ModTime().Format("01/02/2006 3:04PM"),
		}
		items = append(items, item)
	}
	contents, _ := json.Marshal(items)
	session.ControlSocket.Write([]byte(contents))
}

//func (session *Session) CreateIMUXSession(args []string) {

//}

//func (session *Session) RecieveIMUXSession(args []string) {

//}

//func (session *Session) CloseIMUXSession(args []string) {

//}

//func (session *Session) SetChunkSize(args []string) {

//}

//func (session *Session) SetRecycling(args []string) {

//}

//func (session *Session) SendFile(args []string) {

//}

//func (session *Session) RecieveFile(args []string) {

//}

func (session *Session) Close(_ []string) {
	// close each transfer group
	session.ControlSocket.Write([]byte("exiting session"))
	os.Exit(0)
}

func main3() {
	// get the socket name from the first arg
	ipc_file := os.Args[1]
	control_socket, err := net.Dial("unix", ipc_file)
	defer control_socket.Close()
	if err != nil {
		log.Fatal(err)
	}

	// commands the session can run on this system
	commands := map[string]func(*Session, []string){
		"close": (*Session).Close,
		"pwd":   (*Session).WorkingDirectory,
		"cd":    (*Session).ChangeDirectory,
		"mkdir": (*Session).CreateDirectory,
		"ls":    (*Session).List,
		"rm":    (*Session).Remove,
		//	"createsession": (*Session).CreateSession,
		//	"recievesession": (*Session).RecieveSession,
		//	"closesession": (*Session).CloseSession,
		//	"updatechunk": (*Session).UpdateChunk,
		//	"updaterecycle": (*Session).UpdateRecyce,
		//	"sendfile": (*Session).SendFile,
		//	"recievefile": (*Session).RecieveFile,
	}

	// create a session struct
	session := Session{}
	//session.Groups = make(map[string]TranferGroup)
	session.ControlSocket = control_socket

	for {
		command, err := ReadLine(session.ControlSocket)
		if err != nil {
			return
		}
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
