package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"encoding/base64"
	"crypto/sha256"
)

func check_signature() {
	// returns either sig good, sig bad, or sig never seen
}

func main() {
	conn, err := tls.Dial("tcp", "127.0.0.1:8080", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	cert := state.PeerCertificates[0]
	fmt.Println(x509.MarshalPKIXPublicKey(cert.PublicKey))
	fmt.Println(cert.Subject)
	fmt.Println(base64.StdEncoding.EncodeToString(cert.Signature))
	fmt.Println(sha256.Sum256(cert.Signature))
	
	log.Println("client: handshake: ", state.HandshakeComplete)
	log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)
	
	

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	log.Print("client: exiting")
}


// if never seen cert before
	// choose to connect, and if so if the key will be stored
	// if deciding to connect
		// if adding the key
			// add key to knownw hosts
		// set cert as host cert
	// if not deciding to connect
		// disconnect from host


