//package multiplexity
package main

import(
	"fmt"
	"github.com/twinj/uuid"
)

type TransferQueue struct {
	Client Host
	Server Host
	State IMUXState
	Messages chan string
	Pending []Transfer
	SessionKey string
}

// IMUXState object is created and points to a TransferQueue
// when the imux config need be adjusted, adjust the state
// the state is pointing at a transfer queue and from there can grab hosts
// it can then command the hosts to behave
// OR, this is just part of the transfer queue in the first place
// if so, I need to hold the imux state in the transfer queue
// could make the methods on type tranfer queue, then have those refernce the queue imuxstate



func (queue *TransferQueue) Open(state IMUXState) (err error) {
	queue.Messages = make(chan string, 50)
	queue.Pending = make([]Transfer, 0)
	queue.SessionKey = uuid.NewV4().String()
	queue.Messages <- fmt.Sprintf("Opening transfer queue between %s and %s (%d sockets)",
								  queue.Client.Hostname,
								  queue.Server.Hostname,
								  queue.Server.Port)
	err = queue.Server.RecieveIMUXSession(state.ServerConfig())
	if err != nil {
		queue.Messages <- fmt.Sprintf("Failed to open transfer queue: %s could not receive session: %s",
									  queue.Server.Hostname,
									  err)
		return
	}
	err = queue.Client.CreateIMUXSession(state.ClientConfig())
	if err != nil {
		queue.Messages <- fmt.Sprintf("Failed to open transfer queue: %s could not create session: %s",
									  queue.Client.Hostname,
									  err)
		return
	}
	return
}

func (queue *TransferQueue) Status() int {
	return 0
}

func (queue *TransferQueue) UpdateChunkSize() int {
	// tell both the client and server to use the new chunk size
	return 0
}

//TransferQueue.UpdateRecycling()
//TransferQueue.AddTransfer()
// increase/decrease worker size

func main(){
	queue := TransferQueue{}
	queue.Client = Host{}					// defined previously
	queue.Server = Host{}					// defined previously
	queue.Open()
	fmt.Println(<- queue.Messages)
}
