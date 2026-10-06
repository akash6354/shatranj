package matchmaking

import (
	"testing"
	"time"
)

func TestRatingRangeExpansion(t *testing.T) {
	tests := []struct {
		wait time.Duration
		want int
	}{
		{wait: 0, want: 100},
		{wait: 9 * time.Second, want: 100},
		{wait: 10 * time.Second, want: 150},
		{wait: 30 * time.Second, want: 250},
		{wait: time.Hour, want: maximumRatingRange},
	}
	for _, test := range tests {
		if got := RatingRangeForWait(test.wait); got != test.want {
			t.Errorf("RatingRangeForWait(%s) = %d, want %d", test.wait, got, test.want)
		}
	}
}

func TestRatingsCompatible(t *testing.T) {
	if !ratingsCompatible(1200, 0, 1300, 0) {
		t.Fatal("initial-range boundary should match")
	}
	if ratingsCompatible(1200, 0, 1301, 0) {
		t.Fatal("rating gap outside initial range should not match")
	}
	if !ratingsCompatible(1200, 0, 1500, time.Minute) {
		t.Fatal("long-wait range should expand to include distant rating")
	}
}
