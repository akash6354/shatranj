package review

import (
	"math"

	"github.com/shatranj/backend/internal/analysis"
)

func AccuracyFromLoss(centipawnLoss int) float64 {
	return analysis.AccuracyFromLoss(centipawnLoss)
}

func PlayerAccuracy(moves []analysis.Move, color analysis.Color) float64 {
	var sum float64
	count := 0
	for _, move := range moves {
		if move.Color == color {
			sum += move.Accuracy
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Round(sum/float64(count)*100) / 100
}
