package main

func main() {
	for {
		// parse server struct from conf file?
		client := server.RecieveConnection()
		session = Session{client}
		go session.Start()
	}
}
