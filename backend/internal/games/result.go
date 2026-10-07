package games

import "github.com/shatranj/backend/internal/chess"

func determineResult(position chess.Position) (Status, Result, string) {
	if position.IsCheckmate() {
		winner := position.SideToMove.Opposite()
		if winner == chess.White {
			return StatusFinished, ResultWhiteWin, "checkmate"
		}
		return StatusFinished, ResultBlackWin, "checkmate"
	}
	if position.IsStalemate() {
		return StatusFinished, ResultDraw, "stalemate"
	}
	if position.IsInsufficientMaterial() {
		return StatusFinished, ResultDraw, "insufficient_material"
	}
	if position.IsDrawBySeventyFiveMoveRule() {
		return StatusFinished, ResultDraw, "seventyfive_move_rule"
	}
	return StatusActive, ResultOngoing, ""
}

// ResultOnTimeout calculates the result when a clock service confirms that the
// given side has flagged. A flag fall is a draw when neither side can mate.
func ResultOnTimeout(position chess.Position, flaggedColor chess.Color) (Result, string) {
	if flaggedColor != chess.White && flaggedColor != chess.Black {
		return ResultOngoing, "invalid_flagged_color"
	}
	if position.IsInsufficientMaterial() {
		return ResultDraw, "timeout_insufficient_material"
	}
	if flaggedColor == chess.White {
		return ResultBlackWin, "timeout"
	}
	return ResultWhiteWin, "timeout"
}
