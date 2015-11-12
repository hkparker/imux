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

type Chunk struct {
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
	return type_store
}
