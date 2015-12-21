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
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/user"
	"reflect"
	"strings"
	"time"
)

func PrintProgress(completed_files, statuses, finished chan string) {
	last_status := ""
	last_len := 0
	for {
		select {
		case completed_file := <-completed_files:
			fmt.Printf("\r")
			line := "completed: " + completed_file
			fmt.Print(line)
			print_len := len(line)
			trail_len := last_len - print_len
			if trail_len > 0 {
				for i := 0; i < trail_len; i++ {
					fmt.Print(" ")
				}
			}
			fmt.Print("\n" + last_status)
		case status := <-statuses:
			last_status = status
			fmt.Printf("\r")
			fmt.Print(status)
			print_len := len(status)
			trail_len := last_len - print_len
			if trail_len > 0 {
				for i := 0; i < trail_len; i++ {
					fmt.Print(" ")
				}
			}
			last_len = print_len
		case elapsed := <-finished:
			fmt.Println("\n" + elapsed)
			return
		}
	}
}

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
	networks := make(map[string]int)
	networks["0.0.0.0"] = 200
	return networks
}

func ReadPassword() string {
	fmt.Print("Password: ")
	password_bytes, _ := terminal.ReadPassword(0)
	fmt.Println()
	return strings.TrimSpace(string(password_bytes))
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
				AppendHost(conn.RemoteAddr().String(), signature)
			}
		}
	} else {
		connect, save_cert := TrustDialog(hostname, signature)
		if !connect {
			return tlj.Client{}, errors.New("TLS certificate rejected")
		}
		if save_cert {
			AppendHost(conn.RemoteAddr().String(), signature)
		}
	}

	type_store := BuildTypeStore()
	client := tlj.NewClient(conn, &type_store)
	return client, nil
}

func BuildWorkers(
	hostname string,
	port int,
	networks map[string]int,
	nonce string,
	reset bool,
	chunk_size int,
	resume bool,
	buffers *[]WriteBuffer,
) ([]tlj.Client, error) {
	print_progress := make(chan string)
	print_status := make(chan string)
	print_finished := make(chan string)
	go PrintProgress(print_progress, print_status, print_finished)
	built := make(chan bool)
	created := make(chan tlj.Client)
	workers := make([]tlj.Client, 0)
	total_built := 0
	total_failed := 0
	total_sockets := 0
	for _, count := range networks {
		total_sockets += count
	}
	start := time.Now()
	for _, count := range networks {
		for i := 0; i < count; i++ {
			go func() {
				conn, err := tls.Dial(
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{
						InsecureSkipVerify: true,
					},
					// need to specify local bind
					// need to check sig or let it slice based on user selection
				)
				//fmt.Println(local_bind) // then remove this
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
				req.OnResponse(reflect.TypeOf(Message{}), func(iface interface{}) {
				})
				req.OnResponse(reflect.TypeOf(Chunk{}), func(iface interface{}) { // chunk in TLJ isn't the same thing as chunk in WriteBuffer....
					if _, ok := iface.(*Chunk); ok {
						// find or build the currect buffer for this chunk
						// send this chunk to the buffer
						//if buffers[chunk.Data] == nil {
						// create a new buffer and everything
						//}
						//fmt.Println(chunk)
						// map of buffers created in main and passed in here, modified later in real time
						// need to unpack the base64 data and buld the inner chunk
					}
				})
				// on response "nonce ok", send bult and created
				built <- true
				created <- client
			}()
		}
	}
	for _, count := range networks {
		for i := 0; i < count; i++ {
			success := <-built
			if success {
				total_built += 1
				workers = append(workers, <-created)
			} else {
				total_failed += 1
			}
			print_status <- fmt.Sprintf(
				"built %d/%d transfer sockets, %d failed",
				total_built,
				total_sockets,
				total_failed,
			)
		}
	}
	duration := time.Since(start)
	elapsed := fmt.Sprintf(
		"%dm%ds",
		int(duration.Minutes()),
		int(duration.Seconds()),
	)
	print_finished <- fmt.Sprintf(
		"%d/%d transfer sockets built, %d failed in %s",
		total_built,
		total_sockets,
		total_failed,
		elapsed,
	)
	if total_failed == total_sockets {
		return workers, errors.New("all transfer sockets failed to build")
	}
	return workers, nil
}

func CommandLoop(control tlj.Client, workers []tlj.Client) {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("imux> ")
		line, _ := stdin.ReadString('\n')
		text := strings.TrimSpace(line)
		// breakup by whitespace
		if text == "get" {
			// send a Command{} with get and the files as args (server wont respond, will just stream chunks down nonced workers)
			// start PrintProgress() (returning with it, when it finishes blocking)
		} else if text == "put" {
			// tell the workers to start messaging a chunk (server knowns what to do)
			// start PrintProgress() (returning with it, when it finishes blocking)
		} else if text == "exit" {
			control.Dead <- errors.New("user exit")
			break
		} else {
			req, err := control.Request(Command{
				Command: text,
				Args:    make([]string, 0),
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
	var username = flag.String("user", u.Username, "username")
	var hostname = flag.String("host", "", "hostname")
	var port = flag.Int("port", 443, "port")
	var network_config = flag.String("networks", "0.0.0.0:200", "socket configuration string: <bind ip>:<count>;")
	var route = flag.Bool("route", false, "setup ip routing table")
	var reset = flag.Bool("reset", false, "reset the socket after each chunk is transferred")
	var resume = flag.Bool("resume", false, "resume transfers if a part of the file already exists on the destination")
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
	buffers := make([]WriteBuffer, 0)
	workers, err := BuildWorkers(*hostname, *port, networks, nonce, *reset, *chunk_size, *resume, &buffers)
	if err != nil {
		fmt.Println(err)
		return
	}
	go CommandLoop(client, workers)
	err = <-client.Dead
	fmt.Println("control connection closed:", err)
}
