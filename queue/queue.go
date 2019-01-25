package queue

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/fchoquet/golibs/metrics"
	log "github.com/sirupsen/logrus"
)

// default metrics values. Feel free to override in your project
var (
	QueueMessageReceived    = "queue.message_received"
	QueueError              = "queue.error"
	QueueReceiveMessageTime = "queue.receive_message.time"
	QueueAckTried           = "queue.ack.tried"
	QueueAckOk              = "queue.ack.ok"
	QueueAckErr             = "queue.ack.error"
	QueueAckTime            = "queue.ack.time"
)

// Message reprensents a queue message
type Message struct {
	MessageID     string `json:"message_id"`
	Body          string `json:"body"`
	ReceiptHandle string `json:"receipt_handle"`
	ack           chan<- *Message
}

// Ack acknowledges the message
func (m *Message) Ack() {
	m.ack <- m
}

// Listener listens for messages in the queue and sends them to a channel
// Errors are sent to a separate channel
type Listener interface {
	// Listen starts listening for messages
	Listen() (<-chan *Message, <-chan error)
	// LastRequest returns the duration since the last request to the server
	LastRequest() *time.Duration
}

// New creates a new default Listener implementation
func New(url string, logger log.FieldLogger, metrics metrics.Client) (Listener, error) {
	session, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, err
	}

	return &queue{
		url:     url,
		service: sqs.New(session),
		mutex:   &sync.Mutex{},
		logger:  logger,
		metrics: metrics,
	}, nil
}

type queue struct {
	url         string
	service     sqsiface.SQSAPI
	lastRequest *time.Time
	mutex       *sync.Mutex
	logger      log.FieldLogger
	metrics     metrics.Client
}

func (q *queue) Listen() (<-chan *Message, <-chan error) {
	c := make(chan *Message)
	ack := make(chan *Message)
	e := make(chan error)

	// listen to queue messages and pushes them to c. Errors are pushed to e
	go listen(q, c, e, ack)

	// listen to acknowledgement messages and processes them
	go listenAck(q, ack)

	return c, e
}

func (q *queue) LastRequest() *time.Duration {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.lastRequest == nil {
		return nil
	}
	duration := time.Since(*q.lastRequest)
	return &duration
}

func (q *queue) updateLastRequest() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	now := time.Now()
	q.lastRequest = &now
}

func listen(q *queue, c chan *Message, e chan error, ack chan *Message) {
	for {
		start := time.Now()

		output, err := q.service.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(q.url),
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(20),
		})
		q.metrics.Incr(QueueMessageReceived)
		q.metrics.Timing(QueueReceiveMessageTime, start)
		if err != nil {
			q.logger.WithError(err).Error("Could not receive message")
			metrics.Incr(QueueError)
			e <- err
			continue
		}
		// The service is doing its job, so let's say it
		q.updateLastRequest()

		for _, msg := range output.Messages {
			q.logger.WithField("body", *msg.Body).Debug("Message body")

			c <- &Message{
				MessageID:     *msg.MessageId,
				Body:          *msg.Body,
				ReceiptHandle: *msg.ReceiptHandle,
				ack:           ack,
			}
		}

		if len(output.Messages) == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

func listenAck(q *queue, ack <-chan *Message) {
	for msg := range ack {
		q.metrics.Incr(QueueAckTried)
		start := time.Now()

		_, err := q.service.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &q.url,
			ReceiptHandle: &msg.ReceiptHandle,
		})

		if err != nil {
			q.logger.WithError(err).Error("Could not delete message")
			q.metrics.Incr(QueueAckErr)
			// There's not much we can do here. Message is already processed and we can't rollback
			// We'll get a duplicate
			// This is unlikely to happen so let's only monitor it for now and see if an action is needed
		}

		q.metrics.Incr(QueueAckOk)
		q.metrics.Timing(QueueAckTime, start)
	}
}
