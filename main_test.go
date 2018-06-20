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
	messagebird "github.com/messagebird/go-rest-api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestMessageApiGo", func() {
	var (
		req *http.Request
		err error
	)

	router := mux.NewRouter().StrictSlash(true)

	Context("Messages", func() {
		// Build a *messagebird.Message object to return

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

		// Body
		body := "This is a test message."

		mbMessage := messagebird.Message{
			Originator: originator,
			Recipients: mbRecipients,
			Body:       body,
		}

		// Params
		params := &messagebird.MessageParams{}

		var mockClient *client.Mock

		BeforeEach(func() {
			mockClient = new(client.Mock)
			mockClient.On("NewMessage",
				originator,
				recipients,
				body,
				params).Return(mbMessage, nil)
		})

		It("Should correctly post to MessageBird and return a serialized message response", func() {
			reqBody := "{\"recipient\":31612345678,\"originator\":\"MessageBird\",\"message\":\"This is a test message.\"}"
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

			messageJSON, err := json.Marshal(mbMessage)
			if err != nil {
				panic(err)
			}

			Expect(recorder.Body.Bytes()).To(Equal(messageJSON))
			Expect(recorder.Code).To(Equal(http.StatusCreated))
		})

		It("Should correctly return an error with an invalid message", func() {
			reqBody := "{\"recipient\":31612345678,\"originator\":\"MessageBird\"}"
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

			Expect(recorder.Body.Bytes()).To(Equal([]byte("Invalid message: You must have a valid message body under the \"message\" key")))
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
