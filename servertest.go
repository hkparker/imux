package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
)

func main() {
	ca_b, _ := ioutil.ReadFile("ca.pem")
	ca, _ := x509.ParseCertificate(ca_b)
	priv_b, _ := ioutil.ReadFile("ca.key")
	priv, _ := x509.ParsePKCS1PrivateKey(priv_b)
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	cert := tls.Certificate{
		Certificate: [][]byte{ ca_b },
		PrivateKey: priv,
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs: pool,
	}
	config.CipherSuites = []uint16{
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
        tls.TLS_RSA_WITH_AES_256_CBC_SHA,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256}
    config.MinVersion = tls.VersionTLS12
	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", "0.0.0.0:8080", &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("server: conn: read: %s", err)
			break
		}
		log.Printf("server: conn: echo %q\n", string(buf[:n]))
		
		n, err = conn.Write(buf[:n])
		log.Printf("server: conn: wrote %d bytes", n)
		if err != nil {
			log.Printf("server: write: %s", err)
			break
		}
	}
	log.Println("server: conn: closed")
}
