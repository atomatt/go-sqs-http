package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cenkalti/backoff/v4"
)

type SqsHandler func(*sqs.Message) error

// SqsLoop is a SQS message handler loop that retrieves messages from a queue
// and calls a handler for each one. If there is no error, the message is
// removed from the queue.
type SqsLoop struct {
	sqs      *sqs.SQS
	queueUrl string
	handler  SqsHandler

	// Shutdown handling
	stop     context.CancelFunc
	shutdown chan struct{}
}

func NewSqsLoop(awsSession *session.Session, queueUrl string, handler SqsHandler) *SqsLoop {
	return &SqsLoop{
		sqs:      sqs.New(awsSession),
		queueUrl: queueUrl,
		handler:  handler,
	}
}

func (loop *SqsLoop) Run() {
	stopped, stop := context.WithCancel(context.Background())
	defer stop()

	loop.stop = stop
	loop.shutdown = make(chan struct{})
	defer close(loop.shutdown)

	for {
		if err := stopped.Err(); err != nil {
			return
		}

		err := backoff.RetryNotify(loop.poll, backoff.NewExponentialBackOff(), func(err error, d time.Duration) {
			slog.Warn("backoff:", "err", err)
		})
		if err != nil {
			slog.Error("backoff:", "err", err)
		}
	}
}

func (loop *SqsLoop) poll() error {
	slog.Info("Polling SQS queue for messages")

	r, err := loop.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &loop.queueUrl,
		WaitTimeSeconds:     aws.Int64(5),
		MaxNumberOfMessages: aws.Int64(10),
	})
	if err != nil {
		return fmt.Errorf("error polling sqs: %w", err)
	}

	for _, msg := range r.Messages {
		// Maybe: handle panics; treat errors as a permanent failure.
		if err := loop.handler(msg); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}

		if _, err = loop.sqs.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &loop.queueUrl,
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			return fmt.Errorf("error deleting message: %w", err)
		}
	}

	return nil
}

func (loop *SqsLoop) Shutdown(ctx context.Context) error {
	if loop.stop == nil {
		return nil
	}

	loop.stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-loop.shutdown:
		return nil
	}
}
