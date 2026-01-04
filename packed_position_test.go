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

			// Check all pieces match
			for sq := Square(0); sq < 64; sq++ {
				if decoded.PieceAt(sq) != gs.PieceAt(sq) {
					t.Errorf("PieceAt(%s): got %c, want %c", sq, decoded.PieceAt(sq), gs.PieceAt(sq))
				}
			}

			// Check metadata matches
			if decoded.SideToMove != gs.SideToMove {
				t.Errorf("SideToMove: got %v, want %v", decoded.SideToMove, gs.SideToMove)
			}
			if decoded.Castle != gs.Castle {
				t.Errorf("Castle: got %d, want %d", decoded.Castle, gs.Castle)
			}
			if decoded.EP != gs.EP {
				t.Errorf("EP: got %d, want %d", decoded.EP, gs.EP)
			}
		})
	}
}

func TestPackedPosition_Base64RoundTrip(t *testing.T) {
	gs := NewStartingPosition()
	pp := gs.Pack()

	encoded := pp.String()
	t.Logf("Starting position base64: %s (len=%d)", encoded, len(encoded))

	// 26 bytes -> 35 base64 chars
	if len(encoded) != 35 {
		t.Errorf("Expected 35 char base64, got %d", len(encoded))
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
	// PackedPosition stores board + metadata (side, castle, EP), but not move counters
	tests := []string{
		startingFEN,
		"r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
		"8/8/8/8/8/8/8/4K2k w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			gs, _ := NewGame(fen)
			pp := gs.Pack()

			// ToFEN will have default halfmove=0, fullmove=1
			got := pp.ToFEN()
			// Parse both and compare board state
			gsGot, _ := NewGame(got)
			if !gs.BoardEquals(gsGot) {
				t.Errorf("ToFEN board mismatch:\n  got:  %q\n  want: %q", got, fen)
			}
		})
	}
}

func TestPackedPosition_Getters(t *testing.T) {
	gs := NewStartingPosition()
	pp := gs.Pack()

	// Test Occupancy
	occ := pp.Occupancy()
	if occ == 0 {
		t.Error("Occupancy should not be 0 for starting position")
	}

	// Test PieceCount
	if pp.PieceCount() != 32 {
		t.Errorf("PieceCount: got %d, want 32", pp.PieceCount())
	}

	// Test PieceAt
	if pp.PieceAt(SqE1) != 'K' {
		t.Errorf("PieceAt(e1): got %c, want K", pp.PieceAt(SqE1))
	}
	if pp.PieceAt(SqE8) != 'k' {
		t.Errorf("PieceAt(e8): got %c, want k", pp.PieceAt(SqE8))
	}
	if pp.PieceAt(SqE4) != 0 {
		t.Errorf("PieceAt(e4): got %c, want empty", pp.PieceAt(SqE4))
	}
}

func TestPackedPosition_Pieces(t *testing.T) {
	gs := NewStartingPosition()
	pp := gs.Pack()

	var count int
	pp.Pieces(func(sq Square, piece byte) {
		count++
		if piece == 0 {
			t.Errorf("Pieces callback received empty piece at %s", sq)
		}
	})

	if count != 32 {
		t.Errorf("Pieces iterated %d times, want 32", count)
	}
}

func TestPackedPositionFromFEN(t *testing.T) {
	encoded, err := PackedPositionFromFEN(startingFEN)
	if err != nil {
		t.Fatalf("PackedPositionFromFEN: %v", err)
	}
	t.Logf("Starting position key: %s", encoded)

	// 26 bytes -> 35 base64 chars
	if len(encoded) != 35 {
		t.Errorf("Expected 35 char base64, got %d", len(encoded))
	}

	pp, err := ParsePackedPosition(encoded)
	if err != nil {
		t.Fatalf("ParsePackedPosition: %v", err)
	}

	// Position should match
	if pp.PieceAt(SqE1) != 'K' {
		t.Errorf("PieceAt(e1): got %c, want K", pp.PieceAt(SqE1))
	}
	if pp.SideToMove() != White {
		t.Errorf("SideToMove: got %v, want White", pp.SideToMove())
	}
	if pp.CastlingRights() != 0x0F {
		t.Errorf("CastlingRights: got %d, want 15", pp.CastlingRights())
	}
}

func TestPackedPosition_AllPieceTypes(t *testing.T) {
	gs, _ := NewGame(startingFEN)
	pp := gs.Pack()

	expectedPieces := map[Square]byte{
		SqA1: 'R', SqB1: 'N', SqC1: 'B', SqD1: 'Q', SqE1: 'K', SqF1: 'B', SqG1: 'N', SqH1: 'R',
		SqA2: 'P', SqB2: 'P', SqC2: 'P', SqD2: 'P', SqE2: 'P', SqF2: 'P', SqG2: 'P', SqH2: 'P',
		SqA7: 'p', SqB7: 'p', SqC7: 'p', SqD7: 'p', SqE7: 'p', SqF7: 'p', SqG7: 'p', SqH7: 'p',
		SqA8: 'r', SqB8: 'n', SqC8: 'b', SqD8: 'q', SqE8: 'k', SqF8: 'b', SqG8: 'n', SqH8: 'r',
	}

	for sq, expected := range expectedPieces {
		got := pp.PieceAt(sq)
		if got != expected {
			t.Errorf("PieceAt(%s): got %c, want %c", sq, got, expected)
		}
	}

	emptySquares := []Square{SqA3, SqA4, SqA5, SqA6, SqD4, SqE4, SqE5}
	for _, sq := range emptySquares {
		got := pp.PieceAt(sq)
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
		{"wrong length 25 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},  // 34 chars = 25 bytes
		{"wrong length 27 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, // 36 chars = 27 bytes
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

func TestPackedPosition_Metadata(t *testing.T) {
	tests := []struct {
		name      string
		fen       string
		sideToMove Color
		castle    uint8
		epFile    int
	}{
		{"starting", startingFEN, White, 0x0F, -1},
		{"black to move with EP", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", Black, 0x0F, 4},
		{"white kingside only", "r3k2r/8/8/8/8/8/8/R3K2R w K - 0 1", White, 0x01, -1},
		{"no castling", "r3k2r/8/8/8/8/8/8/R3K2R b - - 0 1", Black, 0x00, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs, _ := NewGame(tt.fen)
			pp := gs.Pack()

			if pp.SideToMove() != tt.sideToMove {
				t.Errorf("SideToMove: got %v, want %v", pp.SideToMove(), tt.sideToMove)
			}
			if pp.CastlingRights() != tt.castle {
				t.Errorf("CastlingRights: got %d, want %d", pp.CastlingRights(), tt.castle)
			}
			if pp.EPFile() != tt.epFile {
				t.Errorf("EPFile: got %d, want %d", pp.EPFile(), tt.epFile)
			}
		})
	}
}

func TestPackedPosition_Setters(t *testing.T) {
	gs := NewStartingPosition()
	pp := gs.Pack()

	// Test SetSideToMove
	pp.SetSideToMove(Black)
	if pp.SideToMove() != Black {
		t.Errorf("SetSideToMove: got %v, want Black", pp.SideToMove())
	}

	// Test SetCastlingRights
	pp.SetCastlingRights(0x05) // Kk
	if pp.CastlingRights() != 0x05 {
		t.Errorf("SetCastlingRights: got %d, want 5", pp.CastlingRights())
	}

	// Test SetEPFile
	pp.SetEPFile(3) // d-file
	if pp.EPFile() != 3 {
		t.Errorf("SetEPFile: got %d, want 3", pp.EPFile())
	}

	// Test SetEPFile(-1) clears
	pp.SetEPFile(-1)
	if pp.EPFile() != -1 {
		t.Errorf("SetEPFile(-1): got %d, want -1", pp.EPFile())
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

func BenchmarkPackedPosition_PieceAt(b *testing.B) {
	gs := NewStartingPosition()
	pp := gs.Pack()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pp.PieceAt(Square(i % 64))
	}
}

// -----------------------------------------------------------------------------
// PackedFEN tests (includes metadata)
// -----------------------------------------------------------------------------

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

			// 29 bytes -> 39 base64 chars
			if len(base64Str) != 39 {
				t.Errorf("Expected 39 char base64, got %d", len(base64Str))
			}

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

func TestPackedFEN_Getters(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 5 10"
	gs, _ := NewGame(fen)
	pf := gs.PackFEN()

	// Test board getters
	if pf.PieceAt(SqE1) != 'K' {
		t.Errorf("PieceAt(e1): got %c, want K", pf.PieceAt(SqE1))
	}
	if pf.PieceCount() != 32 {
		t.Errorf("PieceCount: got %d, want 32", pf.PieceCount())
	}

	// Test metadata getters
	if pf.SideToMove() != Black {
		t.Errorf("SideToMove: got %v, want Black", pf.SideToMove())
	}
	if pf.CastlingRights() != 0x0F {
		t.Errorf("CastlingRights: got %d, want 15", pf.CastlingRights())
	}
	if pf.EPFile() != 4 {
		t.Errorf("EPFile: got %d, want 4", pf.EPFile())
	}
	if pf.EPSquare() != SqE3 {
		t.Errorf("EPSquare: got %d, want %d (e3)", pf.EPSquare(), SqE3)
	}
	if pf.Halfmove() != 5 {
		t.Errorf("Halfmove: got %d, want 5", pf.Halfmove())
	}
	if pf.Fullmove() != 10 {
		t.Errorf("Fullmove: got %d, want 10", pf.Fullmove())
	}
}

func TestPackedFEN_Setters(t *testing.T) {
	gs := NewStartingPosition()
	pf := gs.PackFEN()

	// Test SetSideToMove
	pf.SetSideToMove(Black)
	if pf.SideToMove() != Black {
		t.Errorf("SetSideToMove: got %v, want Black", pf.SideToMove())
	}

	// Test SetCastlingRights
	pf.SetCastlingRights(0x05) // Kk
	if pf.CastlingRights() != 0x05 {
		t.Errorf("SetCastlingRights: got %d, want 5", pf.CastlingRights())
	}

	// Test SetEPFile
	pf.SetEPFile(3) // d-file
	if pf.EPFile() != 3 {
		t.Errorf("SetEPFile: got %d, want 3", pf.EPFile())
	}

	// Test SetEPFile(-1) clears
	pf.SetEPFile(-1)
	if pf.EPFile() != -1 {
		t.Errorf("SetEPFile(-1): got %d, want -1", pf.EPFile())
	}

	// Test SetHalfmove
	pf.SetHalfmove(50)
	if pf.Halfmove() != 50 {
		t.Errorf("SetHalfmove: got %d, want 50", pf.Halfmove())
	}

	// Test SetFullmove
	pf.SetFullmove(100)
	if pf.Fullmove() != 100 {
		t.Errorf("SetFullmove: got %d, want 100", pf.Fullmove())
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

func TestPackedFEN_EnPassant(t *testing.T) {
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
			pf := gs.PackFEN()
			decoded := pf.Unpack()

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

func TestPackedFEN_NoEnPassant(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	gs, _ := NewGame(fen)
	pf := gs.PackFEN()

	if pf.EPFile() != -1 {
		t.Errorf("EPFile should be -1, got %d", pf.EPFile())
	}

	decoded := pf.Unpack()
	if decoded.EP >= 0 {
		t.Errorf("EP should be -1, got %d", decoded.EP)
	}
}

func TestPackedFEN_CastlingCombinations(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		castle uint8
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
			pf := gs.PackFEN()
			decoded := pf.Unpack()

			if decoded.Castle != tt.castle {
				t.Errorf("Castle: got %d, want %d", decoded.Castle, tt.castle)
			}
		})
	}
}

func TestPackFEN(t *testing.T) {
	pf, err := PackFEN(startingFEN)
	if err != nil {
		t.Fatalf("PackFEN: %v", err)
	}

	gs := pf.Unpack()
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

func TestParsePackedFEN_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "abc"},
		{"invalid base64", "!!!invalid!!!"},
		{"wrong length 28 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},   // 37 chars = 27 bytes
		{"wrong length 30 bytes", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}, // 40 chars = 30 bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePackedFEN(tt.input)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func BenchmarkPackedFEN_PackFEN(b *testing.B) {
	gs := NewStartingPosition()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gs.PackFEN()
	}
}

func BenchmarkPackedFEN_Unpack(b *testing.B) {
	gs := NewStartingPosition()
	pf := gs.PackFEN()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pf.Unpack()
	}
}

func BenchmarkPackedFEN_String(b *testing.B) {
	gs := NewStartingPosition()
	pf := gs.PackFEN()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pf.String()
	}
}

func BenchmarkParsePackedFEN(b *testing.B) {
	gs := NewStartingPosition()
	pf := gs.PackFEN()
	s := pf.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePackedFEN(s)
	}
}
