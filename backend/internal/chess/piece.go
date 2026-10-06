package chess

import (
	"fmt"
	"strings"
)

// Color identifies which side owns a piece.
type Color uint8

const (
	NoColor Color = iota
	White
	Black
)

func (c Color) Opposite() Color {
	if c == White {
		return Black
	}
	if c == Black {
		return White
	}
	return NoColor
}

// PieceType identifies the kind of a chess piece.
type PieceType uint8

const (
	NoPiece PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Piece is a color and kind pair. A zero Piece represents an empty square.
type Piece struct {
	Color Color
	Type  PieceType
}

func (p Piece) IsEmpty() bool {
	return p.Type == NoPiece
}

func (p Piece) FEN() (byte, error) {
	var symbol byte
	switch p.Type {
	case Pawn:
		symbol = 'p'
	case Knight:
		symbol = 'n'
	case Bishop:
		symbol = 'b'
	case Rook:
		symbol = 'r'
	case Queen:
		symbol = 'q'
	case King:
		symbol = 'k'
	default:
		return 0, fmt.Errorf("invalid piece type %d", p.Type)
	}
	switch p.Color {
	case White:
		symbol -= 'a' - 'A'
	case Black:
	default:
		return 0, fmt.Errorf("invalid piece color %d", p.Color)
	}
	return symbol, nil
}

func pieceFromFEN(symbol byte) (Piece, error) {
	color := Black
	if symbol >= 'A' && symbol <= 'Z' {
		color = White
		symbol += 'a' - 'A'
	}
	var kind PieceType
	switch symbol {
	case 'p':
		kind = Pawn
	case 'n':
		kind = Knight
	case 'b':
		kind = Bishop
	case 'r':
		kind = Rook
	case 'q':
		kind = Queen
	case 'k':
		kind = King
	default:
		return Piece{}, fmt.Errorf("invalid FEN piece %q", symbol)
	}
	return Piece{Color: color, Type: kind}, nil
}

func (p Piece) String() string {
	if p.IsEmpty() {
		return "."
	}
	symbol, err := p.FEN()
	if err != nil {
		return "?"
	}
	return strings.ToUpper(string(symbol))
}
