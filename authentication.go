package imux

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"github.com/hkparker/TLJ"
	b58 "github.com/jbenet/go-base58"
	"github.com/kless/osutil/user/crypt/sha512_crypt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"reflect"
	"strings"
)

var conf_dir_path = "/.imux/"
var known_hosts_path = conf_dir_path + "known_hosts"

func LoadKnownHosts() map[string]string {
	sigs := make(map[string]string)
	conf_dir := os.Getenv("HOME") + conf_dir_path
	if _, err := os.Stat(conf_dir); os.IsNotExist(err) {
		os.Mkdir(conf_dir, 0700)
	}
	filename := os.Getenv("HOME") + known_hosts_path
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

func AppendKnownHost(hostname string, signature string) {
	filename := os.Getenv("HOME") + known_hosts_path
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

func ReadPassword() string {
	fmt.Print("Password: ")
	password_bytes, _ := terminal.ReadPassword(0)
	fmt.Println()
	return strings.TrimSpace(string(password_bytes))
}

func ClientLogin(username string, client tlj.Client) string {
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

func NewNonce() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return b58.Encode(bytes), nil
}

func PrepareTLSConfig(pem, key string) tls.Config {
	ca_b, err := ioutil.ReadFile(pem)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ca, err := x509.ParseCertificate(ca_b)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	priv_b, err := ioutil.ReadFile(key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	priv, err := x509.ParsePKCS1PrivateKey(priv_b)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}
	config.MinVersion = tls.VersionTLS12
	config.Rand = rand.Reader
	return config
}

func LookupHashAndHeader(username string) (string, string) {
	file, err := os.Open("/etc/shadow")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split_colon := func(c rune) bool {
			return c == 58
		}
		split_dollar := func(c rune) bool {
			return c == 36
		}
		fields := strings.FieldsFunc(line, split_colon)
		if fields[0] == username {
			pw_fields := strings.FieldsFunc(fields[1], split_dollar)
			header := "$" + pw_fields[0] + "$" + pw_fields[1]
			return fields[1], header
		}
	}
	return "", ""
}

func Login(username, password string) bool {
	_, err := user.Lookup(username)
	if err != nil {
		return false
	}
	passwd_crypt := sha512_crypt.New()
	hash, header := LookupHashAndHeader(username)
	new_hash, err := passwd_crypt.Generate([]byte(password), []byte(header))
	if err != nil {
		return false
	}
	if new_hash != hash {
		return false
	}
	return true
}
