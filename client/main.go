package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jfo84/message-api-go/utils"
	messagebird "github.com/messagebird/go-rest-api"
	"github.com/stretchr/testify/mock"
)

// Message represents an interface for unmarshalling a received SMS message
type Message struct {
	Recipient  int    `json:"recipient"`
	Originator string `json:"originator"`
	Data       string `json:"message"`
}

// Wrapper is a wrapper over any Client that implements NewMessage
type Wrapper struct {
	Client
}

// Client is the piece of the *messagebird.Client interface that the mock client implements
type Client interface {
	NewMessage(originator string, recipients []string, body string, msgParams *messagebird.MessageParams) (*messagebird.Message, error)
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

	mbMessage := args.Get(0).(messagebird.Message)
	return &mbMessage, args.Error(1)
}

// PostMessage posts a message to MessageBird and responds with the results
func (wrap *Wrapper) PostMessage(w http.ResponseWriter, r *http.Request) {
	// 160 runes for an SMS message
	const runeLimit = 160

	var err error
	var message *Message

	decoder := json.NewDecoder(r.Body)

	message, err = decodeAndValidateMessage(decoder)

	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		errBytes := []byte(err.Error())
		fmt.Println(errBytes)
		w.Write(errBytes)
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

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errBytes)
		return
	}

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

	w.WriteHeader(http.StatusCreated)
	w.Write(messageJSON)
}

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
	params := &messagebird.MessageParams{}

	return wrap.postToMessageBird(message.Originator, recipients, body, params)
}

func (wrap *Wrapper) sendConcatMessage(message *Message, dataRunes []rune) ([]*messagebird.Message, error) {
	// 153 runes gives us space for a UDH header
	const concatRuneLimit = 153

	var body string
	recipients := generateRecipientsSlice(message.Recipient)

	messageRunes := make([]rune, concatRuneLimit)
	messageRuneIdx := 0
	messageCounter := 0
	dataLen := len(dataRunes)

	// This "just works" and we don't have to use floats and rounding
	messageNum := (dataLen / concatRuneLimit) + 1
	mbMessages := make([]*messagebird.Message, messageNum)

	// Generate a random string for identifying the connected SMS messages
	refNumber := utils.RandHex()

	for idx, dataRune := range dataRunes {
		if (messageRuneIdx == concatRuneLimit) || (idx == dataLen-1) {
			// Don't cut off the last value if we're at the end of the slice
			if idx == dataLen-1 {
				messageRunes[messageRuneIdx] = dataRune
			}

			body = string(messageRunes)

			udhString := utils.GenerateUDHString(refNumber, messageNum, messageCounter+1)
			typeDetails := messagebird.TypeDetails{"udh": udhString}
			params := &messagebird.MessageParams{Type: "binary", TypeDetails: typeDetails}

			mbMessage, err := wrap.postToMessageBird(message.Originator, recipients, body, params)
			// Bail out if one of the messages fails
			if err != nil {
				return nil, err
			}

			// Build a slice of messages for our response
			mbMessages[messageCounter] = mbMessage

			// Bail if we're at the end of the slice
			if idx == dataLen-1 {
				messageRunes[messageRuneIdx] = dataRune
			}

			messageCounter++
			// When you serialize a slice with extra length it spits out garbage
			if messageNum == messageCounter+1 {
				remainingRunes := dataLen % concatRuneLimit
				messageRunes = make([]rune, remainingRunes)
			} else {
				messageRunes = make([]rune, concatRuneLimit)
			}
			messageRuneIdx = 0
		}
		messageRunes[messageRuneIdx] = dataRune
		messageRuneIdx++
	}

	return mbMessages, nil
}

func generateRecipientsSlice(recipient int) []string {
	recipients := make([]string, 1)
	recipients[0] = strconv.Itoa(recipient)
	return recipients
}

// New returns a *Wrapper for re-use of the client object
func New() *Wrapper {
	client := messagebird.New("test_22sWNIUrVGyI3J2IheE4SpwUc")
	wrapper := &Wrapper{Client: client}

	return wrapper
}
