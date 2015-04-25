package main

import (
	"log"
	"net"
	"crypto/x509"
	"crypto/rand"
	"crypto/tls"
	"io/ioutil"
	"github.com/twinj/uuid"
	)

type TransferGroup struct {
	PeerAddr string
	Server net.Listener
	Errors chan string
	//Sockets map[string]TransferSocket	// UUID -> socket
	Dialers map[string]net.Dialer
	UUID string
	Signature string
}

func (transfer_group *TransferGroup) Open(bind_ips map[string]int, uuid string) {
	// Make sure a tls.Dialer exists in Dialers for each bind ip address
	for bind_ip, _ := range bind_ips {
		if _ , present := transfer_group.Dialers[bind_ip]; !present {
			transfer_group.Dialers[bind_ip] = &net.Dialer{
                LocalAddr : Addr{
					Network: "tcp",
					String: bind_ip
				},
			}
		}
	}
	
	for bind_ip, count := range bind_ips {
		dialer = transfer_group.Dialers[bind_ip]
		connect := func(transfer_socket *TransferSocket) error {
			conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":8080", &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				return err
			}
			if transfer_group.Signature != SHA256Sig(conn) {		// from package multiplexity?
				return errors.New("TLS signature mismatch on TransferSocket")
			}
			// make sure the uuid of the group is correct
		}
		transfer_socket := TransferSocket {
			Group: transfer_group,
			GenerateSockets: connect,
			Signature: transfer_group.Signature,
		}
		transfer_group.Sockets = append(transfer_group.Sockets, transfer_socket)
	}
	
	async_connect := func(transfer_socket *TransferSocket, done chan string) {
		err = transfer_socket.GenerateSocket(&transfer_socket)
		if err == nil {
			done <- "0"
		} else {
			done <- err.Error()
		}
	}
	
	done = make(chan string)
	for _, transfer_socket := range transfer_group.Sockets {
		go async_connect(&transfer_socket, done)
	}
	
	for _, _ := range transfer_group.Sockets {
		s := <- done
		if s != "0" {
			transfer_group.Errors <- s
		}
	}
}

func (transfer_group *TransferGroup) Recieve(count int, uuid string) {
	if transfer_group.Server == nil {
		// setup TLS server
	}
	
	connect := func(transfer_socket *TransferSocket) error {
		// recieve a socket from transfer_group.Server
		// verify UUID
	}
	
	async_connect := func(transfer_socket *TransferSocket, done chan string) {
		err = transfer_socket.GenerateSocket(&transfer_socket)
		if err == nil {
			done <- "0"
		} else {
			done <- err.Error()
		}
	}
	
	done = make(chan string)
	for i := 0; i < count; i++ {
		transfer_socket = TransferSocket{
			Group: transfer_group,
			GenerateSockets: connect,
		}
		go async_connect(&transfer_socket, done)
	}
	
	for i := 0; i < count; i++ {
		s := <- done
		if s != "0" {
			transfer_group.Errors <- s
		}	
	}
}

func (transfer_group *TransferGroup) ServeFile(file *os.File, starting_position int, chuink_size int) {
	file_queue := ReadQueue{
		FileHead: starting_position,
		ChunkSize: chuink_size,
	}
	file_queue.Process(file)
	done = make(chan int)
	pending := 0
	for _, transfer_socket := range transfer_group.Sockets {
		pending += 1
		transfer_socket.Serve(file_queue, done)
	}
	for i := 0; i < pending; i++ {
		s := <- done
		if s != "0" {
			transfer_group.Errors <- s
		}
	}
}

func (transfer_group *TransferGroup) DownloadFile(file *os.File) {
	file_buffer = WriteBuffer{}
	file_buffer.Process(file)
	done = make(chan int)
	pending := 0
	for _, transfer_socket := range transfer_group.Sockets {
		transfer_socket.Download(buffer, done)
	}
	for i := 0; i < pending; i++ {
		s := <- done
		if s != "0" {
			transfer_group.Errors <- s
		}
	}
}

func (transfer_group *TransferGroup) UpdateRecycling(state bool) {
	for _, transfer_socket := range transfer_group.Sockets {
		transfer_socket.Recycle = state
	}
}

func (transfer_group *TransferGroup) CurrentSpeed() float64 {
	var speed float64 = 0.0
	for _, transfer_socket := range transfer_group.Sockets {
		speed += transfer_socket.LastSpeed
	}
	return speed
}

func (transfer_group *TransferGroup) Close() {
	for _, transfer_socket := range transfer_group.Sockets {
		err := transfer_socket.Close()
		if err != nil {
			transfer_group.Errors <- err.Error()
		}
	}
	// remove from Session
}


func main() {
	file, _ := os.Open("/hayden/Pictures/render.png")
	defer file.Close()
	dst_file, _ := os.Create("/hayden/Pictures/render2.png")
	defer dst_file.Close()
	
	
}
