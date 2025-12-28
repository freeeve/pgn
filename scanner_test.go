package pgn

import (
	"bytes"
	"strings"
	"testing"
)

func TestPGNScanner_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		pgn       string
		wantGames int
		wantMoves int // moves in first game
	}{
		{
			name: "with_NAG",
			pgn: `[Event "Test"]
[Result "1-0"]

1. e4 $1 e5 $2 2. Nf3 $14 1-0
`,
			wantGames: 1,
			wantMoves: 3,
		},
		{
			name: "with_annotations",
			pgn: `[Event "Test"]
[Result "1-0"]

1. e4! e5? 2. Nf3!! Nc6?? 1-0
`,
			wantGames: 1,
			wantMoves: 4,
		},
		{
			name: "with_nested_variations",
			pgn: `[Event "Test"]
[Result "1-0"]

1. e4 (1. d4 d5) e5 (1... c5) 2. Nf3 1-0
`,
			wantGames: 1,
			wantMoves: 3,
		},
		{
			name: "star_result",
			pgn: `[Event "Test"]
[Result "*"]

1. e4 e5 *
`,
			wantGames: 1,
			wantMoves: 2,
		},
		{
			name: "draw_result",
			pgn: `[Event "Test"]
[Result "1/2-1/2"]

1. e4 e5 1/2-1/2
`,
			wantGames: 1,
			wantMoves: 2,
		},
		{
			name: "with_FEN",
			pgn: `[Event "Test"]
[FEN "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"]
[Result "1-0"]

1... e5 2. Nf3 1-0
`,
			wantGames: 1,
			wantMoves: 2,
		},
		{
			name: "castling_moves",
			pgn: `[Event "Test"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 6. Re1 b5 7. Bb3 d6 8. c3 O-O 1-0
`,
			wantGames: 1,
			wantMoves: 16,
		},
		{
			name: "promotion",
			pgn: `[Event "Test"]
[FEN "8/P7/8/8/8/8/8/4K2k w - - 0 1"]
[Result "1-0"]

1. a8=Q 1-0
`,
			wantGames: 1,
			wantMoves: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPGNScanner(strings.NewReader(tt.pgn))
			var gameCount int
			var firstGameMoves int
			for ps.Next() {
				game, err := ps.Scan()
				if err != nil {
					t.Fatalf("parse error: %v", err)
				}
				if len(game.Tags) > 0 || len(game.Moves) > 0 {
					if gameCount == 0 {
						firstGameMoves = len(game.Moves)
					}
					gameCount++
				}
			}

			if gameCount != tt.wantGames {
				t.Errorf("games: got %d, want %d", gameCount, tt.wantGames)
			}
			if tt.wantGames > 0 && firstGameMoves != tt.wantMoves {
				t.Errorf("moves: got %d, want %d", firstGameMoves, tt.wantMoves)
			}
		})
	}
}

func TestFastPGNScanner_Basic(t *testing.T) {
	pgnData := `[Event "Test"]
[White "Player1"]
[Black "Player2"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 1-0

[Event "Test2"]
[White "Player3"]
[Black "Player4"]
[Result "0-1"]

1. d4 d5 2. c4 dxc4 0-1
`

	ps := NewPGNScanner(bytes.NewReader([]byte(pgnData)))

	// First game
	if !ps.Next() {
		t.Fatal("expected first game")
	}
	game1, err := ps.Scan()
	if err != nil {
		t.Fatalf("failed to scan first game: %v", err)
	}
	if game1.Tags["Event"] != "Test" {
		t.Errorf("expected Event=Test, got %q", game1.Tags["Event"])
	}
	if len(game1.Moves) != 5 {
		t.Errorf("expected 5 moves, got %d", len(game1.Moves))
	}

	// Second game
	if !ps.Next() {
		t.Fatal("expected second game")
	}
	game2, err := ps.Scan()
	if err != nil {
		t.Fatalf("failed to scan second game: %v", err)
	}
	if game2.Tags["Event"] != "Test2" {
		t.Errorf("expected Event=Test2, got %q", game2.Tags["Event"])
	}
	if len(game2.Moves) != 4 {
		t.Errorf("expected 4 moves (d4 d5 c4 dxc4), got %d", len(game2.Moves))
	}
}

// TestParallelMatchesSequential verifies parallel parsing produces identical results.
func TestParallelMatchesSequential(t *testing.T) {
	pgnData := `[Event "Game1"]
[White "A"]
[Black "B"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O 1-0

[Event "Game2"]
[White "C"]
[Black "D"]
[Result "0-1"]

1. d4 d5 2. c4 e6 3. Nc3 Nf6 4. Bg5 Be7 5. e3 0-1

[Event "Game3"]
[White "E"]
[Black "F"]
[Result "1/2-1/2"]

1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 a6 1/2-1/2

[Event "Game4"]
[White "G"]
[Black "H"]
[Result "*"]

1. e4 e6 2. d4 d5 3. Nc3 Bb4 4. e5 c5 5. a3 Bxc3+ *
`

	// Parse sequentially
	ps := newStreamingScanner(bytes.NewReader([]byte(pgnData)))
	var seqGames []*Game
	for ps.Next() {
		game, err := ps.Scan()
		if err != nil {
			break
		}
		// Skip empty games
		if len(game.Tags) == 0 && len(game.Moves) == 0 {
			continue
		}
		// Copy since scanner reuses memory
		gameCopy := &Game{
			Tags:  make(map[string]string, len(game.Tags)),
			Moves: make([]Mv, len(game.Moves)),
		}
		for k, v := range game.Tags {
			gameCopy.Tags[k] = v
		}
		copy(gameCopy.Moves, game.Moves)
		seqGames = append(seqGames, gameCopy)
	}

	// Parse in parallel
	var parGames []*Game
	parser := GamesFromReader(bytes.NewReader([]byte(pgnData)), 4)
	for game := range parser.Games {
		parGames = append(parGames, game)
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("parallel parse error: %v", err)
	}

	// Compare counts
	if len(seqGames) != len(parGames) {
		t.Fatalf("game count mismatch: sequential=%d, parallel=%d", len(seqGames), len(parGames))
	}

	// Compare each game
	for i := range seqGames {
		seq := seqGames[i]
		par := parGames[i]

		// Compare tags
		if seq.Tags["Event"] != par.Tags["Event"] {
			t.Errorf("game %d: Event mismatch: %q vs %q", i, seq.Tags["Event"], par.Tags["Event"])
		}
		if seq.Tags["White"] != par.Tags["White"] {
			t.Errorf("game %d: White mismatch: %q vs %q", i, seq.Tags["White"], par.Tags["White"])
		}
		if seq.Tags["Result"] != par.Tags["Result"] {
			t.Errorf("game %d: Result mismatch: %q vs %q", i, seq.Tags["Result"], par.Tags["Result"])
		}

		// Compare move counts
		if len(seq.Moves) != len(par.Moves) {
			t.Errorf("game %d: move count mismatch: %d vs %d", i, len(seq.Moves), len(par.Moves))
			continue
		}

		// Compare each move
		for j := range seq.Moves {
			if seq.Moves[j] != par.Moves[j] {
				t.Errorf("game %d, move %d: mismatch: %+v vs %+v", i, j, seq.Moves[j], par.Moves[j])
			}
		}
	}
}
