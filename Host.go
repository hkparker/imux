package main

import (
	"crypto/tls"
	"log"
	"os"
	"bytes"
	"crypto/sha256"
	"strconv"
	"bufio"
	"strings"
	"fmt"
	"errors"
	"encoding/json"
)

type Host struct {
	IP string
	Port int
	Session *tls.Conn
}

func LoadKnownHosts() map[string]string {
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

func AppendHost(hostname string, signature string) {
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

func CreateHost(ip string, port int, username, password string, trust_dialog , mitm_warning func()(bool, bool)) (Host, error) {
	host := Host{
		IP: ip,
		Port: port,
	}
	
	known_hosts := LoadKnownHosts()
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalf("error connecting to host: %s", err)
	}
	host.Session = conn
	signature := SHA256Sig(conn)
	
	if saved_signature, present := known_hosts[host.IP]; present {
		if signature != saved_signature {
			connect, update := mitm_warning()
			if !connect {
				return host, errors.New("TLS certificate mismatch")
			}
			if update {
				AppendHost(ip, signature)
			}
		}
	} else {
		connect, save_cert := trust_dialog()
		if connect {
			if save_cert {
				AppendHost(ip, signature)
			}
		}
	}
	
	login := host.QuerySession(fmt.Sprintf("%s %s", username, password))
	if err != nil {
		log.Printf("error authenticating: %s", err)
	}
	if login != "Authentication successful" {
		return host, errors.New("Authentication failed")
	}
	
	return host, nil
}

func (host *Host) QuerySession(query string) string {
	_, err := host.Session.Write([]byte(query))
	if err != nil {
		log.Println(err)
	}
	resp, err := ReadLine(host.Session)
	if err != nil {
		log.Println(err)
	}
	return resp
}

func (host *Host) Close() {
	host.QuerySession("close")
}

func (host *Host) WorkingDirectory() string {
	return host.QuerySession("pwd")
}

func (host *Host) ChangeDirectory(dir string) string {
	return host.QuerySession("cd " + dir)
}

func (host *Host) CreateDirectory(dir string) string {
	return host.QuerySession("mkdir " + dir)
}

func (host *Host) List(dir string) []Entry {
	resp := host.QuerySession("ls " + dir)
	files := make([]Entry, 0)
	err := json.Unmarshal([]byte(resp), &files)
	if err != nil {
		log.Println(err)
	}
	return files
}

func (host *Host) Remove(item string) string {
	return host.QuerySession("rm " + item)
}
