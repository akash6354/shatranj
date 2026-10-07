package puzzles

import "github.com/akash6354/shatranj/backend/internal/ratings"

const defaultPuzzleRating = 1200

func updatedPuzzleRatings(userRating, puzzleRating int, correct bool) (int, int, error) {
	score := 0.0
	if correct {
		score = 1
	}
	userAfter, err := ratings.CalculateEloWithK(userRating, puzzleRating, score, 24)
	if err != nil {
		return 0, 0, err
	}
	puzzleAfter, err := ratings.CalculateEloWithK(puzzleRating, userRating, 1-score, 16)
	if err != nil {
		return 0, 0, err
	}
	return userAfter, puzzleAfter, nil
}
