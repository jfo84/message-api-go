package main

import (
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
	"github.com/jfo84/message-api-go/client"
	"github.com/jfo84/message-api-go/message"
)

func main() {
	r := mux.NewRouter().StrictSlash(true)
	clientWrap := client.New()

	messageController := message.NewController(clientWrap)
	// Rate limit to 1 request per second on this route
	rateLimitedHandler := tollbooth.LimitFuncHandler(tollbooth.NewLimiter(1, nil), messageController.Post)
	r.Handle("/messages", rateLimitedHandler).Methods("POST")

	addr := ":7000"
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
