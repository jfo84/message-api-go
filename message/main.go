package message

import (
	"net/http"

	"github.com/jfo84/message-api-go/client"
	"github.com/jfo84/message-api-go/mockclient"
)

// Controller - For re-use of *messagebird.Client
type Controller struct {
	clientWrap *client.Wrapper
}

// TestController - For re-use of *mockclient.MockWrapper
type TestController struct {
	clientWrap *mockclient.MockWrapper
}

// NewController is a constructor for initializing with a *client.Wrapper
func NewController(clientWrap *client.Wrapper) *Controller {
	return &Controller{clientWrap: clientWrap}
}

// NewTestController is a constructor for initializing with a *client.Wrapper
func NewTestController(clientWrap *mockclient.MockWrapper) *TestController {
	return &TestController{clientWrap: clientWrap}
}

// Post writes a message to MessageBird and then responds with the http.ResponseWriter
func (mc *Controller) Post(w http.ResponseWriter, r *http.Request) {
	mc.clientWrap.PostMessage(w, r)
}

// Post writes a message to the fake MessageBird client and then responds with the http.ResponseWriter
func (mc *TestController) Post(w http.ResponseWriter, r *http.Request) {
	mc.clientWrap.PostMessage(w, r)
}
