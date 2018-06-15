package message

import (
	"net/http"
)

// Controller - For re-use of *messagebird.Client
type Controller struct {
	clientWrap *client.Wrapper
}

// NewController is a constructor for initializing with a *db.Wrapper
func NewController(clientWrap *client.Wrapper) *Controller {
	return &Controller{clientWrap: clientWrap}
}

// Post writes a message to MessageBird with the http.ResponseWriter
func (mc *Controller) Post(w http.ResponseWriter, r *http.Request) {
	mc.clientWrap.PostMessage(w, r)
}
