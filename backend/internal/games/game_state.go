package games

import (
	"fmt"
	"regexp"

	"github.com/shatranj/backend/internal/chess"
)

var gameIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validGameID(id string) bool {
	return gameIDPattern.MatchString(id)
}

func positionFromGame(game Game) (chess.Position, error) {
	position, err := chess.ParseFEN(game.CurrentFEN)
	if err != nil {
		return chess.Position{}, fmt.Errorf("parse game FEN: %w", err)
	}
	return position, nil
}

func playerColor(game Game, userID string) (chess.Color, error) {
	switch userID {
	case game.WhitePlayerID:
		return chess.White, nil
	case game.BlackPlayerID:
		if game.BlackPlayerID != "" {
			return chess.Black, nil
		}
	}
	return chess.NoColor, ErrForbidden
}

func opponentID(game Game, color chess.Color) string {
	if color == chess.White {
		return game.BlackPlayerID
	}
	return game.WhitePlayerID
}
