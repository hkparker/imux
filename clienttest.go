package main

import (
	"crypto/tls"
	"fmt"
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

func sha256_sig_to_string(signature [32]byte) string {
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
	signature := sha256_sig_to_string(sha256.Sum256(conn.ConnectionState().PeerCertificates[0].Signature))

	if saved_signature, present := hosts[hostname]; present {
		if signature == saved_signature {
			fmt.Println("Seen it!")
			// we've seen this host with this sign, all good (create a host object)
		} else {
			fmt.Println("MITM!")
			override := false
			if override {
				append_host(hostname, signature)
			} else {
				// abort connection
				return
			}
		}
	} else {
		//prompt the user to continue, and if so save cert?
		fmt.Println("This ok?")
		connect := true
		save_cert := true
		if connect {
			if save_cert {
				append_host(hostname, signature)
			}
			fmt.Println("Looks like it")
			// continue to interact with the host	(create a host object)
		} else {
			// abort attempt to connect to host
			return
		}
	}
	

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	conn.Close()
}



