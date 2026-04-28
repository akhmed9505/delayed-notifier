//go:generate mockgen -source=worker.go -destination=mocks.go -package=worker

// Package worker provides the orchestration logic for running the notification consumption process.
package worker

import "context"

// Consumer defines the contract for a message broker consumer that starts processing notifications.
type Consumer interface {
	Start(ctx context.Context) error
}

// Worker manages the lifecycle of the consumer.
type Worker struct {
	consumer Consumer
}

// New creates a new instance of Worker with the given consumer.
func New(consumer Consumer) *Worker {
	return &Worker{consumer: consumer}
}

// Run starts the consumer and blocks until the context is canceled or an error occurs.
func (w *Worker) Run(ctx context.Context) error {
	return w.consumer.Start(ctx)
}
