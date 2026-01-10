package pgn

import (
	"testing"
)

func TestGenerateParentPositions_StartingPosition(t *testing.T) {
	// Starting position has no parent (or very limited ones)
	pos := NewStartingPosition()
	parents := GenerateParentPositions(pos)

	// Starting position could theoretically have parents, but they'd be
	// unusual positions. The main test is that we don't crash.
	t.Logf("Starting position has %d candidate parents", len(parents))
}

func TestGenerateParentPositions_AfterE4(t *testing.T) {
	// Position after 1. e4
	pos, err := NewGame("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		t.Fatal(err)
	}

	parents := GenerateParentPositions(pos)

	// The starting position should be among the parents
	startingFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	startPos, _ := NewGame(startingFEN)
	startPacked := startPos.Pack()

	found := false
	var foundMove Mv
	for _, p := range parents {
		if p.Position == startPacked {
			found = true
			foundMove = p.Move
			break
		}
	}

	if !found {
		t.Error("Starting position not found among parent candidates")
	} else {
		// The move should be e2-e4
		if foundMove.From != SqE2 || foundMove.To != SqE4 {
			t.Errorf("Expected move e2e4, got %v-%v", foundMove.From, foundMove.To)
		}
		// Validate the move is legal from parent
		if !ValidateParentMoveQuick(startPos, foundMove) {
			t.Error("Move e2e4 should be legal from starting position")
		}
	}

	t.Logf("Position after e4 has %d candidate parents", len(parents))
}

func TestGenerateParentPositions_AfterCapture(t *testing.T) {
	// Position after 1. e4 d5 2. exd5
	pos, err := NewGame("rnbqkbnr/ppp1pppp/8/3P4/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2")
	if err != nil {
		t.Fatal(err)
	}

	parents := GenerateParentPositions(pos)

	// Parent should be position after 1. e4 d5 (white pawn on e4, black pawn on d5)
	parentFEN := "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2"
	parentPos, _ := NewGame(parentFEN)
	parentPacked := parentPos.Pack()

	found := false
	var foundMove Mv
	for _, p := range parents {
		if p.Position == parentPacked {
			found = true
			foundMove = p.Move
			break
		}
	}

	if !found {
		t.Error("Parent position (before exd5) not found among candidates")
	} else {
		// The move should be e4xd5
		if foundMove.From != SqE4 || foundMove.To != SqD5 {
			t.Errorf("Expected move e4d5, got %v-%v", foundMove.From, foundMove.To)
		}
		if !ValidateParentMoveQuick(parentPos, foundMove) {
			t.Error("Move exd5 should be legal from parent position")
		}
	}

	t.Logf("Position after exd5 has %d candidate parents", len(parents))
}

func TestGenerateParentPositions_AfterKnightMove(t *testing.T) {
	// Position after 1. Nf3
	pos, err := NewGame("rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 1 1")
	if err != nil {
		t.Fatal(err)
	}

	parents := GenerateParentPositions(pos)

	// Starting position should be among parents
	startPos := NewStartingPosition()
	startPacked := startPos.Pack()

	found := false
	for _, p := range parents {
		if p.Position == startPacked {
			found = true
			// Move should be g1-f3
			if p.Move.From != SqG1 || p.Move.To != SqF3 {
				t.Errorf("Expected move g1f3, got %v-%v", p.Move.From, p.Move.To)
			}
			break
		}
	}

	if !found {
		t.Error("Starting position not found among parent candidates for Nf3")
	}
}

func TestGenerateParentPositions_AfterCastling(t *testing.T) {
	// Position after white kingside castling
	pos, err := NewGame("r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 3 3")
	if err != nil {
		t.Fatal(err)
	}

	parents := GenerateParentPositions(pos)

	// Parent should have king on e1, rook on h1
	// After 1.e4 e5 2.Nf3 Nc6, before O-O
	parentFEN := "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 3"
	parentPos, _ := NewGame(parentFEN)
	parentPacked := parentPos.Pack()

	found := false
	for _, p := range parents {
		if p.Position == parentPacked {
			found = true
			// Move should be e1-g1 with castle flag
			if p.Move.From != SqE1 || p.Move.To != SqG1 || p.Move.Flags != 4 {
				t.Errorf("Expected castle move e1g1 with flag 4, got from=%v to=%v flags=%d",
					p.Move.From, p.Move.To, p.Move.Flags)
			}
			break
		}
	}

	if !found {
		// Debug: print some candidates to see what we're getting
		t.Logf("Looking for parent: %s", parentFEN)
		t.Logf("Parent packed: %s", parentPacked.String())

		// Check if we have any castle moves
		for _, p := range parents {
			if p.Move.Flags == 4 {
				unp := p.Position.Unpack()
				t.Logf("Found castle candidate: %s, move=%v-%v", unp.ToFEN(), p.Move.From, p.Move.To)
			}
		}
		t.Error("Parent position (before O-O) not found among candidates")
	}
}

func TestGenerateParentPositions_AfterPromotion(t *testing.T) {
	// Position where black just moved (so we're looking for white's previous promotion)
	// Queen on a8 could have been a promoted pawn
	pos, _ := NewGame("Q7/8/8/8/8/8/8/4K2k b - - 0 1")
	parents := GenerateParentPositions(pos)

	// We should find the parent where white had a pawn on a7
	parentFEN := "8/P7/8/8/8/8/8/4K2k w - - 0 1"
	parentPos, _ := NewGame(parentFEN)
	parentPacked := parentPos.Pack()

	found := false
	for _, p := range parents {
		if p.Position == parentPacked {
			found = true
			if p.Move.From != SqA7 || p.Move.To != SqA8 || p.Move.Promo != PromoQueen {
				t.Errorf("Expected promotion move a7a8=Q, got from=%v to=%v promo=%v",
					p.Move.From, p.Move.To, p.Move.Promo)
			}
			break
		}
	}

	if !found {
		t.Error("Parent position (before a8=Q) not found among candidates")
	}

	t.Logf("Position has %d candidate parents (checking for promotion)", len(parents))
}

func TestGenerateParentPositions_BishopMove(t *testing.T) {
	// Simple position with bishop that can be traced back
	pos, _ := NewGame("8/8/8/8/8/5B2/8/4K2k b - - 0 1")
	parents := GenerateParentPositions(pos)

	// Bishop on f3 could have come from many squares
	t.Logf("Simple bishop position has %d candidate parents", len(parents))

	// Verify at least one parent has bishop coming from a diagonal
	hasValidParent := false
	for _, p := range parents {
		parent := p.Position.Unpack()
		// Check if this parent has bishop on a different diagonal square
		if parent.pieces[v2WBishop] != 0 && parent.pieces[v2WBishop] != (1<<SqF3) {
			hasValidParent = true
			break
		}
	}

	if !hasValidParent && len(parents) > 0 {
		// There should be parents with bishop on different squares
		t.Log("Warning: no parents found with bishop on different square")
	}
}

func TestValidateParentMove(t *testing.T) {
	// Test the validation helper
	pos := NewStartingPosition()

	// e2-e4 should be valid
	validMove := Mv{From: SqE2, To: SqE4, Flags: 1}
	if !ValidateParentMoveQuick(pos, validMove) {
		t.Error("e2e4 should be valid from starting position")
	}

	// e2-e5 should be invalid (can't move 3 squares)
	invalidMove := Mv{From: SqE2, To: SqE5}
	if ValidateParentMoveQuick(pos, invalidMove) {
		t.Error("e2e5 should be invalid from starting position")
	}
}

func TestGenerateParentPositions_NoNilPanic(t *testing.T) {
	// Should not panic on nil input
	parents := GenerateParentPositions(nil)
	if parents != nil {
		t.Error("Expected nil result for nil input")
	}
}

func TestRetroMoveCount(t *testing.T) {
	// Test that we generate a reasonable number of candidates
	testCases := []struct {
		name string
		fen  string
	}{
		{"Starting", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"After e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"Middle game", "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
		{"Endgame", "8/8/8/3k4/8/8/3K4/8 w - - 0 1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos, err := NewGame(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			parents := GenerateParentPositions(pos)
			t.Logf("%s: %d candidate parents", tc.name, len(parents))

			// Basic sanity check - should have some candidates (except maybe endgame king-only)
			if tc.name != "Endgame" && tc.name != "Starting" && len(parents) < 5 {
				t.Logf("Warning: unexpectedly few candidates for %s", tc.name)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that applying a move and finding parents works
	pos := NewStartingPosition()

	// Apply e4
	move := Mv{From: SqE2, To: SqE4, Flags: 1}
	child := pos.Copy()
	err := ApplyMove(child, move)
	if err != nil {
		t.Fatal(err)
	}

	// Generate parents of child
	parents := GenerateParentPositions(child)

	// Find the original position
	originalPacked := pos.Pack()
	found := false
	for _, p := range parents {
		if p.Position == originalPacked {
			found = true
			// Verify the connecting move matches
			if p.Move.From != move.From || p.Move.To != move.To {
				t.Errorf("Move mismatch: expected %v-%v, got %v-%v",
					move.From, move.To, p.Move.From, p.Move.To)
			}
			break
		}
	}

	if !found {
		t.Error("Original position not found in parents after applying move")
	}
}

func BenchmarkGenerateParentPositions(b *testing.B) {
	pos, _ := NewGame("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateParentPositions(pos)
	}
}
