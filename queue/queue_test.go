package queue

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fchoquet/golibs/metrics"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestItListensToMessages(t *testing.T) {
	assert := assert.New(t)

	service := mockSQSClient{
		messages: []*sqs.Message{
			{
				Body:          aws.String("this is message #0"),
				ReceiptHandle: aws.String("receipt-handle-0"),
				MessageId:     aws.String("message-id-0"),
			},
			{
				Body:          aws.String("this is message #1"),
				ReceiptHandle: aws.String("receipt-handle-1"),
				MessageId:     aws.String("message-id-1"),
			},
		},
	}

	q := queue{
		url:     "test-url",
		service: &service,
		mutex:   &sync.Mutex{},
		logger:  logrus.StandardLogger(),
		metrics: metrics.Default,
	}

	c, _ := q.Listen()

	msg := <-c
	assert.Equal("this is message #0", msg.Body)
	assert.Equal("receipt-handle-0", msg.ReceiptHandle)
	assert.Equal("message-id-0", msg.MessageID)

	msg = <-c
	assert.Equal("this is message #1", msg.Body)
	assert.Equal("receipt-handle-1", msg.ReceiptHandle)
	assert.Equal("message-id-1", msg.MessageID)
}

func TestItReportsErrors(t *testing.T) {
	assert := assert.New(t)

	q := queue{
		url:     "test-url",
		service: &mockSQSClient{},
		mutex:   &sync.Mutex{},
		logger:  logrus.StandardLogger(),
		metrics: metrics.Default,
	}

	// mockSQSClient triggers an error when it have no messages left to create
	// we need an error, so let's use this one
	_, e := q.Listen()

	err := <-e
	assert.Equal("could not receive more messages", err.Error())
}

func TestItDeletesAcknowledgedMessages(t *testing.T) {
	assert := assert.New(t)

	mutex := &sync.Mutex{}

	service := mockSQSClient{
		messages: []*sqs.Message{
			{
				Body:          aws.String("this is message #0"),
				ReceiptHandle: aws.String("receipt-handle-0"),
				MessageId:     aws.String("message-id-0"),
			},
		},
		mutex: mutex,
	}

	q := queue{
		url:     "test-url",
		service: &service,
		mutex:   &sync.Mutex{},
		logger:  logrus.StandardLogger(),
		metrics: metrics.Default,
	}

	c, _ := q.Listen()

	msg := <-c
	assert.Equal("this is message #0", msg.Body)
	assert.Equal("receipt-handle-0", msg.ReceiptHandle)
	assert.Equal("message-id-0", msg.MessageID)

	msg.Ack()

	// Let's wait for the Ack to really happen
	time.Sleep(1 * time.Millisecond)

	// we need a mutex to protect access to service.deletedMessages
	mutex.Lock()
	assert.Equal(1, len(service.deletedMessages))
	assert.Equal("receipt-handle-0", *service.deletedMessages[0].ReceiptHandle)
	mutex.Unlock()
}

func TestItKeepsTrackOfLastRequest(t *testing.T) {
	assert := assert.New(t)

	service := mockSQSClient{
		messages: []*sqs.Message{
			{
				Body:          aws.String("this is message #0"),
				ReceiptHandle: aws.String("receipt-handle-0"),
				MessageId:     aws.String("message-id-0"),
			},
			{
				Body:          aws.String("this is message #1"),
				ReceiptHandle: aws.String("receipt-handle-1"),
				MessageId:     aws.String("message-id-1"),
			},
		},
	}

	q := queue{
		url:     "test-url",
		service: &service,
		mutex:   &sync.Mutex{},
		logger:  logrus.StandardLogger(),
		metrics: metrics.Default,
	}

	// No previous request
	assert.Nil(q.LastRequest())

	c, _ := q.Listen()

	// blocks until a message is sent to the channel
	<-c
	assert.NotNil(q.LastRequest())

	<-c
	assert.NotNil(q.LastRequest())
}

// Mock implementation of SQS used for these tests
type mockSQSClient struct {
	sqsiface.SQSAPI
	messages        []*sqs.Message
	index           int
	deletedMessages []*sqs.Message
	mutex           *sync.Mutex
}

func (c *mockSQSClient) ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if c.index == len(c.messages) {
		return nil, errors.New("could not receive more messages")
	}
	m := c.messages[c.index]
	c.index++
	return &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{m},
	}, nil
}

func (c *mockSQSClient) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	var msg *sqs.Message
	for _, m := range c.messages {
		if *m.ReceiptHandle == *input.ReceiptHandle {
			msg = m
			break
		}
	}

	if msg == nil {
		return nil, fmt.Errorf("message not found: %s", *input.ReceiptHandle)
	}

	c.mutex.Lock()
	c.deletedMessages = append(c.deletedMessages, msg)
	c.mutex.Unlock()

	return &sqs.DeleteMessageOutput{}, nil
}
