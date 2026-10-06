package chess

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseFEN parses a six-field Forsyth-Edwards Notation position.
func ParseFEN(fen string) (Position, error) {
	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return Position{}, fmt.Errorf("FEN must contain six fields")
	}
	position := Position{SideToMove: White, EnPassant: NoSquare, FullmoveNumber: 1}
	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return Position{}, fmt.Errorf("FEN board must contain eight ranks")
	}
	for fenRank, row := range ranks {
		file := 0
		for i := 0; i < len(row); i++ {
			symbol := row[i]
			if symbol >= '1' && symbol <= '8' {
				file += int(symbol - '0')
				continue
			}
			piece, err := pieceFromFEN(symbol)
			if err != nil {
				return Position{}, err
			}
			if file >= 8 {
				return Position{}, fmt.Errorf("too many squares on FEN rank %d", 8-fenRank)
			}
			position.Board[NewSquare(file, 7-fenRank)] = piece
			file++
		}
		if file != 8 {
			return Position{}, fmt.Errorf("FEN rank %d has %d squares, want 8", 8-fenRank, file)
		}
	}
	switch fields[1] {
	case "w":
		position.SideToMove = White
	case "b":
		position.SideToMove = Black
	default:
		return Position{}, fmt.Errorf("invalid FEN active color %q", fields[1])
	}
	if fields[2] != "-" {
		for i := 0; i < len(fields[2]); i++ {
			var right CastlingRights
			switch fields[2][i] {
			case 'K':
				right = WhiteKingSide
			case 'Q':
				right = WhiteQueenSide
			case 'k':
				right = BlackKingSide
			case 'q':
				right = BlackQueenSide
			default:
				return Position{}, fmt.Errorf("invalid FEN castling right %q", fields[2][i])
			}
			if position.Castling&right != 0 {
				return Position{}, fmt.Errorf("duplicate FEN castling right %q", fields[2][i])
			}
			position.Castling |= right
		}
	}
	if fields[3] != "-" {
		square, err := ParseSquare(fields[3])
		if err != nil {
			return Position{}, fmt.Errorf("invalid FEN en passant square: %w", err)
		}
		expectedRank := 5
		if position.SideToMove == Black {
			expectedRank = 2
		}
		if square.Rank() != expectedRank {
			return Position{}, fmt.Errorf("invalid FEN en passant square %s for side to move", square)
		}
		position.EnPassant = square
	}
	halfmove, err := strconv.Atoi(fields[4])
	if err != nil || halfmove < 0 {
		return Position{}, fmt.Errorf("invalid FEN halfmove clock %q", fields[4])
	}
	fullmove, err := strconv.Atoi(fields[5])
	if err != nil || fullmove < 1 {
		return Position{}, fmt.Errorf("invalid FEN fullmove number %q", fields[5])
	}
	position.HalfmoveClock = halfmove
	position.FullmoveNumber = fullmove
	if err := position.Validate(); err != nil {
		return Position{}, fmt.Errorf("invalid FEN position: %w", err)
	}
	return position, nil
}

// FEN serializes a position using the standard six-field notation.
func (p Position) FEN() string {
	var result strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			piece := p.PieceAt(NewSquare(file, rank))
			if piece.IsEmpty() {
				empty++
				continue
			}
			if empty > 0 {
				result.WriteByte(byte('0' + empty))
				empty = 0
			}
			symbol, err := piece.FEN()
			if err != nil {
				result.WriteByte('?')
			} else {
				result.WriteByte(symbol)
			}
		}
		if empty > 0 {
			result.WriteByte(byte('0' + empty))
		}
		if rank > 0 {
			result.WriteByte('/')
		}
	}
	result.WriteByte(' ')
	if p.SideToMove == Black {
		result.WriteByte('b')
	} else {
		result.WriteByte('w')
	}
	result.WriteByte(' ')
	rights := strings.Builder{}
	for _, entry := range []struct {
		right  CastlingRights
		symbol byte
	}{
		{WhiteKingSide, 'K'}, {WhiteQueenSide, 'Q'}, {BlackKingSide, 'k'}, {BlackQueenSide, 'q'},
	} {
		if p.Castling&entry.right != 0 {
			rights.WriteByte(entry.symbol)
		}
	}
	if rights.Len() == 0 {
		result.WriteByte('-')
	} else {
		result.WriteString(rights.String())
	}
	result.WriteByte(' ')
	result.WriteString(p.EnPassant.String())
	result.WriteByte(' ')
	result.WriteString(strconv.Itoa(p.HalfmoveClock))
	result.WriteByte(' ')
	result.WriteString(strconv.Itoa(p.FullmoveNumber))
	return result.String()
}
