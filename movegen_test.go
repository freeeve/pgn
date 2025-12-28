package pgn

import "testing"

func TestWithLegalMoves(t *testing.T) {
	gs, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	WithLegalMoves(gs, func(moves []Mv) {
		if len(moves) != 20 {
			t.Errorf("starting position: expected 20 legal moves, got %d", len(moves))
		}
	})
}

func TestApplyMove(t *testing.T) {
	gs, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	// Copy initial position
	gsCopy := *gs

	// Apply e4
	mv := Mv{From: 12, To: 28}
	ApplyMove(&gsCopy, mv)

	// Verify pawn moved
	pawnOnE2 := gsCopy.pieces[v2WPawn] & (Bitboard(1) << 12)
	pawnOnE4 := gsCopy.pieces[v2WPawn] & (Bitboard(1) << 28)

	if pawnOnE2 != 0 {
		t.Error("ApplyMove: pawn still on e2")
	}
	if pawnOnE4 == 0 {
		t.Error("ApplyMove: no pawn on e4")
	}
	if gsCopy.SideToMove != Black {
		t.Error("ApplyMove: side to move should be Black")
	}
}
