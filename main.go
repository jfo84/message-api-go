package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	messagebird "github.com/messagebird/go-rest-api"
)

// Message represents our version of an SMS message
type Message struct {
	Recipient  string `json:"recipient"`
	Originator string `json:"originator"`
	Data       string `json:"message"`
}

func createCounterString(num int, counter int) string {
	var buffer bytes.Buffer

	buffer.WriteString(" ")
	buffer.WriteString(string(num))
	buffer.WriteString("(")
	buffer.WriteString(string(counter))
	buffer.WriteString(")")

	return buffer.String()
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		panic(err)
	}

	// MessageBird client
	client := messagebird.New("test")
	params := &messagebird.MessageParams{}

	// Originator - no changes

	// Recipients
	recipients := make([]string, 1)
	recipients[0] = message.Recipient

	// Body
	var messageRunes []rune
	messageCounter := 0

	// 154 is 160 runes minus 6 to add a message like " (1/3)"
	const runeLimit = 154

	dataRunes := []rune(message.Data)
	dataLen := len(dataRunes)

	// This "just works" and we don't have to use floats
	messageNum := (dataLen / runeLimit) + 1

	for idx, rune := range dataRunes {
		messageRunes[idx] = rune

		if len(messageRunes) == 154 {
			body := string(messageRunes)
			// Runes vs. bytes doesn't matter here since we're adding a known string that
			// doesn't have multi-byte runes, e.g. Chinese characters
			counterString := createCounterString(messageNum, messageCounter)
			// String concatenation is fastest and zero-allocation last time
			// I checked: https://gist.github.com/dtjm/c6ebc86abe7515c988ec
			body = body + counterString

			message, err := client.NewMessage(
				message.Originator,
				recipients,
				body,
				params)

			if err != nil {
				return
			}

			// Clear the slice of runes and increment the messageCounter
			messageRunes := messageRunes[:0]
			messageCounter++
		}

	}

	defer r.Body.Close()
}

func main() {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/messages", messageHandler).Methods("POST")

	addr := ":7000"
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
