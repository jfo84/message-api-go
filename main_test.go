package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jfo84/message-api-go/client"
	"github.com/jfo84/message-api-go/message"
	"github.com/jfo84/message-api-go/utils"
	messagebird "github.com/messagebird/go-rest-api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func setupSendMessageContext(body string) (client.Wrapper, string, []byte) {
	// Build a *messagebird.Message object to return except for the body
	// which we pass as an arg in each test

	// Originator
	originator := "MessageBird"

	// Recipients
	recipient := 31612345678
	recipients := make([]string, 1)
	recipients[0] = strconv.Itoa(recipient)

	nowTime := time.Now()
	item := messagebird.Recipient{
		Recipient:      recipient,
		Status:         "sent",
		StatusDatetime: &nowTime,
	}
	items := make([]messagebird.Recipient, 1)
	items[0] = item

	mbRecipients := messagebird.Recipients{
		TotalCount:               1,
		TotalSentCount:           1,
		TotalDeliveredCount:      1,
		TotalDeliveryFailedCount: 0,
		Items: items,
	}

	mbMessage := messagebird.Message{
		Originator: originator,
		Recipients: mbRecipients,
		Body:       body,
	}

	// Params
	params := &messagebird.MessageParams{}

	mockClient := new(client.Mock)
	mockClient.On("NewMessage",
		originator,
		recipients,
		body,
		params).Return(mbMessage, nil)

	clientWrap := client.Wrapper{Client: mockClient}

	reqBody := "{\"recipient\":" + strconv.Itoa(recipient) + ",\"originator\":\"" + originator + "\",\"message\":\"" + body + "\"}"

	messageJSON, err := json.Marshal(mbMessage)
	if err != nil {
		panic(err)
	}

	return clientWrap, reqBody, messageJSON
}

var _ = Describe("TestMessageApiGo", func() {
	var (
		req *http.Request
		err error
	)

	router := mux.NewRouter().StrictSlash(true)

	Context("Messages", func() {

		Context("SendMessage", func() {

			It("Should correctly post to MessageBird and return a serialized message response", func() {
				body := "This is a test message."
				clientWrap, reqBody, messageJSON := setupSendMessageContext(body)

				reader := bytes.NewBufferString(reqBody)

				req, err = http.NewRequest("POST", "/messages", reader)
				if err != nil {
					panic(err)
				}

				recorder := httptest.NewRecorder()

				messageController := message.NewController(&clientWrap)
				router.HandleFunc("/messages", messageController.Post)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Body.Bytes()).To(Equal(messageJSON))
				Expect(recorder.Code).To(Equal(http.StatusCreated))
			})

			It("Should correctly return an error with an invalid message", func() {
				body := ""
				clientWrap, reqBody, _ := setupSendMessageContext(body)

				reader := bytes.NewBufferString(reqBody)

				req, err = http.NewRequest("POST", "/messages", reader)
				if err != nil {
					panic(err)
				}

				recorder := httptest.NewRecorder()

				messageController := message.NewController(&clientWrap)
				router.HandleFunc("/messages", messageController.Post)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Body.Bytes()).To(Equal([]byte("Invalid message: You must have a valid message body under the \"message\" key")))
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("SendConcatMessage", func() {
			// Body
			firstMessageBody := utils.RandStringRunes(153)
			secondMessageBody := "0123456789"
			body := firstMessageBody + secondMessageBody

			// Originator
			originator := "MessageBird"

			// Recipients
			recipient := 31612345678
			recipients := make([]string, 1)
			recipients[0] = strconv.Itoa(recipient)

			nowTime := time.Now()
			item := messagebird.Recipient{
				Recipient:      recipient,
				Status:         "sent",
				StatusDatetime: &nowTime,
			}
			items := make([]messagebird.Recipient, 1)
			items[0] = item

			mbRecipients := messagebird.Recipients{
				TotalCount:               1,
				TotalSentCount:           1,
				TotalDeliveredCount:      1,
				TotalDeliveryFailedCount: 0,
				Items: items,
			}

			mbMessage := messagebird.Message{
				Originator: originator,
				Recipients: mbRecipients,
				Body:       body,
			}

			mockClient := new(client.Mock)

			var udhString string
			var typeDetails messagebird.TypeDetails

			refNumber := utils.RandHex()

			udhString = utils.GenerateUDHString(refNumber, 2, 1)
			typeDetails = messagebird.TypeDetails{"udh": udhString}
			params := &messagebird.MessageParams{Type: "binary", TypeDetails: typeDetails}

			mockClient.On("NewMessage",
				originator,
				recipients,
				firstMessageBody,
				params).Return(mbMessage, nil)

			udhString = utils.GenerateUDHString(refNumber, 2, 2)
			typeDetails = messagebird.TypeDetails{"udh": udhString}
			params = &messagebird.MessageParams{Type: "binary", TypeDetails: typeDetails}

			mockClient.On("NewMessage",
				originator,
				recipients,
				secondMessageBody,
				params).Return(mbMessage, nil)

			It("Should correctly post to MessageBird and return a serialized message response", func() {
				Skip("Skipping SendConcatMessage test")

				reqBody := "{\"recipient\":" + strconv.Itoa(recipient) + ",\"originator\":\"" + originator + "\",\"message\":\"" + body + "\"}"
				reader := bytes.NewBufferString(reqBody)

				req, err = http.NewRequest("POST", "/messages", reader)
				if err != nil {
					panic(err)
				}

				recorder := httptest.NewRecorder()

				clientWrap := client.Wrapper{Client: mockClient}

				messageController := message.NewController(&clientWrap)
				router.HandleFunc("/messages", messageController.Post)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Body.String()).To(Equal("foo"))
				Expect(recorder.Code).To(Equal(http.StatusCreated))
			})
		})
	})
})
