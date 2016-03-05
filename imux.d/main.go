package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hkparker/TLJ"
	"github.com/hkparker/imux"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net"
	"os"
	"os/user"
	"reflect"
	"strings"
	"sync"
	"time"
)

var username string
var hostname string
var port int
var network_config string
var resume bool
var chunk_size int

var type_store tlj.TypeStore

func commandLoop(control tlj.Client, workers []tlj.Client, chunk_size int) {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("imux> ")
		line, _ := stdin.ReadString('\n')
		text := strings.TrimSpace(line)
		cmd := strings.Fields(text)
		if len(cmd) == 0 {
			continue
		}
		command := cmd[0]
		var args []string
		if len(command) > 1 {
			args = cmd[1:]
		}
		if command == "get" {
			//imux.RequestFiles(args)
			// send a Command{} with get and the files as args (server wont respond (or does it need to respond when all done and with updates?), will just stream chunks down nonced workers)
			// file finished are send by write buffer to global current_transfer chan
			// speed updates are 1 per second?  need to ask every worker?  (workers update global speed store, sum that)
			//PrintProgress(file_finished, speed_update, all_done)
		} else if command == "put" {
			//file_list, total_bytes := ParseFileList(args)
			//uploadFiles(file_list, total_bytes, []tlj.StreamWriter{})
		} else if command == "exit" {
			control.Request(Command{
				Command: "exit",
			})
			control.Dead <- errors.New("user exit")
			break
		} else {
			req, err := control.Request(Command{
				Command: command,
				Args:    args,
			})
			if err != nil {
				go func() {
					control.Dead <- errors.New(fmt.Sprintf("error sending command: %v", err))
				}()
				break
			}
			command_output := make(chan string)
			req.OnResponse(reflect.TypeOf(Message{}), func(iface interface{}) {
				if message, ok := iface.(*Message); ok {
					command_output <- message.String
				}
			})
			fmt.Println(<-command_output)
		}
	}
}

func main() {
	u, _ := user.Current()
	username = *flag.String("user", u.Username, "username")
	hostname = *flag.String("host", "", "hostname")
	port = *flag.Int("port", 443, "port")
	network_config = *flag.String("networks", "0.0.0.0:200", "socket configuration string: <bind ip>:<count>;")
	chunk_size = *flag.Int("chunksize", 5*1024*1024, "size of each file chink in byte")
	flag.Parse()

	networks, err := imux.ParseNetworks(network_config)
	if err != nil {
		log.Fatal(err)
	}

	client, err := imux.CreateClient(hostname, port)
	if err != nil {
		log.Fatal(err)
	}

	nonce := imux.ClientLogin(username, client)
	if err != nil {
		log.Fatal(err)
	}

	streamers, err := imux.ConnectWorkers(hostname, port, networks, nonce)
	if err != nil {
		log.Fatal(err)
	}

	go commandLoop(client, streamers, chunk_size)
	err = <-client.Dead
	fmt.Println("control connection closed:", err)
}
