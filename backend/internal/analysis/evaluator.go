package analysis

import (
	"math"

	"github.com/akash6354/shatranj/backend/internal/chess"
)

func centipawnLoss(before, after int, mover chess.Color) int {
	loss := before - after
	if mover == chess.Black {
		loss = after - before
	}
	if loss < 0 {
		return 0
	}
	return loss
}

func AccuracyFromLoss(loss int) float64 {
	if loss < 0 {
		loss = 0
	}
	value := 100 * math.Exp(-float64(loss)/350)
	if value < 0 {
		return 0
	}
	return math.Round(value*100) / 100
}

func moveAccuracy(loss int) float64 {
	return AccuracyFromLoss(loss)
}

func classifyMove(loss int, bestMovePlayed, onlyLegalMove bool, improvement int) Classification {
	if bestMovePlayed && onlyLegalMove && improvement >= 150 {
		return Brilliant
	}
	switch {
	case loss <= 5:
		return Best
	case loss <= 20:
		return Great
	case loss <= 50:
		return Good
	case loss <= 100:
		return Inaccuracy
	case loss <= 200:
		return Mistake
	default:
		return Blunder
	}
}

func Compare(before, after int, mover chess.Color, actualMove, bestMove string) (int, bool) {
	return centipawnLoss(before, after, mover), actualMove != "" && actualMove == bestMove
}
