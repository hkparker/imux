package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/hkparker/TLJ"
	"github.com/hkparker/imux"
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

var listen = *flag.String("listen", "0.0.0.0", "address to listen on")
var port = *flag.Int("port", 443, "port to listen on")
var daemon = *flag.Bool("daemon", false, "run the server in the background")
var cert = *flag.String("cert", "cert.pem", "pem file with certificate to present")
var key = *flag.String("key", "key.pem", "pem file with key for certificate")

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
		type_store := imux.BuildTypeStore()
		client := tlj.NewClient(control_socket, type_store, false)
		user_clients[username] = client
		client_created <- true
	}()

	<-listening
	cmd := exec.Command("./imux-session", ipc_filename)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	cmd.Stdout = os.Stdout
	cmd.Start()
	<-client_created
}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := imux.BuildTypeStore()
	server := tlj.NewServer(listener, imux.TagSocketAll, type_store)
	server.AcceptRequest(
		"all",
		reflect.TypeOf(imux.AuthRequest{}),
		func(iface interface{}, context tlj.TLJContext) {
			if auth_request, ok := iface.(*imux.AuthRequest); ok {
				if imux.Login(auth_request.Username, auth_request.Password) {
					nonce, err := imux.NewNonce()
					if err == nil {
						server.Tags[context.Socket] = append(server.Tags[context.Socket], "control")
						server.Sockets["control"] = append(server.Sockets["control"], context.Socket)
						user_tag := fmt.Sprintf("user:%s", auth_request.Username)
						server.Tags[context.Socket] = append(server.Tags[context.Socket], user_tag)
						server.Sockets[user_tag] = append(server.Sockets[user_tag], context.Socket)
						ForkUserProc(nonce, auth_request.Username)
						good_nonce[nonce] = auth_request.Username
						context.Respond(imux.Message{
							String: nonce,
						})
					}
				} else {
					time.Sleep(3 * time.Second)
					context.Respond(imux.Message{
						String: "failed",
					})
				}
			}
		},
	)

	server.AcceptRequest(
		"all",
		reflect.TypeOf(imux.WorkerAuth{}),
		func(iface interface{}, context tlj.TLJContext) {
			if worker_ready, ok := iface.(*imux.WorkerAuth); ok {
				if _, ok := good_nonce[worker_ready.Nonce]; ok {
					// tag as a worker and with nonce
					//:server.Tags[context.Socket] = append(server.Tags[context.Socket], worker_ready.Nonce)
					//server.Sockets[worker_ready.Nonce] = append(server.Sockets[worker_ready.Nonce], context.Socket)
				}
			}
		},
	)

	server.AcceptRequest(
		"control",
		reflect.TypeOf(imux.Command{}),
		func(iface interface{}, context tlj.TLJContext) {
			if command, ok := iface.(*imux.Command); ok {
				username := imux.UsernameFromTags(server.Tags[context.Socket])
				if client, ok := user_clients[username]; ok {
					req, err := client.Request(command)
					if err != nil {
						fmt.Println(err)
					}
					if command.Command == "exit" {
						delete(user_clients, username)
					}
					req.OnResponse(reflect.TypeOf(imux.Message{}), func(iface interface{}) {
						if message, cast := iface.(*imux.Message); cast {
							context.Respond(message)
						}
					})
					// if command.Command == "get" {
					//	req.OnResponse(reflect.TypeOf(TransferChunk{}), func(iface interface{}) {
					//		if chunk, cast := iface.(*TransferChunk); cast {
					//			chunk_distributor[nonce] <- chunk  // chunks come from session IPC out worker sockets
					//		}
					//	})
					//}
				}
			}
		},
	)

	server.Accept(
		"worker",
		reflect.TypeOf(imux.TransferChunk{}),
		func(iface interface{}, context tlj.TLJContext) {
			if chunk, ok := iface.(*imux.TransferChunk); ok {
				username := imux.UsernameFromTags(server.Tags[context.Socket])
				if client, ok := user_clients[username]; ok {
					client.Message(chunk)
				}
			}
		},
	)

	return server
}

func main() {
	flag.Parse()

	if current_user, _ := user.Current(); current_user.Uid != "0" {
		log.Fatal("Server must run as root.")
	}
	log_file, err := os.OpenFile("/var/log/multiplexity.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("can't open log")
	}
	defer log_file.Close()
	log.SetOutput(log_file)
	log.Println("starting imux server")

	config := imux.PrepareTLSConfig(cert, key)
	address := fmt.Sprintf("%s:%d", listen, port)
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		log.Fatal("error starting server: %s", err)
	}

	server := NewTLJServer(listener)
	if daemon {
		// Daemonize
	}
	err = <-server.FailedServer
	log.Println("server closed: %s", err)
}
