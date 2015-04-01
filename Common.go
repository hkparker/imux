package main

import "net"

type Entry struct {
	Name string
	Size int64
	Perms string
	Mod string
}

func ReadBytes(conn net.Conn) ([]byte, error) {
	resp := make([]byte, 0)
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		buf = buf[:n]
		if err != nil {
			return resp, err
		} else {
			resp = append(resp, buf...)
			if n < 1024 {
				return resp, nil
			}
		}
	}
	return resp, nil
}

func ReadLine(conn net.Conn) (string, error) {
	bytes, err := ReadBytes(conn)
	return string(bytes), err
}

