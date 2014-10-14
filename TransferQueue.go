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



func (queue TransferQueue) Open() int {
	queue.Messages <- fmt.Sprintf("Opening transfer queue between %s and %s (%d sockets)",
								  queue.Client.Hostname,
								  queue.Server.Hostname,
								  queue.Server.Port)
	// here we tell the hosts to recieve or send the sockets
	return 0
}

func (queue TransferQueue) Status() int {
	return 0
}

func (queue TransferQueue) AddTransfer() int {
	return 0
}

func (queue TransferQueue) ListTransfers() int {
	return 0
}

func (queue TransferQueue) RemoveTransfer() int {
	return 0
}

// functions to reorder transfers?
// functions to edit transfer details (like destination name / location )?
// Transfer struct?
// Expose as slice?

//TransferQueue.UpdateChunkSize()
//TransferQueue.UpdateRecycling()
//TransferQueue.Pause()
//TransferQueue.Resume()
//TransferQueue.Clear()

func main(){
	queue := TransferQueue{}
	queue.Client = Host{}					// defined previously
	queue.Server = Host{}					// defined previously
	queue.State = IMUXState{}				// defined here, saved (or tmp)
	queue.Messages = make(chan string, 50)
	queue.Pending = make([]Transfer, 100)	// defined here, may need to increase size
	queue.SessionKey = uuid.NewV4().String()
	
	queue.Open()
	fmt.Println(<- queue.Messages)
}
