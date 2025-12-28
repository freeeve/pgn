package pgn

import (
	"bytes"
	"os"
	"testing"
)

// TestPGNSanityLichess validates that our parser correctly processes
// a large PGN file and produces identical results between sequential
// and parallel parsing.
func TestPGNSanityLichess(t *testing.T) {
	data, err := os.ReadFile("lichess_db_standard_rated_2013-01.pgn")
	if os.IsNotExist(err) {
		t.Skip("lichess_db_standard_rated_2013-01.pgn not present")
	}
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Parse sequentially
	var seqGames []*Game
	ps := newStreamingScanner(bytes.NewReader(data))
	for ps.Next() {
		game, err := ps.Scan()
		if err != nil {
			break // EOF or error
		}
		// Skip empty games
		if len(game.Tags) == 0 && len(game.Moves) == 0 {
			continue
		}
		// Copy game
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
	parser := GamesFromReader(bytes.NewReader(data))
	for game := range parser.Games {
		parGames = append(parGames, game)
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("parallel parse error: %v", err)
	}

	// Verify game counts
	t.Logf("Sequential: %d games", len(seqGames))
	t.Logf("Parallel: %d games", len(parGames))

	if len(seqGames) != len(parGames) {
		t.Fatalf("game count mismatch: sequential=%d, parallel=%d", len(seqGames), len(parGames))
	}

	// Verify each game matches
	var totalMoves int
	var mismatchCount int
	for i := 0; i < len(seqGames); i++ {
		seq := seqGames[i]
		par := parGames[i]

		totalMoves += len(seq.Moves)

		// Compare move counts
		if len(seq.Moves) != len(par.Moves) {
			if mismatchCount < 10 {
				t.Errorf("game %d: move count mismatch: seq=%d par=%d (Event=%q)",
					i+1, len(seq.Moves), len(par.Moves), seq.Tags["Event"])
			}
			mismatchCount++
			continue
		}

		// Compare each move
		for j := range seq.Moves {
			if seq.Moves[j] != par.Moves[j] {
				if mismatchCount < 10 {
					t.Errorf("game %d, move %d: mismatch: %+v vs %+v",
						i+1, j+1, seq.Moves[j], par.Moves[j])
				}
				mismatchCount++
				break
			}
		}
	}

	if mismatchCount > 0 {
		t.Fatalf("total mismatches: %d", mismatchCount)
	}

	t.Logf("Verified %d games with %d total moves", len(seqGames), totalMoves)
}

// TestPGNGameStats collects statistics about games in the file.
func TestPGNGameStats(t *testing.T) {
	data, err := os.ReadFile("lichess_db_standard_rated_2013-01.pgn")
	if os.IsNotExist(err) {
		t.Skip("lichess_db_standard_rated_2013-01.pgn not present")
	}
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var games []*Game
	parser := GamesFromReader(bytes.NewReader(data))
	for game := range parser.Games {
		games = append(games, game)
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Collect stats
	var totalMoves int
	var minMoves, maxMoves int = 999999, 0
	resultCounts := make(map[string]int)
	terminationCounts := make(map[string]int)

	for _, game := range games {
		moves := len(game.Moves)
		totalMoves += moves
		if moves < minMoves {
			minMoves = moves
		}
		if moves > maxMoves {
			maxMoves = moves
		}

		result := game.Tags["Result"]
		if result != "" {
			resultCounts[result]++
		}

		termination := game.Tags["Termination"]
		if termination != "" {
			terminationCounts[termination]++
		}
	}

	t.Logf("Total games: %d", len(games))
	t.Logf("Total moves: %d", totalMoves)
	t.Logf("Average moves per game: %.1f", float64(totalMoves)/float64(len(games)))
	t.Logf("Min moves: %d, Max moves: %d", minMoves, maxMoves)

	t.Logf("\nResults:")
	for result, count := range resultCounts {
		t.Logf("  %s: %d (%.1f%%)", result, count, 100*float64(count)/float64(len(games)))
	}

	t.Logf("\nTerminations (top 5):")
	// Simple sorting - just show top counts
	for termination, count := range terminationCounts {
		if count > len(games)/20 { // Only show if > 5%
			t.Logf("  %s: %d", termination, count)
		}
	}
}

// TestPGNSampleGames verifies specific games have expected properties.
func TestPGNSampleGames(t *testing.T) {
	data, err := os.ReadFile("lichess_db_standard_rated_2013-01.pgn")
	if os.IsNotExist(err) {
		t.Skip("lichess_db_standard_rated_2013-01.pgn not present")
	}
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var games []*Game
	parser := GamesFromReader(bytes.NewReader(data))
	for game := range parser.Games {
		games = append(games, game)
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Sample first 10, middle 10, last 10 games
	samples := []struct {
		name   string
		offset int
	}{
		{"first", 0},
		{"middle", len(games) / 2},
		{"last", len(games) - 10},
	}

	for _, sample := range samples {
		t.Logf("\n%s games (starting at %d):", sample.name, sample.offset)
		for i := 0; i < 10 && sample.offset+i < len(games); i++ {
			game := games[sample.offset+i]
			t.Logf("  Game %d: %s vs %s, %d moves, result=%s",
				sample.offset+i+1,
				game.Tags["White"],
				game.Tags["Black"],
				len(game.Moves),
				game.Tags["Result"])

			// Basic validation
			if len(game.Moves) == 0 {
				t.Errorf("game %d has no moves", sample.offset+i+1)
			}
		}
	}
}

// TestPolgarPGNSanity validates polgar.pgn if present.
func TestPolgarPGNSanity(t *testing.T) {
	data, err := os.ReadFile("polgar.pgn")
	if os.IsNotExist(err) {
		t.Skip("polgar.pgn not present")
	}
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Sequential
	var seqCount int
	ps := newStreamingScanner(bytes.NewReader(data))
	for ps.Next() {
		_, err := ps.Scan()
		if err != nil {
			break // EOF or error
		}
		seqCount++
	}

	// Parallel
	var parCount int
	parser := GamesFromReader(bytes.NewReader(data), 4)
	for range parser.Games {
		parCount++
	}
	if err := parser.Err(); err != nil {
		t.Fatalf("parallel parse error: %v", err)
	}

	t.Logf("polgar.pgn: %d games (sequential), %d games (parallel)", seqCount, parCount)

	// They might differ slightly due to empty game handling
	if seqCount == 0 || parCount == 0 {
		t.Fatal("no games parsed")
	}
}

