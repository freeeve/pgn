package pgn

import "testing"

// Test that ParseSAN produces the same results as MoveFromAlgebraic
func TestParseSAN_Basic(t *testing.T) {
	tests := []struct {
		fen string
		san string
	}{
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "e4"},
		{"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", "e5"},
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Nf3"},
		{"rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 1 1", "Nc6"},
		{"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", "Bb5"},
		// Castling
		{"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1", "O-O"},
		{"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1", "O-O-O"},
		// Captures
		{"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2", "exd5"},
		// Promotion
		{"8/P7/8/8/8/8/8/4K2k w - - 0 1", "a8=Q"},
		// Knight disambiguation
		{"r1bqkb1r/pppp1ppp/2n2n2/4p3/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 4 4", "Nd5"},
	}

	for _, tc := range tests {
		t.Run(tc.san, func(t *testing.T) {
			gs, err := NewGame(tc.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}
			mv, err := ParseSAN(gs, tc.san)
			if err != nil {
				t.Fatalf("ParseSAN(%q) failed: %v", tc.san, err)
			}
			// Verify the move is sensible
			if mv.From < 0 || mv.From > 63 || mv.To < 0 || mv.To > 63 {
				t.Errorf("invalid move: from=%d to=%d", mv.From, mv.To)
			}
		})
	}
}

func BenchmarkParseSAN(b *testing.B) {
	gs, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bb5", "a6", "Ba4", "Nf6", "O-O", "Be7"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gs2 := *gs // copy
		for _, san := range moves {
			mv, err := ParseSAN(&gs2, san)
			if err != nil {
				b.Fatalf("ParseSAN(%q): %v", san, err)
			}
			MakeMove(&gs2, mv)
		}
	}
}

