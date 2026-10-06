package analysis

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStockfishUnavailableAndInvalidFEN(t *testing.T) {
	engine := NewStockfish("", time.Second, 1)
	if engine.Available() {
		t.Fatal("unconfigured Stockfish reports available")
	}
	if _, err := engine.Evaluate(context.Background(), ""); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unconfigured engine error = %v", err)
	}
	engine = NewStockfish("stockfish", time.Second, 1)
	if _, err := engine.Evaluate(context.Background(), "not a fen"); err == nil {
		t.Fatal("invalid FEN accepted")
	}
}

func TestParseStockfishScores(t *testing.T) {
	if score, ok := parseCentipawnScore("info depth 12 score cp -34 nodes 200"); !ok || score != -34 {
		t.Fatalf("centipawn score = %d, %v", score, ok)
	}
	if score, ok := parseMateScore("info depth 12 score mate -3"); !ok || score >= 0 {
		t.Fatalf("negative mate score = %d, %v", score, ok)
	}
}
