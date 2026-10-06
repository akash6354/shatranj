package chess

import "fmt"

var knightOffsets = [][2]int{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
var kingOffsets = [][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
var bishopDirections = [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
var rookDirections = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

func (p Position) IsSquareAttacked(square Square, by Color) bool {
	if !square.Valid() || (by != White && by != Black) {
		return false
	}
	file, rank := square.File(), square.Rank()
	pawnSourceRank := rank - 1
	if by == Black {
		pawnSourceRank = rank + 1
	}
	for _, df := range []int{-1, 1} {
		source := NewSquare(file+df, pawnSourceRank)
		if source.Valid() && p.PieceAt(source) == (Piece{Color: by, Type: Pawn}) {
			return true
		}
	}
	for _, offset := range knightOffsets {
		source := NewSquare(file+offset[0], rank+offset[1])
		if source.Valid() && p.PieceAt(source) == (Piece{Color: by, Type: Knight}) {
			return true
		}
	}
	for _, offset := range kingOffsets {
		source := NewSquare(file+offset[0], rank+offset[1])
		if source.Valid() && p.PieceAt(source) == (Piece{Color: by, Type: King}) {
			return true
		}
	}
	if p.attackedBySlider(file, rank, by, bishopDirections, Bishop) ||
		p.attackedBySlider(file, rank, by, rookDirections, Rook) {
		return true
	}
	return false
}

func (p Position) attackedBySlider(file, rank int, by Color, directions [][2]int, slider PieceType) bool {
	for _, direction := range directions {
		for f, r := file+direction[0], rank+direction[1]; f >= 0 && f < 8 && r >= 0 && r < 8; f, r = f+direction[0], r+direction[1] {
			piece := p.PieceAt(NewSquare(f, r))
			if piece.IsEmpty() {
				continue
			}
			if piece.Color == by && (piece.Type == slider || piece.Type == Queen) {
				return true
			}
			break
		}
	}
	return false
}

func (p Position) IsInCheck(color Color) bool {
	king, ok := p.KingSquare(color)
	return ok && p.IsSquareAttacked(king, color.Opposite())
}

// LegalMoves returns every move that leaves the moving side's king safe.
func (p Position) LegalMoves() []Move {
	if p.SideToMove != White && p.SideToMove != Black {
		return nil
	}
	pseudo := p.pseudoLegalMoves()
	legal := make([]Move, 0, len(pseudo))
	for _, move := range pseudo {
		next, err := p.applyUnchecked(move)
		if err == nil && !next.IsInCheck(p.SideToMove) {
			legal = append(legal, move)
		}
	}
	return legal
}

func (p Position) pseudoLegalMoves() []Move {
	moves := make([]Move, 0, 40)
	for index, piece := range p.Board {
		if piece.Color != p.SideToMove {
			continue
		}
		from := Square(index)
		switch piece.Type {
		case Pawn:
			moves = p.pawnMoves(moves, from, piece)
		case Knight:
			for _, offset := range knightOffsets {
				to := NewSquare(from.File()+offset[0], from.Rank()+offset[1])
				moves = p.addIfAvailable(moves, from, to, piece.Color)
			}
		case Bishop:
			moves = p.slidingMoves(moves, from, piece.Color, bishopDirections)
		case Rook:
			moves = p.slidingMoves(moves, from, piece.Color, rookDirections)
		case Queen:
			moves = p.slidingMoves(moves, from, piece.Color, bishopDirections)
			moves = p.slidingMoves(moves, from, piece.Color, rookDirections)
		case King:
			for _, offset := range kingOffsets {
				to := NewSquare(from.File()+offset[0], from.Rank()+offset[1])
				moves = p.addIfAvailable(moves, from, to, piece.Color)
			}
			moves = p.castlingMoves(moves, from, piece.Color)
		}
	}
	return moves
}

func (p Position) addIfAvailable(moves []Move, from, to Square, color Color) []Move {
	if !to.Valid() {
		return moves
	}
	target := p.PieceAt(to)
	if target.IsEmpty() || target.Color != color && target.Type != King {
		return append(moves, Move{From: from, To: to})
	}
	return moves
}

func (p Position) pawnMoves(moves []Move, from Square, piece Piece) []Move {
	direction, startRank, promotionRank := 1, 1, 7
	if piece.Color == Black {
		direction, startRank, promotionRank = -1, 6, 0
	}
	file, rank := from.File(), from.Rank()
	forward := NewSquare(file, rank+direction)
	if forward.Valid() && p.PieceAt(forward).IsEmpty() {
		moves = p.addPawnMove(moves, from, forward, promotionRank)
		twoForward := NewSquare(file, rank+2*direction)
		if rank == startRank && p.PieceAt(twoForward).IsEmpty() {
			moves = append(moves, Move{From: from, To: twoForward})
		}
	}
	for _, df := range []int{-1, 1} {
		to := NewSquare(file+df, rank+direction)
		if !to.Valid() {
			continue
		}
		target := p.PieceAt(to)
		if !target.IsEmpty() && target.Color != piece.Color && target.Type != King {
			moves = p.addPawnMove(moves, from, to, promotionRank)
		} else if to == p.EnPassant {
			captured := NewSquare(to.File(), to.Rank()-direction)
			if p.PieceAt(captured) == (Piece{Color: piece.Color.Opposite(), Type: Pawn}) {
				moves = append(moves, Move{From: from, To: to})
			}
		}
	}
	return moves
}

func (p Position) addPawnMove(moves []Move, from, to Square, promotionRank int) []Move {
	if to.Rank() != promotionRank {
		return append(moves, Move{From: from, To: to})
	}
	for _, promotion := range []PieceType{Queen, Rook, Bishop, Knight} {
		moves = append(moves, Move{From: from, To: to, Promotion: promotion})
	}
	return moves
}

func (p Position) slidingMoves(moves []Move, from Square, color Color, directions [][2]int) []Move {
	for _, direction := range directions {
		for file, rank := from.File()+direction[0], from.Rank()+direction[1]; file >= 0 && file < 8 && rank >= 0 && rank < 8; file, rank = file+direction[0], rank+direction[1] {
			to := NewSquare(file, rank)
			target := p.PieceAt(to)
			if target.IsEmpty() {
				moves = append(moves, Move{From: from, To: to})
				continue
			}
			if target.Color != color && target.Type != King {
				moves = append(moves, Move{From: from, To: to})
			}
			break
		}
	}
	return moves
}

func (p Position) castlingMoves(moves []Move, king Square, color Color) []Move {
	if p.IsInCheck(color) {
		return moves
	}
	homeRank := 0
	kingSide, queenSide := WhiteKingSide, WhiteQueenSide
	if color == Black {
		homeRank = 7
		kingSide, queenSide = BlackKingSide, BlackQueenSide
	}
	if king != NewSquare(4, homeRank) {
		return moves
	}
	if p.Castling&kingSide != 0 &&
		p.PieceAt(NewSquare(5, homeRank)).IsEmpty() &&
		p.PieceAt(NewSquare(6, homeRank)).IsEmpty() &&
		p.PieceAt(NewSquare(7, homeRank)) == (Piece{Color: color, Type: Rook}) &&
		!p.IsSquareAttacked(NewSquare(5, homeRank), color.Opposite()) &&
		!p.IsSquareAttacked(NewSquare(6, homeRank), color.Opposite()) {
		moves = append(moves, Move{From: king, To: NewSquare(6, homeRank)})
	}
	if p.Castling&queenSide != 0 &&
		p.PieceAt(NewSquare(1, homeRank)).IsEmpty() &&
		p.PieceAt(NewSquare(2, homeRank)).IsEmpty() &&
		p.PieceAt(NewSquare(3, homeRank)).IsEmpty() &&
		p.PieceAt(NewSquare(0, homeRank)) == (Piece{Color: color, Type: Rook}) &&
		!p.IsSquareAttacked(NewSquare(3, homeRank), color.Opposite()) &&
		!p.IsSquareAttacked(NewSquare(2, homeRank), color.Opposite()) {
		moves = append(moves, Move{From: king, To: NewSquare(2, homeRank)})
	}
	return moves
}

// ValidateMove verifies that a move is legal in this position.
func (p Position) ValidateMove(move Move) error {
	if !move.From.Valid() || !move.To.Valid() || move.From == move.To {
		return fmt.Errorf("invalid move squares")
	}
	for _, legal := range p.LegalMoves() {
		if legal == move {
			return nil
		}
	}
	return fmt.Errorf("illegal move %s in position", move)
}

// MakeMove returns a new position after a legal move; the original is unchanged.
func (p Position) MakeMove(move Move) (Position, error) {
	if err := p.ValidateMove(move); err != nil {
		return Position{}, err
	}
	return p.applyUnchecked(move)
}

func (p Position) applyUnchecked(move Move) (Position, error) {
	if !move.From.Valid() || !move.To.Valid() {
		return Position{}, fmt.Errorf("invalid move squares")
	}
	piece := p.PieceAt(move.From)
	if piece.IsEmpty() {
		return Position{}, fmt.Errorf("no piece on %s", move.From)
	}
	movingType := piece.Type
	target := p.PieceAt(move.To)
	isEnPassant := piece.Type == Pawn && move.To == p.EnPassant && target.IsEmpty() && move.From.File() != move.To.File()
	isCastle := piece.Type == King && abs(move.To.File()-move.From.File()) == 2
	if isEnPassant {
		captured := NewSquare(move.To.File(), move.From.Rank())
		p.Board[captured] = Piece{}
	}
	p.Board[move.From] = Piece{}
	if isCastle {
		rank := move.From.Rank()
		rookFrom, rookTo := NewSquare(7, rank), NewSquare(5, rank)
		if move.To.File() < move.From.File() {
			rookFrom, rookTo = NewSquare(0, rank), NewSquare(3, rank)
		}
		p.Board[rookTo] = p.Board[rookFrom]
		p.Board[rookFrom] = Piece{}
	}
	if piece.Type == Pawn && (move.To.Rank() == 0 || move.To.Rank() == 7) {
		piece.Type = move.Promotion
	}
	p.Board[move.To] = piece

	p.updateCastlingRights(move, piece)
	p.EnPassant = NoSquare
	if piece.Type == Pawn && abs(move.To.Rank()-move.From.Rank()) == 2 {
		p.EnPassant = NewSquare(move.From.File(), (move.To.Rank()+move.From.Rank())/2)
	}
	if movingType == Pawn || !target.IsEmpty() || isEnPassant {
		p.HalfmoveClock = 0
	} else {
		p.HalfmoveClock++
	}
	if p.SideToMove == Black {
		p.FullmoveNumber++
	}
	p.SideToMove = p.SideToMove.Opposite()
	return p, nil
}

func (p *Position) updateCastlingRights(move Move, moving Piece) {
	switch moving.Color {
	case White:
		if moving.Type == King {
			p.Castling &^= WhiteKingSide | WhiteQueenSide
		}
		if moving.Type == Rook && move.From == NewSquare(0, 0) {
			p.Castling &^= WhiteQueenSide
		}
		if moving.Type == Rook && move.From == NewSquare(7, 0) {
			p.Castling &^= WhiteKingSide
		}
	case Black:
		if moving.Type == King {
			p.Castling &^= BlackKingSide | BlackQueenSide
		}
		if moving.Type == Rook && move.From == NewSquare(0, 7) {
			p.Castling &^= BlackQueenSide
		}
		if moving.Type == Rook && move.From == NewSquare(7, 7) {
			p.Castling &^= BlackKingSide
		}
	}
	switch move.To {
	case NewSquare(0, 0):
		p.Castling &^= WhiteQueenSide
	case NewSquare(7, 0):
		p.Castling &^= WhiteKingSide
	case NewSquare(0, 7):
		p.Castling &^= BlackQueenSide
	case NewSquare(7, 7):
		p.Castling &^= BlackKingSide
	}
}

func (p Position) IsCheckmate() bool {
	return p.IsInCheck(p.SideToMove) && len(p.LegalMoves()) == 0
}

func (p Position) IsStalemate() bool {
	return !p.IsInCheck(p.SideToMove) && len(p.LegalMoves()) == 0
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
