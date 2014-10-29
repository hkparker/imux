package main

import (
	"fmt"
	"os"
	"time"
	"strconv"
)

type Buffer struct {
	Chunks chan Chunk
	Queue []Chunk
	LastDump int
}

func (buffer *Buffer) Insert(chunk Chunk) {
	if len(buffer.Queue) == 0 {
		buffer.Queue = append(buffer.Queue, chunk)
	} else {
		for iter, nextChunk := range buffer.Queue {
			if iter == len(buffer.Queue)-1 {
				buffer.Queue = append(buffer.Queue, chunk)
				break
			}
			if chunk.ID < nextChunk.ID {
				// if the slice is only size one, insert it in the beginning
				end := buffer.Queue[iter-1:]
				buffer.Queue = append(buffer.Queue[:iter-1], chunk)
				buffer.Queue = append(buffer.Queue, end...)
				break
			}
		}
	}	
}

func (buffer *Buffer) Dump(file *os.File) {
	for iter, chunk := range buffer.Queue {
		if chunk.ID == buffer.LastDump+1 {
			if len(buffer.Queue) == 1 {
				buffer.Queue = buffer.Queue[1:]
			} else {
				buffer.Queue = append(buffer.Queue[:iter-1], buffer.Queue[iter:]...)
			}
			file.WriteString(chunk.Data)
			buffer.LastDump = buffer.LastDump+1
		} else {
			break
		}
	}
}

func (buffer Buffer) Process(file *os.File) {
	fmt.Println("Started processing...")
	for chunk := range buffer.Chunks {
		fmt.Println("Receieved new chunk, id", chunk.ID)
		buffer.Insert(chunk)
		fmt.Println("Inserted, queue is now", buffer.Queue)
		buffer.Dump(file)
		fmt.Println("Dumped to file, queue is now", buffer.Queue)
	}
}

func InsertChunk(number int, buffer Buffer) {
	chunk := Chunk{}
	chunk.Data = strconv.Itoa(number)
	chunk.ID = number
	buffer.Chunks <- chunk
}

func main() {
	// IMUXManager is going to call these when doing a transfer
	buffer := Buffer{}
	buffer.Chunks = make(chan Chunk, 10)
	
	file, _ := os.Create("testfile")
	defer file.Close()
	go buffer.Process(file)
	InsertChunk(1, buffer)
	InsertChunk(3, buffer)
	InsertChunk(4, buffer)
	InsertChunk(2, buffer)
	InsertChunk(5, buffer)
	close(buffer.Chunks)
	time.Sleep(1000 * time.Millisecond)
}
