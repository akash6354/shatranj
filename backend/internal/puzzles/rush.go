package puzzles

import "time"

func validRushDuration(seconds int) bool {
	return seconds >= 30 && seconds <= int((30*time.Minute).Seconds())
}
