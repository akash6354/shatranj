package review

import "github.com/shatranj/backend/internal/analysis"
import "github.com/shatranj/backend/internal/chess"

type MoveComparison struct {
	PlayedMove     string `json:"played_move"`
	BestMove       string `json:"best_move"`
	BestMovePlayed bool   `json:"best_move_played"`
	CentipawnLoss  int    `json:"centipawn_loss"`
}

func CompareMoves(beforeScore, afterScore int, color analysis.Color, playedMove, bestMove string) MoveComparison {
	loss, bestPlayed := analysis.Compare(
		beforeScore, afterScore, colorToChess(color), playedMove, bestMove,
	)
	return MoveComparison{
		PlayedMove: playedMove, BestMove: bestMove,
		BestMovePlayed: bestPlayed, CentipawnLoss: loss,
	}
}

func colorToChess(color analysis.Color) chess.Color {
	if color == analysis.Black {
		return chess.Black
	}
	return chess.White
}
