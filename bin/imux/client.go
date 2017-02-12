package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hkparker/imux"
	"net"
	"os"
	"reflect"
	"strings"
)

// Dial the TLS server specified in the dial address and
// perform Trust Of First Use, interactively checking if
// the presented certificate is safe and if it should be
// saved in ~/.imux/known_hosts or if the connection
// should be aborted.
func TOFU(dial string) *x509.Certificate {
	known_hosts := LoadKnownHosts()
	conn, err := tls.Dial(
		"tcp",
		dial,
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "TOFU",
			"error": err.Error(),
		}).Fatal("unable to dial server")
	}
	signature := SHA256Sig(conn)

	if saved_signature, present := known_hosts[dial]; present {
		if signature != saved_signature {
			connect, update := MitMWarning(signature, saved_signature)
			if !connect {
				log.WithFields(log.Fields{
					"at": "TOFU",
				}).Fatal("TLS certificate mismatch")
			}
			if update {
				AppendHost(dial, signature)
			}
		}
	} else {
		connect, save_cert := TrustDialog(dial, signature)
		if !connect {
			log.WithFields(log.Fields{
				"at": "TOFU",
			}).Fatal("TLS certificate rejected by user")
		} else if save_cert {
			AppendHost(dial, signature)
		}
	}

	return conn.ConnectionState().PeerCertificates[0]
}

// Parse the listen address and return a TCP listsner
func createClientListener(listen string) net.Listener {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.WithFields(log.Fields{
			"at":      "createClientListener",
			"address": listen,
			"error":   err.Error(),
		}).Fatal("unable to open client listener")
	}
	return listener
}

// Create a function that accepts bind address and returns imux.Redialer
// functions that bind to that address and dial the specified dial address
func createRedailerGenerator(dial string, cert *x509.Certificate) imux.RedialerGenerator {
	return func(bind string) imux.Redialer {
		return func() (net.Conn, error) {
			bind_addr, err := net.ResolveTCPAddr("tcp", bind+":0")
			if err != nil {
				log.WithFields(log.Fields{
					"at":      "createRedialerGenerator",
					"address": bind,
					"error":   err.Error(),
				}).Error("error parsing bind address")
			}
			conn, err := tls.DialWithDialer(
				&net.Dialer{
					LocalAddr: bind_addr,
				},
				"tcp",
				dial,
				&tls.Config{InsecureSkipVerify: true},
			)
			if err != nil {
			}
			if !reflect.DeepEqual(cert.Signature, conn.ConnectionState().PeerCertificates[0].Signature) {
				log.Error("holy shit")
			}
			return conn, err
		}
	}
}

func LoadKnownHosts() map[string]string {
	sigs := make(map[string]string)
	filename := os.Getenv("HOME") + "/.imux/known_hosts"
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

func MitMWarning(new_signature, old_signature string) (bool, bool) {
	fmt.Println(fmt.Sprintf(
		"WARNING: Remote certificate has changed!!\nold: %s\nnew: %s",
		old_signature,
		new_signature,
	))
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

func AppendHost(hostname string, signature string) {
	filename := os.Getenv("HOME") + "/.imux/known_hosts"
	known_hosts, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	known_hosts.WriteString(hostname + " " + signature + "\n")
	known_hosts.Close()
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

func SHA256Sig(conn *tls.Conn) string {
	sig := conn.ConnectionState().PeerCertificates[0].Signature
	sha := sha256.Sum256(sig)
	str := hex.EncodeToString(sha[:])
	return str
}
