package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hkparker/TLJ"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"reflect"
)

type CommandRunner func([]string) string

var commands = map[string]CommandRunner{
	"ls":    ListFiles,
	"cd":    ChangeDirectory,
	"pwd":   PrintWorkingDirectory,
	"mkdir": CreateDirectory,
	"rm":    Remove,
	"help":  DisplayHelp,
}

func ListFiles(args []string) string {
	var directory string
	if len(args) > 0 {
		directory = args[0]
	} else {
		directory = "."
	}
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err.Error()
	}

	file_list := ""
	for n, f := range files {
		file_list = file_list + fmt.Sprintf(
			"%s %s\t%s %s",
			f.Mode().String(),
			humanize.IBytes(uint64(f.Size())),
			f.ModTime().Format("01/02/2006 03:04PM"),
			f.Name(),
		)
		if n != len(files)-1 {
			file_list = file_list + "\n"
		}
	}
	return file_list
}

func ChangeDirectory(args []string) string {
	var directory string
	if len(args) > 0 {
		directory = args[0]
	} else {
		account, _ := user.Current()
		directory = account.HomeDir
	}
	err := os.Chdir(directory)
	if err == nil {
		return fmt.Sprintf("changed directory to %s", directory)
	} else {
		return err.Error()
	}
}

func PrintWorkingDirectory(_ []string) string {
	current_directory, err := os.Getwd()
	if err == nil {
		return current_directory
	} else {
		return err.Error()
	}
}

func CreateDirectory(items []string) string {
	if len(items) == 0 {
		return "specify things to create"
	}
	result := ""
	for n, item := range items {
		err := os.MkdirAll(item, 0644)
		if err == nil {
			result = result + fmt.Sprintf("created %s", item)
		} else {
			result = result + fmt.Sprintf("failed to create %s: %v", item, err)
		}
		if n != len(items)-1 {
			result += "\n"
		}
	}
	return result
}

func Remove(items []string) string {
	if len(items) == 0 {
		return "specify things to remove"
	}
	result := ""
	for n, item := range items {
		err := os.RemoveAll(item)
		if err == nil {
			result = result + fmt.Sprintf("removed %s", item)
		} else {
			result = result + fmt.Sprintf("failed to remove %s: %v", item, err)
		}
		if n != len(items)-1 {
			result += "\n"
		}
	}
	return result
}

func DisplayHelp(_ []string) string {
	account, _ := user.Current()
	directory := account.HomeDir
	return "\nMultiplexity by hkparker\n" +
		"\n\tFilesystem:\n\n" +
		"\tls [directory || .]\tlist files\n" +
		"\tcd [directory || " + directory + "]\tchange directory\n" +
		"\tpwd\t\t\tprint working directory\n" +
		"\tmkdir [dir < dir...>]\tcreate directories\n" +
		"\trm [item < item...>]\tdestroy items (rm -rf)\n" +
		"\n\tTransfer:\n\n" +
		"\tget [item < item...>]\tdownload items\n" +
		"\tput [item < item...>]\tupload items\n" +
		"\n\tOther:\n\n" +
		"\thelp\t\t\tprint this message\n" +
		"\texit\t\t\tclose the client\n"
}

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func StreamChunks() {

}

func StoreChunks() {

}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := BuildTypeStore()
	server := tlj.NewServer(listener, TagSocketAll, &type_store)
	server.AcceptRequest(
		"all",
		reflect.TypeOf(Command{}),
		func(iface interface{}, responder tlj.Responder) {
			if command, ok := iface.(*Command); ok {
				if command.Command == "exit" {
					os.Exit(0)
				} else if command.Command == "get" {
					// send a message back explaining what to expect (sizes and such)
					// send the request files down the responder as chunks
				} else if command.Command == "put" {
				} else {
					if function, present := commands[command.Command]; present {
						responder.Respond(Message{
							String: function(command.Args),
						})
					} else {
						responder.Respond(Message{
							String: "command not supported, try \"help\"",
						})
					}
				}
			}
		},
	)
	return server
}

func main() {
	fmt.Println("session executed")
	ipc_file := os.Args[1]
	control_socket, err := net.Dial("unix", ipc_file)
	if err != nil {
		fmt.Println(err)
		return
	}

	discard_listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return
	}
	server := NewTLJServer(discard_listener)
	server.Insert(control_socket)
	err = <-server.FailedServer
	fmt.Println(err)
}
