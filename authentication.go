package imux

import (
	"os"
)

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

func AppendKnownHost(hostname string, signature string) {
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

func ReadPassword() string {
	fmt.Print("Password: ")
	password_bytes, _ := terminal.ReadPassword(0)
	fmt.Println()
	return strings.TrimSpace(string(password_bytes))
}

func ClientLogin(username string, client tlj.Client) string {
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

func MitmWarning(new_signature, old_signature string) (bool, bool) {
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
