//package multiplexity
package main

type IMUXManager struct {
	// server for recieving connections
	// slice of workers	// this slice gets sliced to reduce the number of workers in the next transfer (go routines are started for each fored on the slicec in eace file.)
						// or, go routines in each chunk check a channel to see if they have been killed, and if so they stop moving chunks, allowing for a per chunk host reduction.
						// maybe even a chann runs and checks when a goroutine exits.  when it does it updates the worker list.  that way when they die worker list stays informed.
						
}

func (manager IMUXManager) CreateSockets() int {
	return 0
}

func (manager IMUXManager) RecieveSockets() int {
	// create server
	return 0
}

func (manager IMUXManager) ServeFile() int {
	return 0
}

func (manager IMUXManager) DownloadFile() int {
	return 0
}

func (manager IMUXManager) UpdateRecycling() int {
	return 0
}

func (manager IMUXManager) UpdateChunkSize() int {
	return 0
}

func (manager IMUXManager) Close() int {
	return 0
}
