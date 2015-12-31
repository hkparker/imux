package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hkparker/TLJ"
	"github.com/kless/osutil/user/crypt/sha512_crypt"
	"net"
	"reflect"
	//"github.com/twinj/uuid"
	"io/ioutil"
	"log"
	//"net"
	"flag"
	"os"
	//"os/exec"
	"os/user"
	//"strconv"
	"strings"
	//"syscall"
	"encoding/base64"
	"time"
)

type CommandRunner func([]string) string

// move these all back into sessions.go to run as another process
var commands = map[string]CommandRunner{
	"ls":    ListFiles,
	"cd":    ChangeDirectory,
	"pwd":   PrintWorkingDirectory,
	"mkdir": CreateDirectory,
	"rm":    Remove,
	"exit":  Close,
	//"get": StreamChunks,
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
			"%s user:group %d\t%s %s",
			f.Mode().String(),
			f.Size(), // human format?
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
		directory = "."
	}
	err := os.Chdir(directory)
	if err == nil {
		return fmt.Sprintf("Changed directory to %s", directory)
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

func Close(_ []string) string {
	os.Exit(0)
	return ""
}

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

func Login(username, password string, server tlj.Server, socket net.Conn) bool {
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
	server.Tags[socket] = append(server.Tags[socket], "control")
	server.Sockets["control"] = append(server.Sockets["control"], socket)
	return true
}

// Authenticate user and setup session process running as user
//func ProcessSessionRequest(conn net.Conn) {

// Create a unix socket to pass commands from client to user process
//	uuid.SwitchFormat(uuid.CleanHyphen)
//	ipc_filename := "/tmp/multiplexity_" + uuid.NewV4().String()
//	ipc, err := net.Listen("unix", ipc_filename)
//	defer ipc.Close()
//	defer os.RemoveAll(ipc_filename)
//	uid, _ := strconv.Atoi(account.Uid)
//	gid, _ := strconv.Atoi(account.Gid)
//	os.Chown(ipc_filename, uid, gid)

// Create new process running under authenticated user's account
//	cmd := exec.Command("./session", ipc_filename)
//	cmd.SysProcAttr = &syscall.SysProcAttr{}
//	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
//	cmd.Start()

// Pass messages

// should use io.Copy to pass raw data to a TLJ server inside session
// this means I have no control over what types of structs get sent to session
// this is fine, since at this point everything is under session's level of permission
// but what about the TLJ server already attached to this socket.  need server.Detach(socket)

//	ipc_session, err := ipc.Accept()
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	ipc_session.Write([]byte(fmt.Sprintf("cd %s", account.HomeDir)))
//	ReadBytes(ipc_session)
//
//	for {
//		bytes, err := ReadBytes(conn)
//		if err != nil {
//			log.Println(err)
//			break
//		}
//		ipc_session.Write(bytes)
//		bytes, err = ReadBytes(ipc_session)
//		if err != nil {
//			log.Println(err)
//			break
//		}
//		conn.Write(bytes)
//	}
//}

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func NewTLJServer(listener net.Listener) tlj.Server {
	type_store := BuildTypeStore()
	server := tlj.NewServer(listener, TagSocketAll, &type_store)
	server.AcceptRequest(
		"all",
		reflect.TypeOf(AuthRequest{}),
		func(iface interface{}, responder tlj.Responder) {
			if auth_request, ok := iface.(*AuthRequest); ok {
				if Login(auth_request.Username, auth_request.Password, server, responder.Socket) {
					nonce, err := NewNonce()
					if err == nil {
						//tag as authenticated
						// store in global data tsructre
						responder.Respond(Message{
							String: nonce,
						})
						// now we pass all incoming messages from responder.Socket directly over IPC
						// server.Detach(responder.Socket)
						// ExchangeData(responder.Socket, NewSessionAs(auth_request.username))
					}
				} else {
					time.Sleep(3 * time.Second)
					responder.Respond(Message{
						String: "failed",
					})
				}
			}
		},
	)

	server.AcceptRequest(
		"all",
		reflect.TypeOf(WorkerReady{}),
		func(iface interface{}, responder tlj.Responder) {
			// repond with an OK
			// any time a chunk needs to come down, repond with it (create a chan in the global namespace if missing, read from that?)
		},
	)

	server.AcceptRequest(
		"control",
		reflect.TypeOf(Command{}),
		func(iface interface{}, responder tlj.Responder) {
			if command, ok := iface.(*Command); ok {
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
	err = <-server.FailedServer
	log.Println("server closed: %s", err)
}
