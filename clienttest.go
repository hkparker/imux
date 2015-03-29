package main

import (
	"crypto/tls"
	"io"
	"log"
	"os"
	"bytes"
	"crypto/sha256"
	"strconv"
	"bufio"
	"strings"
)

func load_hosts() map[string]string {
	sigs := make(map[string]string)
	filename := os.Getenv("HOME")+"/.multiplexity/known_hosts"
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

func append_host(hostname string, signature string) {
	filename := os.Getenv("HOME")+"/.multiplexity/known_hosts"
	known_hosts, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	known_hosts.WriteString(hostname+" "+signature+"\n") 
	known_hosts.Close()
}

func SHA256Sig(conn *tls.Conn) string {
	signature := sha256.Sum256(conn.ConnectionState().PeerCertificates[0].Signature)
	var characters bytes.Buffer
	for i, chr := range(signature) {
		characters.WriteString(strconv.Itoa(int(chr)))
		if i != 31 {
			characters.WriteString(":")
		}
	}
	return characters.String()
}

func main() {
	hostname := "127.0.0.1"
	hosts := load_hosts()
	
	conn, err := tls.Dial("tcp", hostname+":8080", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	signature := SHA256Sig(conn)

	if saved_signature, present := hosts[hostname]; present {
		if signature == saved_signature {
			// we've seen this host with this sig, all good (create a host object)
		} else {
			// WARNING!  We have seen this host with a different sig.  Contiue (y/y+update/no)
			update := false
			if update {
				append_host(hostname, signature)
			} else {
				// abort connection
				return
			}
		}
	} else {
		// we have never seen this host before.  here is the sig.  trust?  (once/forever/no)
		connect := true
		save_cert := true
		if connect {
			if save_cert {
				append_host(hostname, signature)
			}
			// continue to interact with the host	(create a host object)
		}
	}
	
	// send authentication
	message := ""
	_, err = io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("server: conn: read: %s", err)
	}
	log.Printf(string(buf[:n]))
	
	n, err = conn.Read(buf)
	if err != nil {
		log.Printf("server: conn: read: %s", err)
	}
	log.Printf(string(buf[:n]))
	
	message = "cd ."
	_, err = io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	buf = make([]byte, 512)
	n, err = conn.Read(buf)
	if err != nil {
		log.Printf("server: conn: read: %s", err)
	}
	log.Printf(string(buf[:n]))
	
	message = "pwd"
	_, err = io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	buf = make([]byte, 512)
	n, err = conn.Read(buf)
	if err != nil {
		log.Printf("server: conn: read: %s", err)
	}
	log.Printf(string(buf[:n]))
	
	
	buf = make([]byte, 512)
	n, err = conn.Read(buf)
	if err != nil {
		log.Printf("server: conn: read: %s", err)
	}
	log.Printf(string(buf[:n]))
	conn.Close()
}



