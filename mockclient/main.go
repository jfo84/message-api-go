package mockclient

import (
	"strconv"
	"time"

	messagebird "github.com/messagebird/go-rest-api"
	"github.com/stretchr/testify/mock"
)

// MockWrapper is a wrapper over *MockClient
type MockWrapper struct {
	client *MockClient
}

// MockClient is used to mock requests to the *messagebird.Client
type MockClient struct {
	mock.Mock
}

// NewMessage records the args and returns a mock *messagebird.Message
func (m *MockClient) NewMessage(
	originator string,
	recipients []string,
	body string,
	msgParams *messagebird.MessageParams) (*messagebird.Message, error) {
	args := m.Called(originator, recipients, body, msgParams)

	// Build a *messagebird.Message object to return
	nowTime := time.Now()
	recipient, _ := strconv.Atoi(recipients[0])
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

	return &mbMessage, args.Error(1)
}
