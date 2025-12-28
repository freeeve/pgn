package pgn_test

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	pgn "github.com/freeeve/pgn/v2"
)

// BenchmarkParsePGN benchmarks the parallel PGN parser (default workers).
func BenchmarkParsePGN(b *testing.B) {
	path := os.Getenv("BENCH_PGN_PATH")
	if path == "" {
		path = "lichess_db_standard_rated_2013-01.pgn"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("PGN file not found at %q: %v", path, err)
	}

	var gameCount int
	workers := runtime.NumCPU()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		parser := pgn.GamesFromReader(bytes.NewReader(data), workers)
		for range parser.Games {
			count++
		}
		if err := parser.Err(); err != nil {
			b.Fatalf("parse error: %v", err)
		}
		gameCount = count
	}
	b.ReportMetric(float64(gameCount*b.N)/b.Elapsed().Seconds(), "games/sec")
	b.ReportMetric(float64(len(data)*b.N)/b.Elapsed().Seconds()/1e6, "MB/sec")
}
