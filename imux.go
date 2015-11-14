package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/hkparker/TLJ"
	"log"
	"net"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
)

func LoadKnownHosts() map[string]string {
	sigs := make(map[string]string)
	filename := os.Getenv("HOME") + "/.multiplexity/known_hosts"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.Create(filename)
		return sigs
	}
	known_hosts, err := os.Open(filename)
	defer known_hosts.Close()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(known_hosts)
	for scanner.Scan() {
		contents := strings.Split(scanner.Text(), " ")
		sigs[contents[0]] = contents[1]
	}
	return sigs
}

func AppendHost(hostname string, signature string) {
	filename := os.Getenv("HOME") + "/.multiplexity/known_hosts"
	known_hosts, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	known_hosts.WriteString(hostname + " " + signature + "\n")
	known_hosts.Close()
}

func SHA256Sig(conn *tls.Conn) string {
	signature := sha256.Sum256(conn.ConnectionState().PeerCertificates[0].Signature)
	var characters bytes.Buffer
	for i, chr := range signature {
		characters.WriteString(strconv.Itoa(int(chr)))
		if i != 31 {
			characters.WriteString(":")
		}
	}
	return characters.String()
}

func ParseNetworks(data string) map[string]int {
	return make(map[string]int)
}

func ReadPassword() string {
	// print password: then disable echo and get input
	return ""
}

func Login(username string, client tlj.Client) string {
	for {
		password := ReadPassword()
		auth_request := AuthRequest{
			Username: username,
			Password: password,
		}
		req, _ := client.Request(auth_request)
		resp_chan := make(chan string)
		req.OnResponse(reflect.TypeOf(Message{}), func(iface interface{}) {
			if message, ok := iface.(*Message); ok {
				resp_chan <- message.String
			}
		})
		response := <-resp_chan
		if response != "failed" {
			return response
		}
	}
}

func SetupRouting(networks map[string]int) string {
	return ""
}

func TeardownRouting(change string) {

}

func TrustDialog(signature string) (bool, bool) {
	// show sig, options are to
	// abort connection
	// continue but dont save
	// continue and save
	return true, true
}

func MitMWarning(new_signature, old_signature string) (bool, bool) {
	// print new sig doesn't match old sig
	// select to abort, continue but dont save, continue and save
	return true, true
}

func CreateClient(ip string, port int) (tlj.Client, error) {
	known_hosts := LoadKnownHosts()
	conn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", ip, port),
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return tlj.Client{}, err
	}
	signature := SHA256Sig(conn)

	if saved_signature, present := known_hosts[conn.RemoteAddr().String()]; present {
		if signature != saved_signature {
			connect, update := MitMWarning(signature, saved_signature)
			if !connect {
				return tlj.Client{}, errors.New("TLS certificate mismatch")
			}
			if update {
				AppendHost(ip, signature)
			}
		}
	} else {
		connect, save_cert := TrustDialog(signature)
		if !connect {
			return tlj.Client{}, errors.New("TLS certificate rejected by user")
		} else if save_cert {
			AppendHost(ip, signature)
		}
	}

	type_store := BuildTypeStore()
	client := tlj.NewClient(conn, &type_store)
	return client, nil
}

func BuildWorkers(hostname string, port int, networks map[string]int, nonce string) []net.Conn {
	built := make(chan bool)
	created := make(chan net.Conn)
	workers := make([]net.Conn)
	for local_bind, count := range networks {
		for i := 0; i < count; i++ {
			go func() {
				conn, err := tls.Dial(
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{},
				)
				if err != nil {
					built <- false
					return
				}
				type_store := BuildTypeStore()
				client := tlj.NewClient(conn, &type_store)
				req, err := client.Request(WorkerReady{
					Nonce: nonce,
				})
				if err != nil {
					built <- false
					return
				}
				req.OnResponse(reflect.TypeOf(Chunk{}), func(iface interface{}) {
					if chunk, ok := iface.(*Chunk); ok {
						// find or build the currect buffer for this chunk
						// send this chunk to the buffer
					}
				})
				built <- true
				created <- client
			}()
		}
	}
	for _, count := range networks {
		for i := 0; i < count; i++ {
			success := <-built
			if success {
				workers = append(workers, <-created...)
				// print updated socket build status
			} else {
				// print updated socket build status
			}
		}
	}
	// print everything finished in n minutes
	return workers
}

func CommandLoop(client tlj.Client, workers []net.Conn) {
	// read command from command line
	// if get, send a download request to start it
	// if put, tell the workers to start Messaging (blocking for prints)
	// if exit, close up pretty
	// if none of those, send it to the server and print the response
}

func main2() {
	u, _ := user.Current()
	var username = &flag.String("user", u.Username, "username")
	var hostname = &flag.String("host", "", "hostname")
	var port = &flag.Int("port", 995, "port")
	var network_config = &flag.String("networks", "0.0.0.0:200", "socket configuration string: 0.0.0.0:200;192.168.1.3:50;")
	var route = &flag.Bool("route", false, "setup ip routing table")
	var reset = &flag.Bool("reset", false, "reset the socket after each chunk is transferred")
	var chunk_size = &flag.Int("chunksize", 5*1024*1024, "size of each file chink in byte")
	flag.Parse()
	networks := ParseNetworks(network_config)
	if route {
		change := SetupRouting(networks)
		defer TeardownRouting(change)
	}
	client, err := CreateClient(hostname, port)
	if err != nil {
		log.Println(err)
		return
	}
	nonce := Login(username, client)
	workers := BuildWorkers(hostname, port, networks, nonce)
	go CommandLoop(client, workers)
	<-client.Dead
	log.Println("control connection closed")
}
