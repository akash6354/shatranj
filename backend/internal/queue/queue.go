// Package queue provides the background-job envelope and a small, process-local
// queue used as the worker scaffold when PostgreSQL is not configured.
//
// This queue is deliberately NOT distributed: jobs live in a Go channel, are
// lost when the process restarts, and are invisible to other worker processes.
// When PostgreSQL is present the worker path is database-backed instead (see
// internal/analysis), where jobs are claimed atomically. Redis-backed job
// coordination is future work; it would publish jobs to a shared stream and
// must not replace the durable database claim that already exists.
package queue

import (
	"context"
)

// Queue is a small in-process queue suitable for the initial worker scaffold.
type Queue struct {
	jobs chan Job
}

// New creates a queue with the requested buffer capacity.
func New(capacity int) *Queue {
	if capacity < 0 {
		capacity = 0
	}
	return &Queue{jobs: make(chan Job, capacity)}
}

// Enqueue adds a job unless the caller's context is canceled.
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	if err := job.Validate(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.jobs <- job:
		return nil
	}
}

// Jobs returns the receive-only job stream consumed by a worker.
func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}
