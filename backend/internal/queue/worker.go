package queue

import (
	"context"
	"fmt"
	"log/slog"
)

// Handler processes one job.
type Handler func(context.Context, Job) error

// Worker dispatches jobs to their registered handlers until its context ends.
type Worker struct {
	queue    *Queue
	logger   *slog.Logger
	handlers map[string]Handler
}

// NewWorker creates a worker for the supplied queue and handler map.
func NewWorker(queue *Queue, logger *slog.Logger, handlers map[string]Handler) *Worker {
	return &Worker{queue: queue, logger: logger, handlers: handlers}
}

// Run consumes jobs until context cancellation or queue closure.
func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case job, ok := <-w.queue.Jobs():
			if !ok {
				return nil
			}
			handler, exists := w.handlers[job.Type]
			if !exists {
				w.logger.ErrorContext(ctx, "no handler registered for job", "job_id", job.ID, "job_type", job.Type)
				continue
			}
			if err := handler(ctx, job); err != nil {
				w.logger.ErrorContext(ctx, "job handler failed", "job_id", job.ID, "job_type", job.Type, "error", err)
			}
		}
	}
}

// Validate checks that a job has the fields required for dispatch.
func (job Job) Validate() error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	if job.Type == "" {
		return fmt.Errorf("job type is required")
	}
	return nil
}
