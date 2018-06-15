package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jfo84/message-api-go/message"
)

func main() {
	r := mux.NewRouter().StrictSlash(true)
	clientWrap := client.New()

	messageController := message.NewController(clientWrap)
	r.HandleFunc("/messages", messageControler.Post).Methods("POST")

	addr := ":7000"
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
