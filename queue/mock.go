package queue

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
)

// Mock implementation of the Listener interface
// Reads messages from a file
// To be used in manual testing and functional testing
type mock struct {
	path   string
	logger log.FieldLogger
}

// NewMock creates a mock listener
func NewMock(path string, logger log.FieldLogger) Listener {
	return &mock{
		path:   path,
		logger: logger,
	}
}

func (m *mock) Listen() (<-chan *Message, <-chan error) {
	c := make(chan *Message)
	ack := make(chan *Message)
	e := make(chan error)

	go func() {
		data, err := ioutil.ReadFile(m.path)
		if err != nil {
			// It should not happen, let's not add confusion to our tests
			m.logger.WithError(err).Panic("Could not open fixture file")
		}

		var fixtures []*Message
		err = json.Unmarshal(data, &fixtures)
		if err != nil {
			m.logger.WithError(err).Panic("Invalid fixture file")
		}

		for _, msg := range fixtures {
			msg.ack = ack
			c <- msg
			time.Sleep(100 * time.Millisecond)
		}

		// push an error when no fixtures left
		e <- errors.New("No more messages")

		time.Sleep(100 * time.Millisecond)
		close(c)
	}()

	go func() {
		// listen to ack message but simply log them
		for msg := range ack {
			m.logger.WithField("ReceiptHandle", msg.ReceiptHandle).Debug("Mock listener: Ack message")
		}
	}()

	return c, e
}

func (m *mock) LastRequest() *time.Duration {
	return nil
}
