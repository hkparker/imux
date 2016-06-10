package session

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hkparker/TLJ"
	"github.com/hkparker/imux/lib/common"
	"io/ioutil"
	"log"
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
	return "\nimux by hkparker\n" +
		"\n\tFilesystem:\n\n" +
		"\tls [directory || .]\n" +
		"\tcd [directory || " + directory + "]\n" +
		"\tpwd\n" +
		"\tmkdir [dir < dir...>]\n" +
		"\trm [item < item...>]\n" +
		"\n\tTransfer:\n\n" +
		"\tget [item < item...>]\n" +
		"\tput [item < item...>]\n" +
		"\n\tOther:\n\n" +
		"\thelp\n" +
		"\texit\n"
}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := common.BuildTypeStore()
	server := tlj.NewServer(listener, common.TagSocketAll, type_store)
	server.AcceptRequest(
		"peer",
		reflect.TypeOf(common.Command{}),
		func(iface interface{}, context tlj.TLJContext) {
			if command, ok := iface.(*common.Command); ok {
				if command.Command == "exit" {
					os.Exit(0)
				} else if command.Command == "get" {
					//file_list := ParseFileList(command.Args)
					//responder.Respond(Message{}) // file names and sizes... already part of previous return?
					//chunks := CreatePooledChunkChan(file_list)
					// send the requested files down the responder as chunks
				} else {
					if function, present := commands[command.Command]; present {
						context.Respond(common.Message{
							String: function(command.Args),
						})
					} else {
						context.Respond(common.Message{
							String: "command not supported, try \"help\"",
						})
					}
				}
			}
		},
	)

	server.Accept(
		"peer",
		reflect.TypeOf(common.TransferChunk{}),
		func(iface interface{}, _ tlj.TLJContext) {
			if chunk, ok := iface.(*common.TransferChunk); ok {
				log.Println("chunk seen in session", chunk.Destination)
				// if buffers[chunk.Destination] == nil {
				// 	assign buffer to a new buffer
				//}
				//buffer.Insert(chunk)
				//if buffer.LastWrite == file size (sync controller call to declare first?) {
				// 	close the buffer
				//}
			}
		},
	)
	return server
}
