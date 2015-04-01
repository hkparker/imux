package main

import "fmt"

// gtk interface

func main() {
	host, _ := CreateHost("127.0.0.1", 8080, "hayden", "", nil, nil)
	fmt.Println(host.WorkingDirectory())
	fmt.Println(host.ChangeDirectory("Downloads"))
	fmt.Println(host.CreateDirectory("/home/hayden/testdir"))
	fmt.Println(host.Remove("/home/hayden/testdir"))
	fmt.Println(host.List("."))
	host.Close()
}
