package common

import (
	"fmt"
	"github.com/hkparker/TLJ"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func TagSocketAll(socket net.Conn, server *tlj.Server) {
	server.TagSocket(socket, "all")
}

func ParseFileList(items []string) ([]string, int) {
	final_list := make([]string, 0)
	total_size := 0

	visitor := func(path string, file os.FileInfo, err error) error {
		fmt.Printf("%s with %d bytes\n", path, file.Size())
		if file.Mode().IsRegular() {
			final_list = append(final_list, path)
			total_size += int(file.Size())
		}
		return nil
	}

	for _, item := range items {
		fh, err := os.Open(item)
		defer fh.Close()
		if err == nil {
			fi, err := fh.Stat()
			if err == nil {
				if fi.Mode().IsDir() {
					filepath.Walk(item, visitor)
				} else if fi.Mode().IsRegular() {
					final_list = append(final_list, item)
					total_size += int(fi.Size())
				}
			}
		}
	}

	return final_list, total_size
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

func ParseNetworks(networks string) map[string]int {
	nets := make(map[string]int)
	pairs := strings.Split(networks, ";")
	for _, pair := range pairs {
		config := strings.Split(pair, ":")
		count, _ := strconv.Atoi(config[1])
		nets[config[0]] = count
	}
	return nets
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
