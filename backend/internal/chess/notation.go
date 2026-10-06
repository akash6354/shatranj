package chess

import (
	"fmt"
	"strings"
)

// SAN returns the standard algebraic notation for a legal move.
func (p Position) SAN(move Move) (string, error) {
	if err := p.ValidateMove(move); err != nil {
		return "", err
	}
	piece := p.PieceAt(move.From)
	if piece.Type == King && abs(move.To.File()-move.From.File()) == 2 {
		san := "O-O"
		if move.To.File() < move.From.File() {
			san = "O-O-O"
		}
		next, _ := p.applyUnchecked(move)
		if next.IsCheckmate() {
			san += "#"
		} else if next.IsInCheck(next.SideToMove) {
			san += "+"
		}
		return san, nil
	}
	var result strings.Builder
	if piece.Type != Pawn {
		result.WriteByte(pieceLetter(piece.Type))
		disambiguation := p.disambiguation(move, piece)
		result.WriteString(disambiguation)
	}
	capture := !p.PieceAt(move.To).IsEmpty() ||
		piece.Type == Pawn && move.From.File() != move.To.File()
	if capture {
		if piece.Type == Pawn {
			result.WriteByte(byte('a' + move.From.File()))
		}
		result.WriteByte('x')
	}
	result.WriteString(move.To.String())
	if move.Promotion != NoPiece {
		result.WriteByte('=')
		result.WriteByte(pieceLetter(move.Promotion))
	}
	next, _ := p.applyUnchecked(move)
	if next.IsCheckmate() {
		result.WriteByte('#')
	} else if next.IsInCheck(next.SideToMove) {
		result.WriteByte('+')
	}
	return result.String(), nil
}

func (p Position) disambiguation(move Move, piece Piece) string {
	var alternatives []Move
	for _, other := range p.LegalMoves() {
		if other != move && other.To == move.To && p.PieceAt(other.From).Type == piece.Type {
			alternatives = append(alternatives, other)
		}
	}
	if len(alternatives) == 0 {
		return ""
	}
	fileUnique, rankUnique := true, true
	for _, other := range alternatives {
		if other.From.File() == move.From.File() {
			fileUnique = false
		}
		if other.From.Rank() == move.From.Rank() {
			rankUnique = false
		}
	}
	switch {
	case fileUnique:
		return string(byte('a' + move.From.File()))
	case rankUnique:
		return string(byte('1' + move.From.Rank()))
	default:
		return move.From.String()
	}
}

// ParseSAN resolves a standard algebraic move in the current position.
func (p Position) ParseSAN(text string) (Move, error) {
	wanted := normalizeSAN(text)
	for _, move := range p.LegalMoves() {
		san, err := p.SAN(move)
		if err == nil && normalizeSAN(san) == wanted {
			return move, nil
		}
	}
	return Move{}, fmt.Errorf("invalid SAN move %q", text)
}

func normalizeSAN(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "0-0-0", "O-O-O")
	text = strings.ReplaceAll(text, "0-0", "O-O")
	text = strings.TrimSuffix(text, "e.p.")
	text = strings.TrimRight(text, "+#!?")
	return strings.TrimSpace(text)
}

func pieceLetter(kind PieceType) byte {
	switch kind {
	case Knight:
		return 'N'
	case Bishop:
		return 'B'
	case Rook:
		return 'R'
	case Queen:
		return 'Q'
	case King:
		return 'K'
	default:
		return 0
	}
}
