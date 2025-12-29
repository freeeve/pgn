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

func TestIsCheckmate(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want bool
	}{
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", false},
		{"fool's mate", "rnb1kbnr/pppp1ppp/4p3/8/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3", true},
		{"back rank mate", "3R2k1/5ppp/8/8/8/8/8/4K3 b - - 0 1", true},
		{"scholar's mate", "r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4", true},
		{"check but not mate", "rnbqkbnr/ppppp1pp/5p2/8/4P2Q/8/PPPP1PPP/RNB1KBNR b KQkq - 1 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame(%q) error: %v", tt.fen, err)
			}
			got := gs.IsCheckmate()
			if got != tt.want {
				t.Errorf("IsCheckmate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStalemate(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want bool
	}{
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", false},
		{"stalemate - king in corner", "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1", true}, // black king h8, white queen f7, white king g6
		{"stalemate - classic", "8/8/8/8/8/5k2/5p2/5K2 w - - 0 1", true},
		{"checkmate not stalemate", "rnb1kbnr/pppp1ppp/4p3/8/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame(%q) error: %v", tt.fen, err)
			}
			got := gs.IsStalemate()
			if got != tt.want {
				t.Errorf("IsStalemate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKingSquare(t *testing.T) {
	tests := []struct {
		name       string
		fen        string
		color      Color
		wantSquare Square
	}{
		{"starting white king", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", White, 4},  // e1
		{"starting black king", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", Black, 60}, // e8
		{"white king on g1", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQ1RK1 w kq - 0 1", White, 6},       // g1
		{"black king on d8", "r1bk1bnr/pppppppp/2n5/8/8/8/PPPPPPPP/RNBQKBNR w KQ - 0 1", Black, 59},    // d8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame(%q) error: %v", tt.fen, err)
			}
			got := gs.KingSquare(tt.color)
			if got != tt.wantSquare {
				t.Errorf("KingSquare(%v) = %d (%s), want %d", tt.color, got, got.String(), tt.wantSquare)
			}
		})
	}
}

func TestIsSquareAttacked(t *testing.T) {
	// Starting position
	gs, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	tests := []struct {
		sq      Square
		byColor Color
		want    bool
	}{
		{16, White, true},  // a3 attacked by white pawn on a2
		{20, White, true},  // e3 attacked by white pawns d2, f2
		{32, White, false}, // a5 not attacked by white
		{40, Black, true},  // a6 attacked by black pawn on b7
		{44, Black, true},  // e6 attacked by black pawns on d7, f7
		{32, Black, false}, // a5 not attacked by black
		{21, White, true},  // f3 attacked by white knight g1
		{42, Black, true},  // c6 attacked by black knight b8
	}

	for _, tt := range tests {
		name := tt.sq.String()
		if tt.byColor == White {
			name += " by white"
		} else {
			name += " by black"
		}
		t.Run(name, func(t *testing.T) {
			got := gs.IsSquareAttacked(tt.sq, tt.byColor)
			if got != tt.want {
				t.Errorf("IsSquareAttacked(%s, %v) = %v, want %v", tt.sq.String(), tt.byColor, got, tt.want)
			}
		})
	}
}

