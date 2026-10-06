package ratings

import (
	"fmt"
	"math"
)

const (
	DefaultRating = 1200
	DefaultK      = 32
)

// ExpectedScore calculates a player's expected score against an opponent.
func ExpectedScore(playerRating, opponentRating int) float64 {
	return 1 / (1 + math.Pow(10, float64(opponentRating-playerRating)/400))
}

// CalculateElo returns the new rating using the standard Elo update formula.
func CalculateElo(playerRating, opponentRating int, score float64) (int, error) {
	return CalculateEloWithK(playerRating, opponentRating, score, DefaultK)
}

func CalculateEloWithK(playerRating, opponentRating int, score float64, k int) (int, error) {
	if playerRating < 0 || opponentRating < 0 || k < 1 {
		return 0, fmt.Errorf("ratings and K factor must be non-negative and K must be positive")
	}
	if score < 0 || score > 1 || math.IsNaN(score) {
		return 0, fmt.Errorf("score must be between 0 and 1")
	}
	rating := int(math.Round(float64(playerRating) + float64(k)*(score-ExpectedScore(playerRating, opponentRating))))
	if rating < 0 {
		rating = 0
	}
	return rating, nil
}

func ScoreForResult(result string, white bool) (float64, error) {
	switch result {
	case "1-0":
		if white {
			return 1, nil
		}
		return 0, nil
	case "0-1":
		if white {
			return 0, nil
		}
		return 1, nil
	case "1/2-1/2":
		return 0.5, nil
	default:
		return 0, ErrInvalidResult
	}
}

func kFactor(mode Mode, games int) int {
	if mode == ModePuzzle {
		return 24
	}
	if games < 30 {
		return 40
	}
	return DefaultK
}
