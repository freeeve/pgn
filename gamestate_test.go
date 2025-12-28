package pgn

import "testing"

func TestGameState_ToFEN(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"starting", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"after e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"sicilian", "rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2"},
		{"no castling", "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w - - 0 1"},
		{"white only castle", "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQ - 0 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame(%q) error: %v", tt.fen, err)
			}
			got := gs.ToFEN()
			if got != tt.fen {
				t.Errorf("ToFEN() = %q, want %q", got, tt.fen)
			}
		})
	}
}

func TestGameState_RoundTrip(t *testing.T) {
	// Test known positions round-trip correctly
	positions := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	}

	for _, fen := range positions {
		gs, err := NewGame(fen)
		if err != nil {
			t.Fatalf("NewGame(%q) error: %v", fen, err)
		}
		got := gs.ToFEN()
		if got != fen {
			t.Errorf("Round-trip failed:\n  got:  %q\n  want: %q", got, fen)
		}
	}
}

func TestPieceAtV2(t *testing.T) {
	gs, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	// Test some known squares
	tests := []struct {
		sq   Square
		want byte
	}{
		{0, 'R'},  // a1 = white rook
		{4, 'K'},  // e1 = white king
		{8, 'P'},  // a2 = white pawn
		{56, 'r'}, // a8 = black rook
		{60, 'k'}, // e8 = black king
		{48, 'p'}, // a7 = black pawn
		{32, 0},   // a5 = empty
	}

	for _, tt := range tests {
		got := pieceAtV2(gs, tt.sq)
		if got != tt.want {
			t.Errorf("pieceAtV2(sq=%d) = %c, want %c", tt.sq, got, tt.want)
		}
	}
}

func TestMoveString(t *testing.T) {
	// Test the legacy Move.String() method
	m := Move{From: D2, To: D4}
	s := m.String()
	// Position uses file byte + rank byte, e.g. 'd' + '2'
	if len(s) < 4 {
		t.Errorf("Move.String() = %q, expected at least 4 chars", s)
	}
	t.Logf("Move.String() = %q", s)

	// With promotion
	m2 := Move{From: E7, To: E8, Promote: WhiteQueen}
	s2 := m2.String()
	if len(s2) < 5 {
		t.Errorf("Move.String() with promo = %q, expected at least 5 chars", s2)
	}
	t.Logf("Move.String() with promo = %q", s2)
}

