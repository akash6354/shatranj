package puzzles

import "time"

func dailyDate(now time.Time) string {
	return now.UTC().Format("2006-01-02")
}
