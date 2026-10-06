package chess

import (
	"fmt"
	"strings"
)

// Square uses the conventional mapping a1=0 through h8=63.
type Square uint8

const NoSquare Square = 64

func NewSquare(file, rank int) Square {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare
	}
	return Square(rank*8 + file)
}

func (s Square) Valid() bool {
	return s < 64
}

func (s Square) File() int {
	if !s.Valid() {
		return -1
	}
	return int(s % 8)
}

func (s Square) Rank() int {
	if !s.Valid() {
		return -1
	}
	return int(s / 8)
}

func (s Square) String() string {
	if !s.Valid() {
		return "-"
	}
	return string([]byte{byte('a' + s.File()), byte('1' + s.Rank())})
}

func ParseSquare(text string) (Square, error) {
	if len(text) != 2 || text[0] < 'a' || text[0] > 'h' || text[1] < '1' || text[1] > '8' {
		return NoSquare, fmt.Errorf("invalid square %q", text)
	}
	return NewSquare(int(text[0]-'a'), int(text[1]-'1')), nil
}

// Move describes a move. Promotion must be Queen, Rook, Bishop, or Knight.
type Move struct {
	From      Square
	To        Square
	Promotion PieceType
}

func (m Move) String() string {
	text := m.From.String() + m.To.String()
	switch m.Promotion {
	case Queen:
		text += "q"
	case Rook:
		text += "r"
	case Bishop:
		text += "b"
	case Knight:
		text += "n"
	}
	return text
}

func ParseMove(text string) (Move, error) {
	text = strings.TrimSpace(text)
	if len(text) != 4 && len(text) != 5 {
		return Move{}, fmt.Errorf("invalid coordinate move %q", text)
	}
	from, err := ParseSquare(text[:2])
	if err != nil {
		return Move{}, err
	}
	to, err := ParseSquare(text[2:4])
	if err != nil {
		return Move{}, err
	}
	move := Move{From: from, To: to}
	if len(text) == 5 {
		switch text[4] {
		case 'q', 'Q':
			move.Promotion = Queen
		case 'r', 'R':
			move.Promotion = Rook
		case 'b', 'B':
			move.Promotion = Bishop
		case 'n', 'N':
			move.Promotion = Knight
		default:
			return Move{}, fmt.Errorf("invalid promotion in move %q", text)
		}
	}
	return move, nil
}
