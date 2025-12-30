package pgn

import "testing"

func TestPackedPosition_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"starting", startingFEN},
		{"after e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"ruy lopez", "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
		{"castling both", "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1"},
		{"castling partial", "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w Kq - 0 1"},
		{"no castling", "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w - - 0 1"},
		{"endgame", "8/8/8/8/8/8/8/4K2k w - - 0 1"},
		{"black to move", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"},
		{"complex midgame", "r1bq1rk1/ppp2ppp/2np1n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQ1RK1 w - - 4 8"},
		{"queens only", "8/8/8/3q4/4Q3/8/8/8 w - - 0 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame: %v", err)
			}

			pp := gs.Pack()
			decoded := pp.Unpack()

			if decoded.SideToMove != gs.SideToMove {
				t.Errorf("SideToMove: got %v, want %v", decoded.SideToMove, gs.SideToMove)
			}
			if decoded.Castle != gs.Castle {
				t.Errorf("Castle: got %d, want %d", decoded.Castle, gs.Castle)
			}

			for sq := Square(0); sq < 64; sq++ {
				if decoded.PieceAt(sq) != gs.PieceAt(sq) {
					t.Errorf("PieceAt(%s): got %c, want %c", sq, decoded.PieceAt(sq), gs.PieceAt(sq))
				}
			}
		})
	}
}

func TestPackedPosition_Base64RoundTrip(t *testing.T) {
	gs := NewStartingPosition()
	pp := gs.Pack()

	encoded := pp.String()
	t.Logf("Starting position base64: %s (len=%d)", encoded, len(encoded))

	if len(encoded) != 46 {
		t.Errorf("Expected 46 char base64, got %d", len(encoded))
	}

	decoded, err := ParsePackedPosition(encoded)
	if err != nil {
		t.Fatalf("ParsePackedPosition: %v", err)
	}

	if pp != decoded {
		t.Errorf("PackedPositions don't match after base64 round-trip")
	}
}

func TestPackedPosition_ToFEN(t *testing.T) {
	tests := []string{
		startingFEN,
		"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
		"8/8/8/8/8/8/8/4K2k w - - 0 1",
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			pp, err := PackFEN(fen)
			if err != nil {
				t.Fatalf("PackFEN: %v", err)
			}

			got := pp.ToFEN()
			if got != fen {
				t.Errorf("ToFEN:\n  got:  %q\n  want: %q", got, fen)
			}
		})
	}
}

func TestPackFEN(t *testing.T) {
	pp, err := PackFEN(startingFEN)
	if err != nil {
		t.Fatalf("PackFEN: %v", err)
	}

	gs := pp.Unpack()
	if gs.SideToMove != White {
		t.Errorf("SideToMove: got %v, want White", gs.SideToMove)
	}
	if gs.PieceAt(SqE1) != 'K' {
		t.Errorf("PieceAt(e1): got %c, want K", gs.PieceAt(SqE1))
	}
	if gs.PieceAt(SqE8) != 'k' {
		t.Errorf("PieceAt(e8): got %c, want k", gs.PieceAt(SqE8))
	}
}

func TestPackedPositionFromFEN(t *testing.T) {
	encoded, err := PackedPositionFromFEN(startingFEN)
	if err != nil {
		t.Fatalf("PackedPositionFromFEN: %v", err)
	}
	t.Logf("Starting position key: %s", encoded)

	if len(encoded) != 46 {
		t.Errorf("Expected 46 char base64, got %d", len(encoded))
	}

	pp, err := ParsePackedPosition(encoded)
	if err != nil {
		t.Fatalf("ParsePackedPosition: %v", err)
	}

	fen := pp.ToFEN()
	if fen != startingFEN {
		t.Errorf("Round-trip failed: got %q", fen)
	}
}

func TestPackedPosition_EnPassant(t *testing.T) {
	tests := []struct {
		fen    string
		epFile int
	}{
		{"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", 4},
		{"rnbqkbnr/pppp1ppp/8/4pP2/8/8/PPPPP1PP/RNBQKBNR w KQkq e6 0 3", 4},
		{"rnbqkbnr/1ppppppp/8/pP6/8/8/P1PPPPPP/RNBQKBNR w KQkq a6 0 3", 0},
		{"rnbqkbnr/ppppppp1/8/6Pp/8/8/PPPPPP1P/RNBQKBNR w KQkq h6 0 3", 7},
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			gs, _ := NewGame(tt.fen)
			pp := gs.Pack()
			decoded := pp.Unpack()

			if decoded.EP < 0 {
				t.Errorf("EP should be set, got %d", decoded.EP)
				return
			}

			epFile := int(decoded.EP % 8)
			if epFile != tt.epFile {
				t.Errorf("EP file: got %d, want %d", epFile, tt.epFile)
			}
		})
	}
}

func TestPackedPosition_NoEnPassant(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	gs, _ := NewGame(fen)
	pp := gs.Pack()
	decoded := pp.Unpack()

	if decoded.EP >= 0 {
		t.Errorf("EP should be -1, got %d", decoded.EP)
	}
}

func TestPackedPosition_AllPieceTypes(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	gs, _ := NewGame(fen)
	pp := gs.Pack()
	decoded := pp.Unpack()

	expectedPieces := map[Square]byte{
		SqA1: 'R', SqB1: 'N', SqC1: 'B', SqD1: 'Q', SqE1: 'K', SqF1: 'B', SqG1: 'N', SqH1: 'R',
		SqA2: 'P', SqB2: 'P', SqC2: 'P', SqD2: 'P', SqE2: 'P', SqF2: 'P', SqG2: 'P', SqH2: 'P',
		SqA7: 'p', SqB7: 'p', SqC7: 'p', SqD7: 'p', SqE7: 'p', SqF7: 'p', SqG7: 'p', SqH7: 'p',
		SqA8: 'r', SqB8: 'n', SqC8: 'b', SqD8: 'q', SqE8: 'k', SqF8: 'b', SqG8: 'n', SqH8: 'r',
	}

	for sq, expected := range expectedPieces {
		got := decoded.PieceAt(sq)
		if got != expected {
			t.Errorf("PieceAt(%s): got %c, want %c", sq, got, expected)
		}
	}

	emptySquares := []Square{SqA3, SqA4, SqA5, SqA6, SqD4, SqE4, SqE5}
	for _, sq := range emptySquares {
		got := decoded.PieceAt(sq)
		if got != 0 {
			t.Errorf("PieceAt(%s): got %c, want empty", sq, got)
		}
	}
}

func TestParsePackedPosition_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "abc"},
		{"invalid base64", "!!!invalid!!!"},
		{"wrong length 33 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{"wrong length 35 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePackedPosition(tt.input)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestPackedPosition_CastlingCombinations(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		castle  uint8
	}{
		{"all", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", 0x0F},
		{"white only", "r3k2r/8/8/8/8/8/8/R3K2R w KQ - 0 1", 0x03},
		{"black only", "r3k2r/8/8/8/8/8/8/R3K2R w kq - 0 1", 0x0C},
		{"kingside only", "r3k2r/8/8/8/8/8/8/R3K2R w Kk - 0 1", 0x05},
		{"queenside only", "r3k2r/8/8/8/8/8/8/R3K2R w Qq - 0 1", 0x0A},
		{"none", "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1", 0x00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, _ := NewGame(tt.fen)
			pp := gs.Pack()
			decoded := pp.Unpack()

			if decoded.Castle != tt.castle {
				t.Errorf("Castle: got %d, want %d", decoded.Castle, tt.castle)
			}
		})
	}
}

func BenchmarkPackedPosition_Pack(b *testing.B) {
	gs := NewStartingPosition()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gs.Pack()
	}
}

func BenchmarkPackedPosition_Unpack(b *testing.B) {
	gs := NewStartingPosition()
	pp := gs.Pack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pp.Unpack()
	}
}

func BenchmarkPackedPosition_String(b *testing.B) {
	gs := NewStartingPosition()
	pp := gs.Pack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pp.String()
	}
}

func BenchmarkParsePackedPosition(b *testing.B) {
	gs := NewStartingPosition()
	pp := gs.Pack()
	s := pp.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePackedPosition(s)
	}
}


func TestPackedFEN_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"starting", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"after_e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"midgame", "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
		{"high_moves", "8/8/8/8/8/8/8/4K2k w - - 99 250"},
		{"very_high_moves", "8/8/8/8/8/8/8/4K2k b - - 0 1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, err := NewGame(tt.fen)
			if err != nil {
				t.Fatalf("NewGame failed: %v", err)
			}

			pf := gs.PackFEN()
			base64Str := pf.String()
			t.Logf("FEN: %s -> base64: %s (len=%d)", tt.fen, base64Str, len(base64Str))

			// Parse back
			pf2, err := ParsePackedFEN(base64Str)
			if err != nil {
				t.Fatalf("ParsePackedFEN failed: %v", err)
			}

			// Unpack and compare
			gs2 := pf2.Unpack()
			fen2 := gs2.ToFEN()

			if fen2 != tt.fen {
				t.Errorf("Round trip failed:\n  input:  %s\n  output: %s", tt.fen, fen2)
			}
		})
	}
}

func TestPackedFEN_ToPackedPosition(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 5 10"
	gs, _ := NewGame(fen)
	
	pf := gs.PackFEN()
	pp := pf.ToPackedPosition()
	
	// PackedPosition should match direct Pack()
	ppDirect := gs.Pack()
	if pp != ppDirect {
		t.Error("ToPackedPosition doesn't match direct Pack()")
	}
}
