//go:generate mockgen -source=worker.go -destination=mocks.go -package=worker
package worker

import "context"

type Consumer interface {
	Start(ctx context.Context) error
}

type Worker struct {
	consumer Consumer
}

func New(consumer Consumer) *Worker {
	return &Worker{consumer: consumer}
}

func (w *Worker) Run(ctx context.Context) error {
	return w.consumer.Start(ctx)
}
