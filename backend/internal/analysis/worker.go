package analysis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/akash6354/shatranj/backend/internal/chess"
)

const defaultPollInterval = time.Second

type Worker struct {
	store     Store
	evaluator Evaluator
	logger    *slog.Logger
}

func NewWorker(store Store, evaluator Evaluator, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{store: store, evaluator: evaluator, logger: logger}
}

func (w *Worker) Run(ctx context.Context, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}
	for {
		task, err := w.store.Claim(ctx)
		if errors.Is(err, ErrNoJob) {
			timer := time.NewTimer(pollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil
			case <-timer.C:
				continue
			}
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.ErrorContext(ctx, "claim analysis job", "error", err)
			if err := waitContext(ctx, pollInterval); err != nil {
				return nil
			}
			continue
		}
		if err := w.process(ctx, task); err != nil {
			if errors.Is(err, ErrUnavailable) {
				if markErr := w.store.MarkUnavailable(ctx, task.ReviewID); markErr != nil {
					w.logger.ErrorContext(ctx, "mark analysis unavailable", "review_id", task.ReviewID, "error", markErr)
				}
				continue
			}
			w.logger.ErrorContext(ctx, "process analysis job", "review_id", task.ReviewID, "error", err)
			if failErr := w.store.Fail(ctx, task.ReviewID, err); failErr != nil {
				w.logger.ErrorContext(ctx, "record analysis failure", "review_id", task.ReviewID, "error", failErr)
			}
		}
	}
}

func (w *Worker) process(ctx context.Context, task Task) error {
	if w.evaluator == nil {
		return ErrUnavailable
	}
	positions := make([]chess.Position, len(task.Moves)+1)
	positions[0] = chess.StartingPosition()
	for index, move := range task.Moves {
		if move.PlayerID == "" || move.UCI == "" {
			return fmt.Errorf("invalid stored move at ply %d", index+1)
		}
		expectedPlayer := task.WhiteID
		if positions[index].SideToMove == chess.Black {
			expectedPlayer = task.BlackID
		}
		if move.PlayerID != expectedPlayer {
			return fmt.Errorf("player color mismatch at ply %d", index+1)
		}
		parsed, err := chess.ParseMove(move.UCI)
		if err != nil {
			return fmt.Errorf("parse game move at ply %d: %w", index+1, err)
		}
		positions[index+1], err = positions[index].MakeMove(parsed)
		if err != nil {
			return fmt.Errorf("validate game move at ply %d: %w", index+1, err)
		}
		if positions[index+1].FEN() != move.FENAfter {
			return fmt.Errorf("stored position mismatch after ply %d", index+1)
		}
	}

	fens := make([]string, len(positions))
	for index, position := range positions {
		fens[index] = position.FEN()
	}
	var evaluations []Evaluation
	var err error
	if batch, ok := w.evaluator.(BatchEvaluator); ok {
		evaluations, err = batch.EvaluateMany(ctx, fens)
		if err != nil {
			return fmt.Errorf("evaluate game positions: %w", err)
		}
		if len(evaluations) != len(positions) {
			return fmt.Errorf("engine returned %d evaluations for %d positions", len(evaluations), len(positions))
		}
	} else {
		evaluations = make([]Evaluation, len(positions))
		for index, fen := range fens {
			evaluation, err := w.evaluator.Evaluate(ctx, fen)
			if err != nil {
				return fmt.Errorf("evaluate position after ply %d: %w", index, err)
			}
			evaluations[index] = evaluation
		}
	}
	result := Result{GameID: task.GameID, Moves: make([]Move, 0, len(task.Moves))}
	var whiteAccuracy, blackAccuracy float64
	var whiteCount, blackCount int
	for index, stored := range task.Moves {
		before := evaluations[index]
		after := evaluations[index+1]
		mover := positions[index].SideToMove
		loss, bestPlayed := Compare(before.ScoreCP, after.ScoreCP, mover, stored.UCI, before.BestMove)
		improvement := after.ScoreCP - before.ScoreCP
		if mover == chess.Black {
			improvement = -improvement
		}
		color := White
		if mover == chess.Black {
			color = Black
		}
		plyAccuracy := moveAccuracy(loss)
		moveResult := Move{
			Ply: index + 1, MoveNumber: (index / 2) + 1, Color: color, PlayerID: stored.PlayerID,
			UCI: stored.UCI, SAN: stored.SAN, ScoreCP: after.ScoreCP, BestMove: before.BestMove,
			CentipawnLoss:  loss,
			Classification: string(classifyMove(loss, bestPlayed, len(positions[index].LegalMoves()) == 1, improvement)),
			Accuracy:       plyAccuracy,
		}
		result.Moves = append(result.Moves, moveResult)
		if mover == chess.White {
			whiteAccuracy += plyAccuracy
			whiteCount++
		} else {
			blackAccuracy += plyAccuracy
			blackCount++
		}
	}
	if whiteCount > 0 {
		result.WhiteAccuracy = roundAccuracy(whiteAccuracy / float64(whiteCount))
	}
	if blackCount > 0 {
		result.BlackAccuracy = roundAccuracy(blackAccuracy / float64(blackCount))
	}
	if len(result.Moves) > 0 {
		result.OverallAccuracy = roundAccuracy((whiteAccuracy + blackAccuracy) / float64(len(result.Moves)))
	}
	if err := w.store.Complete(ctx, result); err != nil {
		return fmt.Errorf("save analysis result: %w", err)
	}
	return nil
}

func waitContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func roundAccuracy(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
