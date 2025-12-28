package pgn

import "testing"

// Standard perft test suite from Chess Programming Wiki and other sources.
// These positions are well-established with known correct node counts.

// perftCase defines a perft test case.
type perftCase struct {
	name  string
	fen   string
	depth int
	nodes int64
}

// Standard perft positions (Chess Programming Wiki):
// Position 1: Starting position
// Position 2: Kiwipete (Peter McKenzie)
// Position 3: Endgame with promotions
// Position 4: Many promotions and checks
// Position 5: Promotion-heavy middlegame
// Position 6: Mirrored position

var perftCases = []perftCase{
	// Starting position
	{"startpos_d1", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1, 20},
	{"startpos_d2", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 2, 400},
	{"startpos_d3", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 3, 8902},
	{"startpos_d4", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 4, 197281},

	// Kiwipete by Peter McKenzie - tests castling, en passant, promotions
	{"kiwipete_d1", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 1, 48},
	{"kiwipete_d2", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 2, 2039},
	{"kiwipete_d3", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 3, 97862},

	// Position 3 - Endgame with rook and pawns
	{"pos3_d1", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 1, 14},
	{"pos3_d2", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 2, 191},
	{"pos3_d3", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 3, 2812},
	{"pos3_d4", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 4, 43238},

	// Position 4 - Many promotions, discovered checks
	{"pos4_d1", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 1, 6},
	{"pos4_d2", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 2, 264},
	{"pos4_d3", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 3, 9467},

	// Position 5 - Tests promotion with check
	{"pos5_d1", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 1, 44},
	{"pos5_d2", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 2, 1486},
	{"pos5_d3", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 3, 62379},

	// Position 6 - Symmetric/mirrored position by Steven Edwards
	{"pos6_d1", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 1, 46},
	{"pos6_d2", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 2, 2079},
	{"pos6_d3", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 3, 89890},

	// Additional edge cases

	// En passant discovered check - EP is illegal because it exposes king to queen
	{"ep_check_d1", "8/8/8/8/k2Pp2Q/8/8/3K4 b - d3 0 1", 1, 6},
	{"ep_check_d2", "8/8/8/8/k2Pp2Q/8/8/3K4 b - d3 0 1", 2, 136},

	// Promotion with all 4 piece types
	{"promo_all_d1", "n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1", 1, 24},
	{"promo_all_d2", "n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1", 2, 496},

	// Scholar's mate pattern (after 1.e4 e5 2.Bc4 Nc6 3.Qh5 Nf6)
	{"scholar_d1", "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4", 1, 43},

	// Castling with both sides having rights, open position
	{"castle_open_d1", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", 1, 26},
	{"castle_open_d2", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", 2, 568},
	{"castle_open_d3", "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", 3, 13744},
}

func TestPerft_Suite(t *testing.T) {
	for _, tc := range perftCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip deeper tests in short mode
			if testing.Short() && tc.depth > 3 {
				t.Skip("skipping deep perft in short mode")
			}

			pos, err := NewGame(tc.fen)
			if err != nil {
				t.Fatalf("fen parse error: %v", err)
			}
			got := Perft(pos, tc.depth)
			if got != tc.nodes {
				t.Errorf("%s depth %d = %d, want %d", tc.name, tc.depth, got, tc.nodes)
			}
		})
	}
}

// Helpers to list legal moves for debugging.
func listMoves(pos *GameState) []Mv {
	return GenerateLegalMoves(pos)
}

// BenchmarkPerft benchmarks perft at various depths.
// To see nodes/second: go test -bench=BenchmarkPerft -benchtime=3s

func BenchmarkPerft_Startpos_D5(b *testing.B) {
	pos, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Perft(pos, 5)
	}
	b.ReportMetric(float64(4865609*b.N)/b.Elapsed().Seconds(), "nodes/sec")
}

func BenchmarkPerft_Startpos_D6(b *testing.B) {
	pos, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Perft(pos, 6)
	}
	b.ReportMetric(float64(119060324*b.N)/b.Elapsed().Seconds(), "nodes/sec")
}

func BenchmarkPerft_Kiwipete_D4(b *testing.B) {
	pos, _ := NewGame("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Perft(pos, 4)
	}
	b.ReportMetric(float64(4085603*b.N)/b.Elapsed().Seconds(), "nodes/sec")
}

func BenchmarkPerft_Kiwipete_D5(b *testing.B) {
	pos, _ := NewGame("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Perft(pos, 5)
	}
	b.ReportMetric(float64(193690690*b.N)/b.Elapsed().Seconds(), "nodes/sec")
}

// BenchmarkMoveGen benchmarks just move generation (no recursion)
func BenchmarkMoveGen_Startpos(b *testing.B) {
	pos, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateLegalMoves(pos)
	}
}

func BenchmarkMoveGen_Kiwipete(b *testing.B) {
	pos, _ := NewGame("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateLegalMoves(pos)
	}
}

func BenchmarkMoveGen_Middlegame(b *testing.B) {
	// Complex middlegame position
	pos, _ := NewGame("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateLegalMoves(pos)
	}
}
