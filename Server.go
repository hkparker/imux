//package multiplexity
package main

func main() {
	for {
		// parse server struxt from conf file?
		client := server.RecieveConnection()
		session = Session{client}
		go session.Start()
	}
}
