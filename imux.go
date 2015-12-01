package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/hkparker/TLJ"
	"log"
	"os"
	"os/user"
	"reflect"
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
	sig := conn.ConnectionState().PeerCertificates[0].Signature
	sha := sha256.Sum256(sig)
	str := hex.EncodeToString(sha[:])
	return str
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

func TrustDialog(hostname, signature string) (bool, bool) {
	fmt.Println(fmt.Sprintf(
		"%s presents certificate with signature:\n%s",
		hostname,
		signature,
	))
	fmt.Println("[A]bort, [C]ontinue without saving, [S]ave and continue?")
	connect := false
	save := false
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, _ := stdin.ReadString('\n')
		text := strings.TrimSpace(line)
		if text == "A" {
			break
		} else if text == "C" {
			connect = true
			break
		} else if text == "S" {
			connect = true
			save = true
			break
		}
	}
	return connect, save
}

func MitMWarning(new_signature, old_signature string) (bool, bool) {
	fmt.Println(
		"WARNING: Remote certificate has changed!!\nold: %s\nnew: %s",
		old_signature,
		new_signature,
	)
	fmt.Println("[A]bort, [C]ontinue without updating, [U]pdate and continue?")
	connect := false
	update := false
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, _ := stdin.ReadString('\n')
		text := strings.TrimSpace(line)
		if text == "A" {
			break
		} else if text == "C" {
			connect = true
			break
		} else if text == "U" {
			connect = true
			update = true
			break
		}
	}
	return connect, update
}

func CreateClient(hostname string, port int) (tlj.Client, error) {
	known_hosts := LoadKnownHosts()
	conn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", hostname, port),
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
				AppendHost(hostname, signature)
			}
		}
	} else {
		connect, save_cert := TrustDialog(hostname, signature)
		if !connect {
			return tlj.Client{}, errors.New("TLS certificate rejected by user")
		} else if save_cert {
			AppendHost(hostname, signature)
		}
	}

	type_store := BuildTypeStore()
	client := tlj.NewClient(conn, &type_store)
	return client, nil
}

func BuildWorkers(hostname string, port int, networks map[string]int, nonce string) []tlj.Client {
	built := make(chan bool)
	created := make(chan tlj.Client)
	workers := make([]tlj.Client, 0)
	for local_bind, count := range networks {
		for i := 0; i < count; i++ {
			go func() {
				conn, err := tls.Dial(
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{},
				)
				fmt.Println(local_bind)
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
						fmt.Println(chunk)
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
				workers = append(workers, <-created)
				// print updated socket build status
			} else {
				// print updated socket build status
			}
		}
	}
	// print everything finished in n minutes
	return workers
}

func CommandLoop(control tlj.Client, workers []tlj.Client) {
	// read command from command line
	// if get, send a download request to start it
	// if put, tell the workers to start Messaging (blocking for prints)
	// if exit, close up pretty
	// if none of those, send it to the server and print the response
}

func main() {
	u, _ := user.Current()
	var username = flag.String("user", u.Username, "username")
	var hostname = flag.String("host", "", "hostname")
	var port = flag.Int("port", 443, "port")
	var network_config = flag.String("networks", "0.0.0.0:200", "socket configuration string: <bind ip>:<count>;")
	var route = flag.Bool("route", false, "setup ip routing table")
	var reset = flag.Bool("reset", false, "reset the socket after each chunk is transferred")
	var chunk_size = flag.Int("chunksize", 5*1024*1024, "size of each file chink in byte")
	flag.Parse()
	networks := ParseNetworks(*network_config)
	if *route {
		change := SetupRouting(networks)
		defer TeardownRouting(change)
	}
	client, err := CreateClient(*hostname, *port)
	if err != nil {
		fmt.Println(err)
		return
	}
	nonce := Login(*username, client)
	workers := BuildWorkers(*hostname, *port, networks, nonce)
	fmt.Println(*reset)
	fmt.Println(*chunk_size)
	go CommandLoop(client, workers)
	<-client.Dead
	fmt.Println("control connection closed")
}
