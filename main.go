package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jfo84/message-api-go/api/message"
	messagebird "github.com/messagebird/go-rest-api"
)

func main() {
	r := mux.NewRouter().StrictSlash(true)
	client := messagebird.New("test_22sWNIUrVGyI3J2IheE4SpwUc")

	// TODO: Needs client wrapper, not client
	messageController := message.NewController(client)
	r.HandleFunc("/messages", messageHandler).Methods("POST")

	addr := ":7000"
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
