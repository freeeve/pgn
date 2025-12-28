package pgn_test

import (
	"testing"

	pgn "github.com/freeeve/pgn/v2"
)

// A modest Ruy Lopez line to exercise ParseSAN + MakeMove.
var sanLine = []string{
	"e4", "e5",
	"Nf3", "Nc6",
	"Bb5", "a6",
	"Ba4", "Nf6",
	"O-O", "Be7",
	"Re1", "b5",
	"Bb3", "d6",
	"c3", "O-O",
	"h3", "Nb8",
	"d4", "Nbd7",
}

func BenchmarkMakeMovesRuyLopez(b *testing.B) {
	startPos := pgn.NewStartingPosition()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gs := startPos.Copy()
		for _, san := range sanLine {
			move, err := pgn.ParseSAN(gs, san)
			if err != nil {
				b.Fatalf("move %s: %v", san, err)
			}
			pgn.MakeMove(gs, move)
		}
	}
}
