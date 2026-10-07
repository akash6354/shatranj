package analysis

import (
	"context"
	"errors"

	"github.com/akash6354/shatranj/backend/internal/chess"
)

var (
	ErrUnavailable = errors.New("chess engine is unavailable")
	ErrNoJob       = errors.New("no analysis job available")
)

type Evaluation struct {
	ScoreCP  int    `json:"score_cp"`
	BestMove string `json:"best_move"`
}

type Evaluator interface {
	Evaluate(context.Context, string) (Evaluation, error)
}

type BatchEvaluator interface {
	EvaluateMany(context.Context, []string) ([]Evaluation, error)
}

type Color string

const (
	White Color = "white"
	Black Color = "black"
)

type Move struct {
	Ply            int     `json:"ply"`
	MoveNumber     int     `json:"move_number"`
	Color          Color   `json:"color"`
	PlayerID       string  `json:"player_id"`
	UCI            string  `json:"uci"`
	SAN            string  `json:"san"`
	ScoreCP        int     `json:"score_cp"`
	BestMove       string  `json:"best_move"`
	CentipawnLoss  int     `json:"centipawn_loss"`
	Classification string  `json:"classification"`
	Accuracy       float64 `json:"accuracy"`
}

type Result struct {
	GameID          string  `json:"game_id"`
	WhiteAccuracy   float64 `json:"white_accuracy"`
	BlackAccuracy   float64 `json:"black_accuracy"`
	OverallAccuracy float64 `json:"overall_accuracy"`
	Moves           []Move  `json:"moves"`
}

type PlayedMove struct {
	PlayerID string
	UCI      string
	SAN      string
	FENAfter string
}

type Task struct {
	ReviewID string
	GameID   string
	WhiteID  string
	BlackID  string
	Moves    []PlayedMove
}

type Store interface {
	Claim(context.Context) (Task, error)
	Complete(context.Context, Result) error
	Fail(context.Context, string, error) error
	MarkUnavailable(context.Context, string) error
}

type Classification string

const (
	Brilliant  Classification = "brilliant"
	Great      Classification = "great"
	Best       Classification = "best"
	Good       Classification = "good"
	Inaccuracy Classification = "inaccuracy"
	Mistake    Classification = "mistake"
	Blunder    Classification = "blunder"
)

type movePosition struct {
	position chess.Position
	move     PlayedMove
}
