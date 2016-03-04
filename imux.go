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
	"golang.org/x/crypto/ssh/terminal"
	"log"
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

func printProgress(completed_files, statuses, finished chan string) {
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

func loadKnownHosts() map[string]string {
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

func appendHost(hostname string, signature string) {
	filename := os.Getenv("HOME") + "/.multiplexity/known_hosts"
	known_hosts, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	known_hosts.WriteString(hostname + " " + signature + "\n")
	known_hosts.Close()
}

func sha256Sig(conn *tls.Conn) string {
	sig := conn.ConnectionState().PeerCertificates[0].Signature
	sha := sha256.Sum256(sig)
	str := hex.EncodeToString(sha[:])
	return str
}

func parseNetworks(data string) (map[string]int, error) {
	networks := make(map[string]int)
	networks["0.0.0.0"] = 2
	return networks, nil
}

func readPassword() string {
	fmt.Print("Password: ")
	password_bytes, _ := terminal.ReadPassword(0)
	fmt.Println()
	return strings.TrimSpace(string(password_bytes))
}

func clientLogin(username string, client tlj.Client) string {
	for {
		password := readPassword()
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

func trustDialog(hostname, signature string) (bool, bool) {
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

func mitmWarning(new_signature, old_signature string) (bool, bool) {
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

func createClient(hostname string, port int) (tlj.Client, error) {
	known_hosts := loadKnownHosts()
	conn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", hostname, port),
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return tlj.Client{}, err
	}
	signature := sha256Sig(conn)
	if saved_signature, present := known_hosts[conn.RemoteAddr().String()]; present {
		if signature != saved_signature {
			connect, update := mitmWarning(signature, saved_signature)
			if !connect {
				return tlj.Client{}, errors.New("TLS certificate mismatch")
			}
			if update {
				appendHost(conn.RemoteAddr().String(), signature)
			}
		}
	} else {
		connect, save_cert := trustDialog(hostname, signature)
		if !connect {
			return tlj.Client{}, errors.New("TLS certificate rejected")
		}
		if save_cert {
			appendHost(conn.RemoteAddr().String(), signature)
		}
	}

	type_store := BuildTypeStore()
	client := tlj.NewClient(conn, type_store, false)
	return client, nil
}

func connectWorkers(hostname string, port int, worker_server tlj.Server, networks map[string]int, nonce string, resume bool) (streamers []tlj.StreamWriter, err error) {
	worker_status_update_text := make(chan string)
	worker_build_finished_text := make(chan string)
	go printProgress(
		make(chan string),
		worker_status_update_text,
		worker_build_finished_text,
	)

	streamer_chan := make(chan tlj.StreamWriter)
	success_worker_count := 0
	failed_worker_count := 0
	total_worker_count := 0
	var worker_waiter sync.WaitGroup
	for _, count := range networks {
		worker_waiter.Add(count)
		total_worker_count += count
	}

	failed_worker_reporter := make(chan bool, total_worker_count)
	start := time.Now()
	for bind, count := range networks {
		for i := 0; i < count; i++ {
			go func() {
				defer worker_waiter.Done()
				dialer := bind
				conn, err := tls.Dial(
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{
						InsecureSkipVerify: true,
					},
					// need to check sig too
				)
				if err != nil {
					failed_worker_reporter <- true
					return
				}

				type_store := BuildTypeStore()
				client := tlj.NewClient(conn, type_store, true)
				err = client.Message(WorkerAuth{
					Nonce: nonce,
				})
				if err != nil {
					failed_worker_reporter <- true
					return
				}

				writer, _ := tlj.NewStreamWriter(
					conn,
					type_store,
					reflect.TypeOf(TransferChunk{}),
				)
				streamer_chan <- writer
			}()
		}
	}

	halt_prints := make(chan bool, 1)
	go func() {
		for {
			select {
			case stream_writer := <-streamer_chan:
				success_worker_count += 1
				streamers = append(streamers, stream_writer)
				worker_server.Insert(stream_writer.Socket)
			case <-failed_worker_reporter:
				failed_worker_count += 1
			case <-halt_prints:
				return
			}
			worker_status_update_text <- fmt.Sprintf(
				"built %d/%d transfer sockets, %d failed",
				success_worker_count,
				total_worker_count,
				failed_worker_count,
			)
		}
	}()
	worker_waiter.Wait()
	halt_prints <- true

	worker_build_finished_text <- fmt.Sprintf(
		"%d/%d transfer sockets built, %d failed in %s",
		success_worker_count,
		total_worker_count,
		failed_worker_count,
		time.Since(start).String(),
	)
	if total_worker_count == failed_worker_count {
		err = errors.New("all transfer sockets failed to build")
		return
	}
	return
}

func timeRemaining(speed, remaining int) string {
	seconds_left := float64(remaining) / float64(speed)
	str, _ := time.ParseDuration(fmt.Sprintf("%fs", seconds_left))
	return str.String()
}

func uploadFiles(file_list []string, total_bytes int, streamers []tlj.StreamWriter) {
	chunks, file_finished := CreatePooledChunkChan(file_list, chunk_size)
	file_finished_print := make(chan string)
	status_update := make(chan string)
	all_done := make(chan string)
	worker_speeds := make(map[int]int)
	moved_bytes := 0
	total_update := make(chan int)
	finished := false
	start := time.Now()
	for iter, worker := range streamers {
		worker_speeds[iter] = 0
		speed_update := make(chan int)
		go StreamChunksToPut(worker, chunks, speed_update, total_update)
		go func(liter int) {
			for speed := range speed_update {
				worker_speeds[liter] = speed
			}
		}(iter)
		go func() {
			for moved := range total_update {
				moved_bytes += moved
			}
		}()
	}
	go func() {
		for _, _ = range file_list {
			file_finished_print <- <-file_finished
		}
		all_done <- fmt.Sprintf(
			"%d file%s (%s) transferred in %s",
			len(file_list),
			(map[bool]string{true: "s", false: ""})[len(file_list) > 1], // deal with it ಠ_ಠ
			humanize.Bytes(uint64(total_bytes)),
			time.Since(start).String(),
		)
		finished = true
	}()
	go func() {
		for {
			if finished {
				return
			}
			pool_speed := 0
			for _, speed := range worker_speeds {
				pool_speed += speed
			}
			byte_progress := moved_bytes / total_bytes
			status_update <- fmt.Sprintf(
				"transferring %d files (%s) at %s/s %s%% complete %s remaining",
				len(file_list),
				humanize.Bytes(uint64(total_bytes)),
				humanize.Bytes(uint64(pool_speed)),
				humanize.Ftoa(float64(int(byte_progress*10000))/100),
				timeRemaining(pool_speed, total_bytes-moved_bytes),
			)
			time.Sleep(1 * time.Second)
		}
	}()
	printProgress(file_finished_print, status_update, all_done)
}

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
			// send a Command{} with get and the files as args (server wont respond (or does it need to respond when all done and with updates?), will just stream chunks down nonced workers)
			// file finished are send by write buffer to global current_transfer chan
			// speed updates are 1 per second?  need to ask every worker?  (workers update global speed store, sum that)
			//PrintProgress(file_finished, speed_update, all_done)
		} else if command == "put" {
			file_list, total_bytes := ParseFileList(args)
			uploadFiles(file_list, total_bytes, []tlj.StreamWriter{})
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
	resume = *flag.Bool("resume", true, "resume transfers if a part of the file already exists on the destination")
	chunk_size = *flag.Int("chunksize", 5*1024*1024, "size of each file chink in byte")
	flag.Parse()

	networks, err := parseNetworks(network_config)
	if err != nil {
		log.Fatal(err)
	}
	client, err := createClient(hostname, port)
	if err != nil {
		log.Fatal(err)
	}
	nonce := clientLogin(username, client)
	if err != nil {
		log.Fatal(err)
	}

	worker_server := tlj.Server{
		TypeStore:       type_store,
		Tag:             tag,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, tlj.TLJContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	go worker_server.process()
	worker_server.Accept("peer", reflect.TypeOf(TransferChunk{}), func(_ tlj.TLJContext, iface interface{}) {
		if chunk, ok := iface.(*Chunk); ok {
			sentToChunkWriter(chunk)
		}
	})
	streamers, err := ConnectWorkers(hostname, port, worker_server, networks, nonce, resume)
	if err != nil {
		log.Fatal(err)
	}

	go CommandLoop(client, streamers, chunk_size)
	err = <-client.Dead
	fmt.Println("control connection closed:", err)
}
