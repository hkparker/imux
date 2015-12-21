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
	//config.CipherSuites = []uint16{
	//	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	//	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256}
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

// Authenticate user and setup session process running as user
//func ProcessSessionRequest(conn net.Conn) {
//	log.Printf("successful login as %s from %s", username, conn.RemoteAddr())
//	conn.Write([]byte(fmt.Sprintf("%s", "Authentication successful")))

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
//	cmd := exec.Command("./Session", ipc_filename)
//	cmd.SysProcAttr = &syscall.SysProcAttr{}
//	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
//	cmd.Start()

// Pass messages
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
				if Login(auth_request.Username, auth_request.Password) {
					nonce, err := NewNonce()
					if err == nil {
						//tag as authenticated
						// store in global data tsructre
						responder.Respond(Message{
							String: nonce,
						})
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
