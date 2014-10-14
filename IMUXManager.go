//package multiplexity
package main

type IMUXManager struct {
	// server for recieving connections
	// slice of workers
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
