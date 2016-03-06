package imux

import (
	"fmt"
	"github.com/hkparker/TLJ"
	"net"
)

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func ParseFileList(items []string) ([]string, int) {
	// for each string which should be a full path
	// if it is a directory
	all_files := items //make([]string, 0)
	//visitor := func() {
	// if the item is a readable file, add it to the list
	//}
	//walk
	return all_files, 100 // also return the size in bytes of all files
}

func PrintProgress(completed_files, statuses, finished chan string) {
	last_status := ""
	last_len := 0
	for {
		select {
		case completed_file := <-completed_files:
			fmt.Printf("\r")
			line := "completed: " + completed_file
			fmt.Print(line)
			print_len := len(line)
			trail_len := last_len - print_len
			if trail_len > 0 {
				for i := 0; i < trail_len; i++ {
					fmt.Print(" ")
				}
			}
			fmt.Print("\n" + last_status)
		case status := <-statuses:
			last_status = status
			fmt.Printf("\r")
			fmt.Print(status)
			print_len := len(status)
			trail_len := last_len - print_len
			if trail_len > 0 {
				for i := 0; i < trail_len; i++ {
					fmt.Print(" ")
				}
			}
			last_len = print_len
		case elapsed := <-finished:
			fmt.Println("\n" + elapsed)
			return
		}
	}
}

func ParseNetworks(data string) (map[string]int, error) {
	networks := make(map[string]int)
	networks["0.0.0.0"] = 20
	return networks, nil
}

func UsernameFromTags(tags []string) string {
	for _, tag := range tags {
		if len(tag) > 5 {
			if tag[:5] == "user:" {
				return tag[5:]
			}
		}
	}
	return ""
}
