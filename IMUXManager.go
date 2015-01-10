package main

type IMUXManager struct {
	State string
	// server for recieving connections
	// slice of workers	// this slice gets sliced to reduce the number of workers in the next transfer (go routines are started for each fored on the slicec in eace file.)
						// or, go routines in each chunk check a channel to see if they have been killed, and if so they stop moving chunks, allowing for a per chunk host reduction.
						// maybe even a chan runs and checks when a goroutine exits.  when it does it updates the worker list.  that way when they die worker list stays informed.
						
	
}

func (manager *IMUXManager) CreateSockets(bind_ips map[string]int) int {
	workers_created := 0
	// Lock manager?
	// create a dialer object for each ip in the hashof ip:count
	// iterate over bind_ips
	//   create a dialer for key
	//   for value times
	//     create an imuxsock
	//     open it with the dialer
	//     if all that worked out
	//       add it to the workers slice
	//       increment workers_created
	// Unlock manager
	return workers_created
}

func (manager *IMUXManager) RecieveSockets() int {
	// create server here if needed
	return 0
}

func (manager *IMUXManager) ServeFile(filename string, starting_position int) int {
	// Lock manager
	// Create read queue
	// for each manager.Workers
	//   go serve the read queue
	// wait for all the goroutines to push 
	return 0
}

func (manager *IMUXManager) DownloadFile() int {
	// goroutine for each socket
	// if sockets report error, ?
	return 0
}

func (manager *IMUXManager) UpdateRecycling() int {
	return 0
}

func (manager *IMUXManager) UpdateChunkSize() int {
	return 0
}

func (manager *IMUXManager) Close() int {
	return 0
}
