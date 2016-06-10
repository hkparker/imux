package common

import (
	"encoding/json"
	"github.com/hkparker/TLJ"
	"reflect"
)

var type_store = tlj.TypeStore{}

type AuthRequest struct {
	Username string
	Password string
}

func NewAuthRequest(data []byte, _ tlj.TLJContext) interface{} {
	auth_request := &AuthRequest{}
	err := json.Unmarshal(data, &auth_request)
	if err != nil {
		return nil
	}
	return auth_request
}

type Message struct {
	String string
}

func NewMessage(data []byte, _ tlj.TLJContext) interface{} {
	message := &Message{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		return nil
	}
	return message
}

type WorkerAuth struct {
	Nonce string
}

func NewWorkerAuth(data []byte, _ tlj.TLJContext) interface{} {
	worker_auth := &WorkerAuth{}
	err := json.Unmarshal(data, &worker_auth)
	if err != nil {
		return nil
	}
	return worker_auth
}

type TransferChunk struct {
	Filename    string
	Destination string
	ID          int
	Data        []byte
}

func NewTransferChunk(data []byte, _ tlj.TLJContext) interface{} {
	chunk := &TransferChunk{}
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		return nil
	}
	return chunk
}

type Command struct {
	Command string
	Args    []string
}

func NewCommand(data []byte, _ tlj.TLJContext) interface{} {
	command := &Command{}
	err := json.Unmarshal(data, &command)
	if err != nil {
		return nil
	}
	return command
}

func BuildTypeStore() tlj.TypeStore {
	type_store := tlj.NewTypeStore()
	type_store.AddType(
		reflect.TypeOf(Message{}),
		reflect.TypeOf(&Message{}),
		NewMessage,
	)
	type_store.AddType(
		reflect.TypeOf(AuthRequest{}),
		reflect.TypeOf(&AuthRequest{}),
		NewAuthRequest,
	)
	type_store.AddType(
		reflect.TypeOf(WorkerAuth{}),
		reflect.TypeOf(&WorkerAuth{}),
		NewWorkerAuth,
	)
	type_store.AddType(
		reflect.TypeOf(TransferChunk{}),
		reflect.TypeOf(&TransferChunk{}),
		NewTransferChunk,
	)
	type_store.AddType(
		reflect.TypeOf(Command{}),
		reflect.TypeOf(&Command{}),
		NewCommand,
	)
	return type_store
}
