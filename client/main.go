package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	messagebird "github.com/messagebird/go-rest-api"
)

// Message represents an interface for unmarshalling a received SMS message
type Message struct {
	Recipient  int    `json:"recipient"`
	Originator string `json:"originator"`
	Data       string `json:"message"`
}

// Wrapper is a wrapper over messagebird.Client
type Wrapper struct {
	client *messagebird.Client
}

// 160 runes
const runeLimit = 160

// 153 runes gives us space for a UDH header
const concatRuneLimit = 153

// TODO
// func verifyMessage(message *Message) error {
// 	return error
// }

// func sendMessage(w http.ResponseWriter, mbMessage *messagebird.Message) {
// 	responseBytes := []byte("foo")
// 	w.Write(responseBytes)
// }

func generateUDHString(num int, counter int) string {
	var buffer bytes.Buffer

	// TODO
	// buffer.WriteString(" ")
	// buffer.WriteString(string(num))
	// buffer.WriteString("(")
	// buffer.WriteString(string(counter))
	// buffer.WriteString(")")

	return buffer.String()
}

func generateRecipientsSlice(recipient int) []string {
	recipients := make([]string, 1)
	recipients[0] = string(recipient)
	return recipients
}

// PostMessage posts a message to MessageBird and responds with the results
func (wrap *Wrapper) PostMessage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		panic(err)
	}

	dataRunes := []rune(message.Data)

	if len(dataRunes) < runeLimit {
		wrap.sendMessage(&message, dataRunes)
	} else {
		wrap.sendConcatMessage(&message, dataRunes)
	}
}

func (wrap *Wrapper) postToMessageBird(originator string, recipients []string, body string) {
	params := &messagebird.MessageParams{}
	mbMessage, err := wrap.client.NewMessage(
		originator,
		recipients,
		body,
		params)

	if err != nil {
		panic(err)
	}

	fmt.Println(mbMessage)
}

func (wrap *Wrapper) sendMessage(message *Message, messageRunes []rune) {
	recipients := generateRecipientsSlice(message.Recipient)
	body := string(messageRunes)

	wrap.postToMessageBird(message.Originator, recipients, body)
}

func (wrap *Wrapper) sendConcatMessage(message *Message, dataRunes []rune) {
	recipients := generateRecipientsSlice(message.Recipient)
	var body string

	messageRunes := make([]rune, concatRuneLimit)
	messageCounter := 0
	dataLen := len(dataRunes)

	// This "just works" and we don't have to use floats and rounding
	messageNum := (dataLen / concatRuneLimit) + 1

	for idx, dataRune := range dataRunes {
		messageRunes[idx] = dataRune

		if (len(messageRunes) == concatRuneLimit) || (idx == dataLen-1) {
			body = string(messageRunes)
			// Runes vs. bytes doesn't matter here since we're adding a known string that
			// doesn't have multi-byte runes, e.g. Chinese characters
			udhString := generateUDHString(messageNum, messageCounter)
			// It looks dirty but string concatenation is fastest and
			// zero-allocation: https://gist.github.com/dtjm/c6ebc86abe7515c988ec
			body = udhString + body

			wrap.postToMessageBird(message.Originator, recipients, body)

			// Clear the slice of runes and increment the messageCounter
			messageRunes = messageRunes[:0]
			messageCounter++
		}

	}
}
