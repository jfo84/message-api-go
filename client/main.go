package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	messagebird "github.com/jfo84/go-rest-api"
	"github.com/jfo84/message-api-go/utils"
	"github.com/stretchr/testify/mock"
)

// Message represents an interface for unmarshalling a received SMS message
type Message struct {
	Recipient  int    `json:"recipient"`
	Originator string `json:"originator"`
	Data       string `json:"message"`
}

// Mock is used to mock requests to the *messagebird.Client
type Mock struct {
	mock.Mock
}

// NewMessage records the args and returns a mock *messagebird.Message
func (m *Mock) NewMessage(
	originator string,
	recipients []string,
	body string,
	msgParams *messagebird.MessageParams) (*messagebird.Message, error) {
	args := m.Called(originator, recipients, body, msgParams)

	return args.Get(0).(*messagebird.Message), args.Error(1)
}

// Wrapper is a wrapper over any Client that implements NewMessage
type Wrapper struct {
	Client
}

// Client is the piece of the *messagebird.Client interface that the mock client implements
type Client interface {
	NewMessage(originator string, recipients []string, body string, msgParams *messagebird.MessageParams) (*messagebird.Message, error)
}

// 160 runes for an SMS message
const runeLimit = 160

// 153 runes gives us space for a UDH header
const concatRuneLimit = 153

func decodeAndValidateMessage(decoder *json.Decoder) (*Message, error) {
	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		return &message, errors.New("Invalid JSON. To post a message you must send JSON in the format: " +
			"{\"recipient\":31612345678,\"originator\":\"MessageBird\",\"message\":\"This is a test message.\"}")
	}

	// 0 is the zero value for int
	if message.Recipient == 0 {
		return &message, errors.New("Invalid message: You must have a valid recipient phone number under the \"recipient\" key")
	}

	if len(message.Data) == 0 {
		return &message, errors.New("Invalid message: You must have a valid message body under the \"message\" key")
	}

	if len(message.Originator) == 0 {
		return &message, errors.New("Invalid message: You must have a valid originator under the \"originator\" key")
	}

	return &message, nil
}

// For generating the UDH string for concatenated messages
// https://en.wikipedia.org/wiki/Concatenated_SMS#Sending_a_concatenated_SMS_using_a_User_Data_Header
func generateUDHString(ref string, num int, counter int) string {
	return "050003" + ref + utils.IntToHex(num) + utils.IntToHex(counter)
}

func generateRecipientsSlice(recipient int) []string {
	recipients := make([]string, 1)
	recipients[0] = strconv.Itoa(recipient)
	return recipients
}

// PostMessage posts a message to MessageBird and responds with the results
func (wrap *Wrapper) PostMessage(w http.ResponseWriter, r *http.Request) {
	var err error
	var message *Message

	decoder := json.NewDecoder(r.Body)

	message, err = decodeAndValidateMessage(decoder)

	defer r.Body.Close()

	if err != nil {
		errBytes := []byte(err.Error())
		fmt.Println(errBytes)
		w.Write(errBytes)
		// TODO: This doesn't work for some reason
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Validated")

	dataRunes := []rune(message.Data)
	var mbMessage *messagebird.Message
	var mbMessages []*messagebird.Message

	if len(dataRunes) < runeLimit {
		mbMessage, err = wrap.sendMessage(message, dataRunes)
	} else {
		mbMessages, err = wrap.sendConcatMessage(message, dataRunes)
	}

	if err != nil {
		errBytes := []byte(err.Error())

		w.Write(errBytes)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		var messageJSON []byte
		if len(mbMessages) == 0 {
			messageJSON, err = json.Marshal(mbMessage)
		} else {
			messageJSON, err = json.Marshal(mbMessages)
		}
		// Bail out because this shouldn't happen
		if err != nil {
			panic(err)
		}

		w.Write(messageJSON)
		w.WriteHeader(http.StatusCreated)
	}
}

func (wrap *Wrapper) postToMessageBird(
	originator string,
	recipients []string,
	body string,
	params *messagebird.MessageParams) (*messagebird.Message, error) {
	mbMessage, err := wrap.Client.NewMessage(
		originator,
		recipients,
		body,
		params)

	fmt.Println(mbMessage)
	fmt.Println(err)
	return mbMessage, err
}

func (wrap *Wrapper) sendMessage(message *Message, messageRunes []rune) (*messagebird.Message, error) {
	recipients := generateRecipientsSlice(message.Recipient)
	body := string(messageRunes)
	fmt.Println(recipients)
	fmt.Println(body)
	return wrap.postToMessageBird(message.Originator, recipients, body, nil)
}

func (wrap *Wrapper) sendConcatMessage(message *Message, dataRunes []rune) ([]*messagebird.Message, error) {
	var body string
	var mbMessages []*messagebird.Message
	recipients := generateRecipientsSlice(message.Recipient)

	messageRunes := make([]rune, concatRuneLimit)
	messageCounter := 0
	dataLen := len(dataRunes)

	// This "just works" and we don't have to use floats and rounding
	messageNum := (dataLen / concatRuneLimit) + 1

	// Generate a random string for identifying the connected SMS messages
	refNumber := utils.RandHex()
	udhString := generateUDHString(refNumber, messageNum, messageCounter)

	// Tell the API we're sending binary data with a UDH header
	typeDetails := messagebird.TypeDetails{"udh": udhString}
	params := &messagebird.MessageParams{Type: "binary", TypeDetails: typeDetails}

	for idx, dataRune := range dataRunes {
		messageRunes[idx] = dataRune

		if (len(messageRunes) == concatRuneLimit) || (idx == dataLen-1) {
			body = string(messageRunes)

			mbMessage, err := wrap.postToMessageBird(message.Originator, recipients, body, params)
			// Bail out if one of the messages fails
			if err != nil {
				return nil, err
			}

			// Build a slice of messages for our response
			mbMessages[messageCounter] = mbMessage

			// Clear the slice of runes and increment the messageCounter
			messageRunes = messageRunes[:0]
			messageCounter++
		}
	}

	return mbMessages, nil
}

// New returns a *Wrapper for re-use of the client object
func New() *Wrapper {
	client := messagebird.New("test_22sWNIUrVGyI3J2IheE4SpwUc")
	wrapper := &Wrapper{Client: client}

	return wrapper
}
