package pgn_test

import (
	"testing"

	pgn "github.com/freeeve/pgn/v2"
)

func BenchmarkPerft_Startpos_D6(b *testing.B) {
	pos, err := pgn.NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		b.Fatalf("parse FEN: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pgn.Perft(pos, 6)
	}
	b.ReportMetric(float64(119060324*b.N)/b.Elapsed().Seconds(), "nodes/sec")
}
