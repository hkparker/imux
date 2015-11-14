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

func NewAuthRequest(data []byte) interface{} {
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

func NewMessage(data []byte) interface{} {
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

func NewWorkerReady(data []byte) interface{} {
	worker_ready := &WorkerReady{}
	err := json.Unmarshal(data, &worker_ready)
	if err != nil {
		return nil
	}
	return worker_Ready
}

type Chunk struct {
	Filename string
	ID       int
	Data     string
}

func NewChunk(data []byte) interface{} {
	chunk := &Chunk{}
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		return nil
	}
	return chunk
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
		reflect.TypeOf(Chunk{}),
		reflect.TypeOf(&Chunk{}),
		NewChunk,
	)
	return type_store
}
