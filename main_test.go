package main

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/jfo84/message-api-go/client"
	"github.com/jfo84/message-api-go/message"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestMessageApiGo", func() {
	var (
		req *http.Request
		err error
	)

	router := mux.NewRouter().StrictSlash(true)
	clientWrap := client.New()

	Context("Pull Requests", func() {
		It("Should correctly return a pull", func() {
			req, err = http.NewRequest("POST", "/messages", nil)

			recorder := httptest.NewRecorder()

			messageController := message.NewController(clientWrap)
			router.HandleFunc("/messages", messageController.Post)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusCreated))
		})
	})
})
