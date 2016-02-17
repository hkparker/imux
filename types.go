package main

import (
	"encoding/json"
	"github.com/hkparker/TLJ"
	"reflect"
)

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

type WorkerReady struct {
	Nonce string
}

func NewWorkerReady(data []byte, _ tlj.TLJContext) interface{} {
	worker_ready := &WorkerReady{}
	err := json.Unmarshal(data, &worker_ready)
	if err != nil {
		return nil
	}
	return worker_ready
}

type TransferChunk struct {
	Filename string
	ID       int
	Data     []byte
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
		reflect.TypeOf(WorkerReady{}),
		reflect.TypeOf(&WorkerReady{}),
		NewWorkerReady,
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
