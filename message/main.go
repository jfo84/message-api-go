package message

import (
	"net/http"

	"github.com/jfo84/message-api-go/client"
)

// Controller - For re-use of *messagebird.Client
type Controller struct {
	clientWrap *client.Wrapper
}

// NewController is a constructor for initializing with a *client.Wrapper
func NewController(clientWrap *client.Wrapper) *Controller {
	return &Controller{clientWrap: clientWrap}
}

// Post writes a message to MessageBird and then responds with the http.ResponseWriter
func (mc *Controller) Post(w http.ResponseWriter, r *http.Request) {
	mc.clientWrap.PostMessage(w, r)
}
