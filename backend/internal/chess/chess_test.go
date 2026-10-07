package chess

import (
	"strings"
	"testing"
)

func mustFEN(t *testing.T, fen string) Position {
	t.Helper()
	position, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN(%q): %v", fen, err)
	}
	return position
}

func mustMove(t *testing.T, position Position, coordinate string) Position {
	t.Helper()
	move, err := ParseMove(coordinate)
	if err != nil {
		t.Fatal(err)
	}
	next, err := position.MakeMove(move)
	if err != nil {
		t.Fatalf("MakeMove(%s): %v", coordinate, err)
	}
	return next
}

func TestStartingPositionAndFENRoundTrip(t *testing.T) {
	position := StartingPosition()
	if got, want := len(position.LegalMoves()), 20; got != want {
		t.Fatalf("starting legal moves = %d, want %d", got, want)
	}
	const startingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if got := position.FEN(); got != startingFEN {
		t.Fatalf("starting FEN = %q, want %q", got, startingFEN)
	}
	parsed := mustFEN(t, startingFEN)
	if got := parsed.FEN(); got != startingFEN {
		t.Fatalf("round-trip FEN = %q, want %q", got, startingFEN)
	}
	if got := perft(position, 3); got != 8902 {
		t.Fatalf("starting-position perft(3) = %d, want 8902", got)
	}
}

func TestPawnMovesAndIllegalMoves(t *testing.T) {
	position := StartingPosition()
	position = mustMove(t, position, "e2e4")
	if position.EnPassant.String() != "e3" {
		t.Fatalf("en passant target = %s, want e3", position.EnPassant)
	}
	if position.PieceAt(mustSquare(t, "e2")).Type != NoPiece {
		t.Fatal("original pawn square was not cleared")
	}
	if _, err := StartingPosition().MakeMove(Move{From: mustSquare(t, "e2"), To: mustSquare(t, "e5")}); err == nil {
		t.Fatal("illegal three-square pawn move accepted")
	}
	if _, err := StartingPosition().MakeMove(Move{From: mustSquare(t, "e2"), To: mustSquare(t, "e4"), Promotion: Queen}); err == nil {
		t.Fatal("promotion on a non-promotion rank accepted")
	}
}

func TestCastling(t *testing.T) {
	position := mustFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	position = mustMove(t, position, "e1g1")
	if position.PieceAt(mustSquare(t, "g1")) != (Piece{Color: White, Type: King}) ||
		position.PieceAt(mustSquare(t, "f1")) != (Piece{Color: White, Type: Rook}) {
		t.Fatalf("castling pieces not placed correctly: %s", position.FEN())
	}
	if position.Castling&(WhiteKingSide|WhiteQueenSide) != 0 {
		t.Fatal("white castling rights remain after king moved")
	}

	attacked := mustFEN(t, "r3k2r/8/8/8/8/8/5r2/R3K2R w KQkq - 0 1")
	if _, err := attacked.MakeMove(Move{From: mustSquare(t, "e1"), To: mustSquare(t, "g1")}); err == nil {
		t.Fatal("castling through an attacked square accepted")
	}
}

func TestEnPassant(t *testing.T) {
	position := mustFEN(t, "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1")
	position = mustMove(t, position, "e5d6")
	if position.PieceAt(mustSquare(t, "d5")).Type != NoPiece {
		t.Fatal("en passant captured pawn remains on board")
	}
	if position.PieceAt(mustSquare(t, "d6")) != (Piece{Color: White, Type: Pawn}) {
		t.Fatal("capturing pawn not placed on en passant target")
	}
}

func TestPromotion(t *testing.T) {
	position := mustFEN(t, "4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	position = mustMove(t, position, "a7a8n")
	if position.PieceAt(mustSquare(t, "a8")) != (Piece{Color: White, Type: Knight}) {
		t.Fatalf("promotion result = %+v", position.PieceAt(mustSquare(t, "a8")))
	}
	if position.HalfmoveClock != 0 {
		t.Fatalf("halfmove clock after pawn promotion = %d, want 0", position.HalfmoveClock)
	}
}

func TestCheckMateAndStalemate(t *testing.T) {
	checkmate := mustFEN(t, "7k/6Q1/5K2/8/8/8/8/8 b - - 0 1")
	if !checkmate.IsInCheck(Black) || !checkmate.IsCheckmate() {
		t.Fatalf("expected checkmate: check=%v legal=%v", checkmate.IsInCheck(Black), checkmate.LegalMoves())
	}
	stalemate := mustFEN(t, "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")
	if stalemate.IsInCheck(Black) || !stalemate.IsStalemate() {
		t.Fatalf("expected stalemate: check=%v legal=%v", stalemate.IsInCheck(Black), stalemate.LegalMoves())
	}
}

func TestSANAndPGNRoundTrip(t *testing.T) {
	position := StartingPosition()
	move, err := position.ParseSAN("e4")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := position.SAN(move); err != nil || got != "e4" {
		t.Fatalf("SAN = %q, err=%v; want e4", got, err)
	}
	game := PGN{
		Tags: map[string]string{"Event": "Test"},
		Moves: []Move{
			move,
			{From: mustSquare(t, "e7"), To: mustSquare(t, "e5")},
			{From: mustSquare(t, "g1"), To: mustSquare(t, "f3")},
		},
		Result: "*",
	}
	text, err := FormatPGN(game)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParsePGN(text)
	if err != nil {
		t.Fatalf("ParsePGN(%q): %v", text, err)
	}
	if len(parsed.Moves) != len(game.Moves) || parsed.Tags["Event"] != "Test" {
		t.Fatalf("parsed PGN = %+v", parsed)
	}
	if !strings.Contains(text, "1. e4 e5 2. Nf3") {
		t.Fatalf("unexpected PGN text %q", text)
	}
}

func TestInsufficientMaterialAndDrawClock(t *testing.T) {
	position := mustFEN(t, "4k3/8/8/8/8/8/8/4K3 w - - 100 1")
	if !position.IsInsufficientMaterial() || !position.IsDrawByFiftyMoveRule() {
		t.Fatal("bare kings at halfmove 100 should be drawn")
	}
	if position.IsDrawBySeventyFiveMoveRule() {
		t.Fatal("halfmove 100 is claimable, not an automatic 75-move draw")
	}
	position.HalfmoveClock = 150
	if !position.IsDrawBySeventyFiveMoveRule() {
		t.Fatal("halfmove 150 should be an automatic 75-move draw")
	}
}

func TestRepetitionKeyNormalizesUncapturableEnPassant(t *testing.T) {
	withoutEnPassant := mustFEN(t, "4k3/8/8/3p4/8/8/8/4K3 w - - 0 1")
	withUncapturableEnPassant := mustFEN(t, "4k3/8/8/3p4/8/8/8/4K3 w - d6 0 1")
	if withoutEnPassant.RepetitionKey() != withUncapturableEnPassant.RepetitionKey() {
		t.Fatalf("uncapturable en passant changed repetition key: %q != %q", withoutEnPassant.RepetitionKey(), withUncapturableEnPassant.RepetitionKey())
	}
	withCapturableEnPassant := mustFEN(t, "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1")
	if withCapturableEnPassant.RepetitionKey() == withoutEnPassant.RepetitionKey() {
		t.Fatal("capturable en passant must distinguish repetition positions")
	}
}

func mustSquare(t *testing.T, name string) Square {
	t.Helper()
	square, err := ParseSquare(name)
	if err != nil {
		t.Fatal(err)
	}
	return square
}

func perft(position Position, depth int) int {
	if depth == 0 {
		return 1
	}
	nodes := 0
	for _, move := range position.LegalMoves() {
		next, err := position.applyUnchecked(move)
		if err != nil {
			panic(err)
		}
		nodes += perft(next, depth-1)
	}
	return nodes
}
