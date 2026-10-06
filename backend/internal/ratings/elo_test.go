package ratings

import "testing"

func TestExpectedScore(t *testing.T) {
	tests := []struct {
		player, opponent int
		want             float64
	}{
		{player: 1200, opponent: 1200, want: 0.5},
		{player: 1600, opponent: 1200, want: 10.0 / 11.0},
		{player: 1200, opponent: 1600, want: 1.0 / 11.0},
	}
	for _, test := range tests {
		if got := ExpectedScore(test.player, test.opponent); got-test.want > 1e-9 || test.want-got > 1e-9 {
			t.Errorf("ExpectedScore(%d, %d) = %.10f, want %.10f", test.player, test.opponent, got, test.want)
		}
	}
}

func TestCalculateElo(t *testing.T) {
	winner, err := CalculateElo(1200, 1200, 1)
	if err != nil || winner != 1216 {
		t.Fatalf("winner rating = %d, %v; want 1216", winner, err)
	}
	loser, err := CalculateElo(1200, 1200, 0)
	if err != nil || loser != 1184 {
		t.Fatalf("loser rating = %d, %v; want 1184", loser, err)
	}
	draw, err := CalculateElo(1200, 1200, 0.5)
	if err != nil || draw != 1200 {
		t.Fatalf("draw rating = %d, %v; want 1200", draw, err)
	}
	if _, err := CalculateElo(1200, 1200, 1.1); err == nil {
		t.Fatal("invalid score accepted")
	}
}

func TestScoresAndModes(t *testing.T) {
	for _, mode := range []Mode{ModeBullet, ModeBlitz, ModeRapid, ModePuzzle} {
		if !mode.Valid() {
			t.Errorf("mode %q not accepted", mode)
		}
	}
	if _, err := ScoreForResult("pending", true); err == nil {
		t.Fatal("ongoing game result accepted")
	}
}
