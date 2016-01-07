package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/hkparker/TLJ"
	"github.com/kless/osutil/user/crypt/sha512_crypt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var user_clients = make(map[string]tlj.Client)
var good_nonce = make(map[string]string)
var hooked_up = make(map[string]bool)

func NewNonce() (string, error) {
	bytes := make([]byte, 64)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func PrepareTLSConfig(pem, key string) tls.Config {
	ca_b, _ := ioutil.ReadFile(pem)
	ca, _ := x509.ParseCertificate(ca_b)
	priv_b, _ := ioutil.ReadFile(key)
	priv, _ := x509.ParsePKCS1PrivateKey(priv_b)
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}
	config.MinVersion = tls.VersionTLS12
	config.Rand = rand.Reader
	return config
}

func LookupHashAndHeader(username string) (string, string) {
	file, err := os.Open("/etc/shadow")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split_colon := func(c rune) bool {
			return c == 58
		}
		split_dollar := func(c rune) bool {
			return c == 36
		}
		fields := strings.FieldsFunc(line, split_colon)
		if fields[0] == username {
			pw_fields := strings.FieldsFunc(fields[1], split_dollar)
			header := "$" + pw_fields[0] + "$" + pw_fields[1]
			return fields[1], header
		}
	}
	return "", ""
}

func Login(username, password string) bool {
	_, err := user.Lookup(username)
	if err != nil {
		return false
	}
	passwd_crypt := sha512_crypt.New()
	hash, header := LookupHashAndHeader(username)
	new_hash, err := passwd_crypt.Generate([]byte(password), []byte(header))
	if err != nil {
		return false
	}
	if new_hash != hash {
		return false
	}
	return true
}

func UsernameFromTags(tags []string) string {
	for _, tag := range tags {
		if len(tag) > 5 {
			if tag[:5] == "user:" {
				return tag[5:]
			}
		}
	}
	return ""
}

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
		type_store := BuildTypeStore()
		client := tlj.NewClient(control_socket, &type_store)
		user_clients[username] = client
		client_created <- true
	}()

	<-listening
	cmd := exec.Command("./session", ipc_filename)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	cmd.Stdout = os.Stdout
	cmd.Start()
	<-client_created
}

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func HookupChunkAction(server tlj.Server, nonce, username string, user_clients map[string]tlj.Client) {
	if ok, _ := hooked_up[nonce]; ok {
		return
	}
	hooked_up[nonce] = true
	server.Accept(
		nonce,
		reflect.TypeOf(TransferChunk{}),
		func(iface interface{}) {
			if chunk, ok := iface.(*TransferChunk); ok {
				if client, ok := user_clients[username]; ok {
					client.Message(chunk)
				}
			}
		},
	)
}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := BuildTypeStore()
	server := tlj.NewServer(listener, TagSocketAll, &type_store)
	server.AcceptRequest(
		"all",
		reflect.TypeOf(AuthRequest{}),
		func(iface interface{}, responder tlj.Responder) {
			if auth_request, ok := iface.(*AuthRequest); ok {
				if Login(auth_request.Username, auth_request.Password) {
					nonce, err := NewNonce()
					if err == nil {
						server.Tags[responder.Socket] = append(server.Tags[responder.Socket], "control")
						server.Sockets["control"] = append(server.Sockets["control"], responder.Socket)
						user_tag := fmt.Sprintf("user:%s", auth_request.Username)
						server.Tags[responder.Socket] = append(server.Tags[responder.Socket], user_tag)
						server.Sockets[user_tag] = append(server.Sockets[user_tag], responder.Socket)
						ForkUserProc(nonce, auth_request.Username)
						good_nonce[nonce] = auth_request.Username
						responder.Respond(Message{
							String: nonce,
						})
					}
				} else {
					time.Sleep(3 * time.Second)
					responder.Respond(Message{
						String: "authentication failed",
					})
				}
			}
		},
	)

	server.AcceptRequest(
		"all",
		reflect.TypeOf(WorkerReady{}),
		func(iface interface{}, responder tlj.Responder) {
			if worker_ready, ok := iface.(*WorkerReady); ok {
				if username, ok := good_nonce[worker_ready.Nonce]; ok {
					server.Tags[responder.Socket] = append(server.Tags[responder.Socket], worker_ready.Nonce)
					server.Sockets[worker_ready.Nonce] = append(server.Sockets[worker_ready.Nonce], responder.Socket)
					HookupChunkAction(server, worker_ready.Nonce, username, user_clients)
					// create the chunk distributor if needed
					// 	for
					// 		read from the chunk channel
					// 		responder.Respond with the chunk
					// 		break if there was an error
				}
			}
		},
	)

	server.AcceptRequest(
		"control",
		reflect.TypeOf(Command{}),
		func(iface interface{}, responder tlj.Responder) {
			if command, ok := iface.(*Command); ok {
				username := UsernameFromTags(server.Tags[responder.Socket])
				if client, ok := user_clients[username]; ok {
					req, err := client.Request(command)
					if err != nil {
						fmt.Println(err)
					}
					if command.Command == "exit" {
						delete(user_clients, username)
					}
					req.OnResponse(reflect.TypeOf(Message{}), func(iface interface{}) {
						if message, cast := iface.(*Message); cast {
							responder.Respond(message)
						}
					})
					// if command.Command == "get" {
					//	req := ...
					//	req.OnResponse(reflect.TypeOf(TransferChunk{}), func(iface interface{}) {
					//		if chunk, cast := iface.(*TransferChunk); cast {
					//			chunk_distributor[nonce] <- chunk
					//		}
					//	})
					//	// on response message, send that back down the socket
					//}
				}
			}
		},
	)

	//server.AcceptRequest(
	//	"",
	//	reflect.TypeOf(),
	//	func(iface interface{}, responder tlj.Responder) {
	//	},
	//)

	return server
}

func main() {
	var listen = flag.String("listen", "0.0.0.0", "address to listen on")
	var port = flag.Int("port", 443, "port to listen on")
	//var daemon = flag.Bool("daemon", false, "run the server in the background")
	var cert = flag.String("cert", "ca.pem", "pem file with certificate to present")
	var key = flag.String("key", "ca.key", "pem file with key for certificate")
	flag.Parse()

	if current_user, _ := user.Current(); current_user.Uid != "0" {
		log.Fatal("Server must run as root.")
	}
	log_file, err := os.OpenFile("/var/log/multiplexity.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("can't open log")
		return
	}
	defer log_file.Close()
	log.SetOutput(log_file)
	log.Println("starting imux server")

	config := PrepareTLSConfig(*cert, *key)
	address := fmt.Sprintf(
		"%s:%d",
		*listen,
		*port,
	)
	listener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		fmt.Println("error starting server: %s", err)
		return
	}

	server := NewTLJServer(listener)

	// Daemonize

	err = <-server.FailedServer
	log.Println("server closed: %s", err)
}
