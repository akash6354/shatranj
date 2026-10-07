package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shatranj/backend/internal/analysis"
	"github.com/shatranj/backend/internal/cache"
	"github.com/shatranj/backend/internal/config"
	"github.com/shatranj/backend/internal/database"
	"github.com/shatranj/backend/internal/observability"
	"github.com/shatranj/backend/internal/queue"
	"github.com/shatranj/backend/internal/review"
)

func main() {
	logger := observability.NewLogger()
	if err := run(logger); err != nil {
		logger.Error("worker stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStartup()
	db, err := database.OpenPostgres(startupCtx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				logger.Error("close PostgreSQL connection pool", "error", err)
			}
		}()
		logger.Info("PostgreSQL connected")
	}
	cacheClient, err := cache.Open(startupCtx, cfg.RedisURL)
	if err != nil {
		logger.Warn("Redis unavailable, using in-memory cache fallback", "error", err)
	}
	defer func() {
		if err := cacheClient.Close(); err != nil {
			logger.Error("close cache client", "error", err)
		}
	}()
	logger.Info(cacheClient.Describe())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("worker process started")
	if db != nil {
		evaluator := analysis.NewStockfish(os.Getenv("STOCKFISH_PATH"), 15*time.Second, 14)
		worker := analysis.NewWorker(review.NewRepository(db), evaluator, logger)
		logger.Info("game analysis worker configured", "stockfish_available", evaluator.Available())
		return worker.Run(ctx, time.Second)
	}
	jobs := queue.New(100)
	worker := queue.NewWorker(jobs, logger, map[string]queue.Handler{})
	return worker.Run(ctx)
}
