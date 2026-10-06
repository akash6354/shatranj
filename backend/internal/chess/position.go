package chess

import "fmt"

// CastlingRights stores the four FEN castling-right flags.
type CastlingRights uint8

const (
	WhiteKingSide CastlingRights = 1 << iota
	WhiteQueenSide
	BlackKingSide
	BlackQueenSide
)

// Position is a complete chess position. Its value semantics make copies safe
// to modify independently.
type Position struct {
	Board          [64]Piece
	SideToMove     Color
	Castling       CastlingRights
	EnPassant      Square
	HalfmoveClock  int
	FullmoveNumber int
}

// StartingPosition returns the standard initial chess position.
func StartingPosition() Position {
	position, _ := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return position
}

func (p Position) PieceAt(square Square) Piece {
	if !square.Valid() {
		return Piece{}
	}
	return p.Board[square]
}

func (p *Position) SetPiece(square Square, piece Piece) error {
	if !square.Valid() {
		return fmt.Errorf("invalid square %s", square)
	}
	if piece.IsEmpty() {
		p.Board[square] = Piece{}
		return nil
	}
	if (piece.Color != White && piece.Color != Black) || piece.Type < Pawn || piece.Type > King {
		return fmt.Errorf("invalid piece %+v", piece)
	}
	p.Board[square] = piece
	return nil
}

func (p Position) KingSquare(color Color) (Square, bool) {
	for square, piece := range p.Board {
		if piece.Type == King && piece.Color == color {
			return Square(square), true
		}
	}
	return NoSquare, false
}

func (p Position) Validate() error {
	if p.SideToMove != White && p.SideToMove != Black {
		return fmt.Errorf("invalid side to move %d", p.SideToMove)
	}
	if p.FullmoveNumber < 1 {
		return fmt.Errorf("fullmove number must be positive")
	}
	if p.HalfmoveClock < 0 {
		return fmt.Errorf("halfmove clock cannot be negative")
	}
	if p.Castling&^CastlingRights(15) != 0 {
		return fmt.Errorf("invalid castling rights")
	}
	kings := [2]int{}
	for _, piece := range p.Board {
		if piece.IsEmpty() {
			continue
		}
		if (piece.Color != White && piece.Color != Black) || piece.Type < Pawn || piece.Type > King {
			return fmt.Errorf("invalid piece %+v", piece)
		}
		if piece.Type == King {
			kings[piece.Color-1]++
		}
	}
	if kings[0] != 1 || kings[1] != 1 {
		return fmt.Errorf("position must contain exactly one king per side")
	}
	if p.EnPassant.Valid() && (p.EnPassant.Rank() != 2 && p.EnPassant.Rank() != 5 || !p.PieceAt(p.EnPassant).IsEmpty()) {
		return fmt.Errorf("invalid en passant square %s", p.EnPassant)
	}
	return nil
}

func (p Position) IsDrawByFiftyMoveRule() bool {
	return p.HalfmoveClock >= 100
}

func (p Position) IsInsufficientMaterial() bool {
	minorCount := 0
	bishopColor := -1
	bishopsOnSameColor := true
	for square, piece := range p.Board {
		switch piece.Type {
		case Pawn, Rook, Queen:
			return false
		case Knight:
			minorCount++
		case Bishop:
			minorCount++
			color := (square%8 + square/8) % 2
			if bishopColor == -1 {
				bishopColor = color
			} else if bishopColor != color {
				bishopsOnSameColor = false
			}
		}
	}
	if minorCount <= 1 {
		return true
	}
	if bishopsOnSameColor {
		for _, piece := range p.Board {
			if piece.Type == Knight {
				return false
			}
		}
		return true
	}
	return false
}

func (p Position) IsDraw() bool {
	return p.IsDrawByFiftyMoveRule() || p.IsInsufficientMaterial() || p.IsStalemate()
}
