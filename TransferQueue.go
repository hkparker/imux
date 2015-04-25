package main

import(
	"fmt"
	"github.com/twinj/uuid"
)

type TransferQueue struct {
	Client *Host
	Server *Host
	Messages chan string
	//Pending []Transfer
	SessionKey string
}

func CreateTransferQueue(client, server *Host) (TransferQueue, error) {//, sockets map[string]int, recycling bool, chunk_size int) (TransferQueue, error) {
	queue := TransferQueue{}
	queue.Client = client
	queue.Server = server
	queue.Messages = make(chan string, 50)
	//queue.Pending = make([]Transfer, 0)
	queue.SessionKey = uuid.NewV4().String()
	queue.Messages <- fmt.Sprintf("Opening transfer queue between %s and %s",
								  queue.Client.IP,
								  queue.Server.IP)
	//resp = queue.Server.QuerySession("createsession ")
	//if resp == nil {
	//	queue.Messages <- fmt.Sprintf("Failed to open transfer queue: %s could not receive session: %s",
	//								  queue.Server.Hostname,
	//								  err)
	//	return queue, errors.New()
	//}
	// parse resp
	//err = queue.Client.QuerySession("recievesession")
	//if err != nil {
	//	queue.Messages <- fmt.Sprintf("Failed to open transfer queue: %s could not create session: %s",
	//								  queue.Client.Hostname,
	//								  err)
	//	return
	//}
	return queue, nil
}

func (queue *TransferQueue) Status() int {
	return 0
}

//func (host *Host) ServeFile() int {
	//return 0
//}

//func (host *Host) RecieveFile() int {
	//return 0
//}

func (queue *TransferQueue) UpdateChunkSize() int {
	// tell both the client and server to use the new chunk size
	return 0
}

//TransferQueue.UpdateRecycling()
//TransferQueue.AddTransfer()
