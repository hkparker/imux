package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hkparker/TLJ"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

func StreamChunksDownWorker(worker tlj.StreamWriter, chunks, stale_chunks chan TransferChunk) {
	for chunk := range chunks {
		err := worker.Write(chunk)
		if err != nil {
			stale_chunks <- chunk
		}
	}
}

func CreatePooledChunkChan(files []string, chunk_size, total_bytes int) (chan TransferChunk, chan string, chan string, chan string, chan TransferChunk) {
	all_chunks := make(chan TransferChunk)
	file_done := make(chan string)
	progress_update := make(chan string)
	transfer_finished := make(chan string)
	stale_chunks := make(chan TransferChunk)
	all_files := &sync.WaitGroup{}
	all_files.Add(len(files))

	pool_speed := 0
	moved_bytes := 0
	moved_bytes_update := &sync.Mutex{}

	start := time.Now()
	for _, file := range files {
		queue := ReadQueue{
			ChunkSize:   chunk_size,
			Chunks:      make(chan TransferChunk, 1),
			Destination: file,
		}
		fh, ferr := os.Open(file)
		if ferr == nil {
			go func(file_handler *os.File, filename string, read_queue ReadQueue) {
				read_queue.Process(file_handler)
				all_files.Done()
				file_done <- filename
			}(fh, file, queue)
			go func(filename string, read_queue ReadQueue) {
				for chunk := range read_queue.Chunks {
					//chunk_start := time.Now()
					all_chunks <- chunk
					moved_bytes_update.Lock()
					moved_bytes += len(chunk.Data)
					moved_bytes_update.Unlock()
					// update my part of the pool speed
				}
			}(file, queue)
		} else {
			fmt.Println(fmt.Sprintf("skipping %s: %v", file, ferr))
		}
	}

	go func() {
		//stale := <-stale_chunks
		//lookup the read queue and write it back if it exists
	}()

	go func() {
		all_files.Wait()
		transfer_finished <- fmt.Sprintf(
			"%d file%s (%s) transferred in %s",
			len(files),
			(map[bool]string{true: "s", false: ""})[len(files) > 1],
			humanize.Bytes(uint64(total_bytes)),
			time.Since(start).String(),
		)
	}()

	go func() {
		for {
			progress_update <- fmt.Sprintf(
				"transferring %d files (%s) at %s/s %s%% complete %s remaining",
				len(files),
				humanize.Bytes(uint64(total_bytes)),
				humanize.Bytes(uint64(pool_speed)),
				humanize.Ftoa(float64(int((moved_bytes/total_bytes)*10000))/100),
				timeRemaining(pool_speed, total_bytes-moved_bytes),
			)
			time.Sleep(1 * time.Second)
		}
	}()

	return all_chunks, file_done, progress_update, transfer_finished, stale_chunks
}

func UploadFiles(file_list []string, total_bytes int, streamers []tlj.StreamWriter, chunk_size int) {
	all_chunks, file_finished, progress_update, transfer_finished, stale_chunks := CreatePooledChunkChan(file_list, chunk_size, total_bytes)
	for _, worker := range streamers {
		go StreamChunksDownWorker(worker, all_chunks, stale_chunks)
	}
	PrintProgress(file_finished, progress_update, transfer_finished)
}

func timeRemaining(speed, remaining int) string {
	seconds_left := float64(remaining) / float64(speed)
	str, _ := time.ParseDuration(fmt.Sprintf("%fs", seconds_left))
	return str.String()
}

func ConnectWorkers(hostname string, port int, networks map[string]int, nonce string) (streamers []tlj.StreamWriter, err error) {
	discard_listener, _ := net.Listen("tcp", "127.0.0.1:0")
	worker_server := tlj.NewServer(
		discard_listener,
		TagSocketAll,
		type_store,
	)
	worker_server.Accept("peer", reflect.TypeOf(TransferChunk{}), func(iface interface{}, _ tlj.TLJContext) {
		if _, ok := iface.(*TransferChunk); ok {
			//SentToChunkWriter(chunk)
		}
	})

	worker_status_update_text := make(chan string)
	worker_build_finished_text := make(chan string)

	streamer_chan := make(chan tlj.StreamWriter)
	success_worker_count := 0
	failed_worker_count := 0
	total_worker_count := 0
	var worker_waiter sync.WaitGroup
	for _, count := range networks {
		worker_waiter.Add(count)
		total_worker_count += count
	}

	failed_worker_reporter := make(chan bool, total_worker_count)
	start := time.Now()
	for _, count := range networks {
		for i := 0; i < count; i++ {
			go func() {
				//dialer := net.Dialer{
				//	LocalAddr: net.ParseIP(network),
				//}
				conn, err := tls.Dial( //WithDialer(
					//&dialer,
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{
						InsecureSkipVerify: true,
					},
				)
				signature := SHA256Sig(conn)
				known_hosts := LoadKnownHosts()
				if saved_signature, present := known_hosts[conn.RemoteAddr().String()]; present {
					if signature != saved_signature {
						err = errors.New("TLS signature validation for worker failed")
					}
				}
				if err != nil {
					failed_worker_reporter <- true
					return
				}

				type_store := BuildTypeStore()
				client := tlj.NewClient(conn, type_store, true)
				err = client.Message(WorkerAuth{
					Nonce: nonce,
				})
				if err != nil {
					failed_worker_reporter <- true
					return
				}

				writer, _ := tlj.NewStreamWriter(
					conn,
					type_store,
					reflect.TypeOf(TransferChunk{}),
				)
				streamer_chan <- writer
			}()
		}
	}

	halt_prints := make(chan bool, 1)
	go func() {
		for {
			select {
			case stream_writer := <-streamer_chan:
				success_worker_count += 1
				streamers = append(streamers, stream_writer)
				worker_server.Insert(stream_writer.Socket)
				worker_server.TagSocket(stream_writer.Socket, "peer")
			case <-failed_worker_reporter:
				failed_worker_count += 1
			case <-halt_prints:
				return
			}
			worker_status_update_text <- fmt.Sprintf(
				"built %d/%d transfer sockets, %d failed",
				success_worker_count,
				total_worker_count,
				failed_worker_count,
			)
			worker_waiter.Done()
		}
	}()

	go func() {
		worker_waiter.Wait()
		halt_prints <- true

		worker_build_finished_text <- fmt.Sprintf(
			"%d/%d transfer sockets built, %d failed in %s",
			success_worker_count,
			total_worker_count,
			failed_worker_count,
			time.Since(start).String(),
		)
	}()

	PrintProgress(
		make(chan string),
		worker_status_update_text,
		worker_build_finished_text,
	)

	if total_worker_count == failed_worker_count {
		err = errors.New("all transfer sockets failed to build")
	}
	return
}

func CreateClient(hostname string, port int) (tlj.Client, error) {
	known_hosts := LoadKnownHosts()
	conn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", hostname, port),
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return tlj.Client{}, err
	}
	signature := SHA256Sig(conn)
	if saved_signature, present := known_hosts[conn.RemoteAddr().String()]; present {
		if signature != saved_signature {
			connect, update := MitMWarning(signature, saved_signature)
			if !connect {
				return tlj.Client{}, errors.New("TLS certificate mismatch")
			}
			if update {
				AppendKnownHost(conn.RemoteAddr().String(), signature)
			}
		}
	} else {
		connect, save_cert := TrustDialog(hostname, signature)
		if !connect {
			return tlj.Client{}, errors.New("TLS certificate rejected")
		}
		if save_cert {
			AppendKnownHost(conn.RemoteAddr().String(), signature)
		}
	}

	type_store := BuildTypeStore()
	client := tlj.NewClient(conn, type_store, false)
	return client, nil
}
