package matchmaking

import "time"

const (
	initialRatingRange = 100
	rangeExpansionStep = 50
	rangeExpansionTime = 10 * time.Second
	maximumRatingRange = 1000
)

// RatingRangeForWait increases the acceptable rating gap as a player waits.
func RatingRangeForWait(waited time.Duration) int {
	if waited < 0 {
		waited = 0
	}
	steps := int(waited / rangeExpansionTime)
	value := initialRatingRange + steps*rangeExpansionStep
	if value > maximumRatingRange {
		return maximumRatingRange
	}
	return value
}

func ratingsCompatible(firstRating int, firstWait time.Duration, secondRating int, secondWait time.Duration) bool {
	limit := RatingRangeForWait(firstWait)
	if secondRange := RatingRangeForWait(secondWait); secondRange > limit {
		limit = secondRange
	}
	difference := firstRating - secondRating
	if difference < 0 {
		difference = -difference
	}
	return difference <= limit
}
