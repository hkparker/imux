package server

import (
	"fmt"
	"github.com/hkparker/TLJ"
	"github.com/hkparker/imux/lib/common"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"strconv"
	"syscall"
	"time"
)

var user_clients = make(map[string]tlj.Client)
var good_nonce = make(map[string]string)
var hooked_up = make(map[string]bool)

func ForkUserProc(nonce, username string) {
	client_created := make(chan bool, 1)
	listening := make(chan bool, 1)
	ipc_filename := "/tmp/multiplexity_" + nonce
	account, _ := user.Lookup(username)
	uid, _ := strconv.Atoi(account.Uid)
	gid, _ := strconv.Atoi(account.Gid)
	go func() {
		ipc, err := net.Listen("unix", ipc_filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = os.Chown(ipc_filename, uid, gid)
		if err != nil {
			fmt.Println(err)
		}
		listening <- true
		control_socket, err := ipc.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		type_store := common.BuildTypeStore()
		client := tlj.NewClient(control_socket, type_store, false)
		user_clients[username] = client
		client_created <- true
	}()

	<-listening
	cmd := exec.Command("imux", fmt.Sprintf("--session=%s", ipc_filename))
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	cmd.Stdout = os.Stdout
	cmd.Start()
	<-client_created
}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := common.BuildTypeStore()
	server := tlj.NewServer(listener, common.TagSocketAll, type_store)
	server.AcceptRequest(
		"all",
		reflect.TypeOf(common.AuthRequest{}),
		func(iface interface{}, context tlj.TLJContext) {
			if auth_request, ok := iface.(*common.AuthRequest); ok {
				if common.Login(auth_request.Username, auth_request.Password) {
					nonce, err := common.NewNonce()
					if err == nil {
						server.Tags[context.Socket] = append(server.Tags[context.Socket], "control")
						server.Sockets["control"] = append(server.Sockets["control"], context.Socket)
						user_tag := fmt.Sprintf("user:%s", auth_request.Username)
						server.Tags[context.Socket] = append(server.Tags[context.Socket], user_tag)
						server.Sockets[user_tag] = append(server.Sockets[user_tag], context.Socket)
						ForkUserProc(nonce, auth_request.Username)
						good_nonce[nonce] = auth_request.Username
						context.Respond(common.Message{
							String: nonce,
						})
					}
				} else {
					time.Sleep(3 * time.Second)
					context.Respond(common.Message{
						String: "failed",
					})
				}
			}
		},
	)

	server.Accept(
		"all",
		reflect.TypeOf(common.WorkerAuth{}),
		func(iface interface{}, context tlj.TLJContext) {
			if worker_ready, ok := iface.(*common.WorkerAuth); ok {
				if _, ok := good_nonce[worker_ready.Nonce]; ok {
					server.TagSocket(context.Socket, "worker")
					server.TagSocket(context.Socket, worker_ready.Nonce)
					if username, ok := good_nonce[worker_ready.Nonce]; ok {
						server.TagSocket(context.Socket, "user:"+username)
					}
				}
			}
		},
	)

	server.AcceptRequest(
		"control",
		reflect.TypeOf(common.Command{}),
		func(iface interface{}, context tlj.TLJContext) {
			if command, ok := iface.(*common.Command); ok {
				username := common.UsernameFromTags(server.Tags[context.Socket])
				if client, ok := user_clients[username]; ok {
					req, err := client.Request(command)
					if err != nil {
						fmt.Println(err)
					}
					if command.Command == "exit" {
						delete(user_clients, username)
					}
					req.OnResponse(reflect.TypeOf(common.Message{}), func(iface interface{}) {
						if message, cast := iface.(*common.Message); cast {
							context.Respond(message)
						}
					})
				}
			}
		},
	)

	server.Accept(
		"worker",
		reflect.TypeOf(common.TransferChunk{}),
		func(iface interface{}, context tlj.TLJContext) {
			if chunk, ok := iface.(*common.TransferChunk); ok {
				log.Println("chunk seen in server", chunk.Destination)
				username := common.UsernameFromTags(server.Tags[context.Socket])
				if client, ok := user_clients[username]; ok {
					client.Message(chunk)
				}
			}
		},
	)

	return server
}
