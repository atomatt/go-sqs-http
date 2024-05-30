package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	logFormat := flag.String("log-format", "json", "log format")
	flag.Parse()

	configureLogging(*logFormat)

	httpServer := http.Server{
		Addr:    ":10000",
		Handler: newHandler(),
	}

	sqsLoop := NewSqsLoop(
		session.Must(session.NewSession(&aws.Config{
			Endpoint: aws.String("http://localstack:4566"),
		})),
		"http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/outbox",
		func(msg *sqs.Message) error {
			slog.Info(*msg.Body)
			return nil
		},
	)

	slog.Info("Starting HTTP server", "addr", httpServer.Addr)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error(err.Error())
			}
		}
	}()

	slog.Info("Starting SQS loop")
	go func() {
		sqsLoop.Run()
	}()

	// Catch signals for graceful shutdown
	interrupt, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-interrupt.Done()

	// Limit total shutdown time to ensure nothing hangs.
	shutdown, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	slog.Info("Stopping HTTP server ...")
	if err := httpServer.Shutdown(shutdown); err != nil {
		slog.Warn(err.Error())
	}

	slog.Info("Stopping SQS loop ...")
	if err := sqsLoop.Shutdown(shutdown); err != nil {
		slog.Warn(err.Error())
	}
}

func configureLogging(format string) {
	switch format {
	case "json":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	case "text":
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	case "default":
	default:
		slog.Warn("Ignoring log format", "format", format)
	}
}
