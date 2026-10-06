package queue

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestWorkerStartsAndStopsOnContextCancellation(t *testing.T) {
	jobs := New(1)
	worker := NewWorker(jobs, slog.New(slog.NewTextHandler(io.Discard, nil)), map[string]Handler{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}
