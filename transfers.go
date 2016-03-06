package imux

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hkparker/TLJ"
	"net"
	"reflect"
	"sync"
	"time"
)

func MoveFiles(file_list []string, total_bytes int, streamers []tlj.StreamWriter) {
	_, file_finished := CreatePooledChunkChan(file_list, 5*1024*1024)
	file_finished_print := make(chan string)
	status_update := make(chan string)
	all_done := make(chan string)
	worker_speeds := make(map[int]int)
	moved_bytes := 0
	total_update := make(chan int)
	finished := false
	start := time.Now()
	for iter, _ := range streamers {
		worker_speeds[iter] = 0
		speed_update := make(chan int)
		//go StreamChunksToPut(worker, chunks, speed_update, total_update)
		go func(liter int) {
			for speed := range speed_update {
				worker_speeds[liter] = speed
			}
		}(iter)
		go func() {
			for moved := range total_update {
				moved_bytes += moved
			}
		}()
	}
	go func() {
		for _, _ = range file_list {
			file_finished_print <- <-file_finished
		}
		all_done <- fmt.Sprintf(
			"%d file%s (%s) transferred in %s",
			len(file_list),
			(map[bool]string{true: "s", false: ""})[len(file_list) > 1], // deal with it ಠ_ಠ
			humanize.Bytes(uint64(total_bytes)),
			time.Since(start).String(),
		)
		finished = true
	}()
	go func() {
		for {
			if finished {
				return
			}
			pool_speed := 0
			for _, speed := range worker_speeds {
				pool_speed += speed
			}
			byte_progress := moved_bytes / total_bytes
			status_update <- fmt.Sprintf(
				"transferring %d files (%s) at %s/s %s%% complete %s remaining",
				len(file_list),
				humanize.Bytes(uint64(total_bytes)),
				humanize.Bytes(uint64(pool_speed)),
				humanize.Ftoa(float64(int(byte_progress*10000))/100),
				timeRemaining(pool_speed, total_bytes-moved_bytes),
			)
			time.Sleep(1 * time.Second)
		}
	}()
	PrintProgress(file_finished_print, status_update, all_done)
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
				defer worker_waiter.Done()
				//dialer := bind
				conn, err := tls.Dial(
					"tcp",
					fmt.Sprintf("%s:%d", hostname, port),
					&tls.Config{
						InsecureSkipVerify: true,
					},
					// need to check sig too
				)
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
				worker_server.Insert(stream_writer.Socket) // also tag as a peer
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
