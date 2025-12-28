package pgn

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestStreamingScanner_Basic(t *testing.T) {
	pgn := `[Event "Test"]
[White "Player1"]
[Black "Player2"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 1-0

[Event "Test2"]
[Result "*"]

1. d4 d5 *
`
	scanner := newStreamingScanner(strings.NewReader(pgn))

	var games []*Game
	for scanner.Next() {
		game, err := scanner.Scan()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		gameCopy := &Game{
			Tags:  make(map[string]string, len(game.Tags)),
			Moves: make([]Mv, len(game.Moves)),
		}
		for k, v := range game.Tags {
			gameCopy.Tags[k] = v
		}
		copy(gameCopy.Moves, game.Moves)
		games = append(games, gameCopy)
	}

	if len(games) != 2 {
		t.Errorf("Expected 2 games, got %d", len(games))
	}

	if games[0].Tags["Event"] != "Test" {
		t.Errorf("Game 1 Event: got %q", games[0].Tags["Event"])
	}
	if len(games[0].Moves) != 5 {
		t.Errorf("Game 1 moves: got %d, want 5", len(games[0].Moves))
	}

	if games[1].Tags["Event"] != "Test2" {
		t.Errorf("Game 2 Event: got %q", games[1].Tags["Event"])
	}
	if len(games[1].Moves) != 2 {
		t.Errorf("Game 2 moves: got %d, want 2", len(games[1].Moves))
	}
}

func TestStreamingScanner_MatchesPGNScanner(t *testing.T) {
	pgn := `[Event "Test"]
[Site "Internet"]
[Date "2024.01.01"]
[White "Alice"]
[Black "Bob"]
[Result "1/2-1/2"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 1/2-1/2
`
	streamScanner := newStreamingScanner(strings.NewReader(pgn))
	bufScanner := NewPGNScanner(strings.NewReader(pgn))

	if !streamScanner.Next() || !bufScanner.Next() {
		t.Fatal("No games found")
	}

	streamGame, err1 := streamScanner.Scan()
	bufGame, err2 := bufScanner.Scan()

	if err1 != nil || err2 != nil {
		t.Fatalf("Scan errors: stream=%v, buf=%v", err1, err2)
	}

	if len(streamGame.Tags) != len(bufGame.Tags) {
		t.Errorf("Tag count: stream=%d, buf=%d", len(streamGame.Tags), len(bufGame.Tags))
	}

	for k, v := range bufGame.Tags {
		if streamGame.Tags[k] != v {
			t.Errorf("Tag %s: stream=%q, buf=%q", k, streamGame.Tags[k], v)
		}
	}

	if len(streamGame.Moves) != len(bufGame.Moves) {
		t.Errorf("Move count: stream=%d, buf=%d", len(streamGame.Moves), len(bufGame.Moves))
	}

	for i := range bufGame.Moves {
		if streamGame.Moves[i] != bufGame.Moves[i] {
			t.Errorf("Move %d: stream=%v, buf=%v", i, streamGame.Moves[i], bufGame.Moves[i])
		}
	}
}

func TestStreamingScanner_LargeBuffer(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		buf.WriteString(`[Event "Game"]
[Result "*"]

1. e4 e5 2. Nf3 *

`)
	}

	scanner := newStreamingScanner(&buf)
	count := 0
	for scanner.Next() {
		_, err := scanner.Scan()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Scan error at game %d: %v", count, err)
		}
		count++
	}

	if count != 100 {
		t.Errorf("Expected 100 games, got %d", count)
	}
}

func BenchmarkStreamingScanner(b *testing.B) {
	pgn := `[Event "Test"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 1-0

`
	data := strings.Repeat(pgn, 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		scanner := newStreamingScanner(strings.NewReader(data))
		count := 0
		for scanner.Next() {
			_, err := scanner.Scan()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			count++
		}
		if count != 1000 {
			b.Fatalf("Expected 1000 games, got %d", count)
		}
	}
}

func TestGamesFromReader(t *testing.T) {
	pgn := `[Event "Game1"]
[Result "1-0"]

1. e4 e5 2. Nf3 1-0

[Event "Game2"]
[Result "0-1"]

1. d4 d5 0-1

[Event "Game3"]
[Result "*"]

1. c4 *
`
	var games []*Game
	parser := GamesFromReader(strings.NewReader(pgn), 2)
	for game := range parser.Games {
		games = append(games, game)
	}

	if err := parser.Err(); err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(games) != 3 {
		t.Errorf("Expected 3 games, got %d", len(games))
	}

	expectedEvents := []string{"Game1", "Game2", "Game3"}
	for i, expected := range expectedEvents {
		if i >= len(games) {
			break
		}
		if games[i].Tags["Event"] != expected {
			t.Errorf("Game %d Event: got %q, want %q", i, games[i].Tags["Event"], expected)
		}
	}
}

func TestGamesFromReader_Order(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		buf.WriteString(fmt.Sprintf(`[Event "Game%03d"]
[Result "*"]

1. e4 *

`, i))
	}

	var events []string
	parser := GamesFromReader(&buf, 4)
	for game := range parser.Games {
		events = append(events, game.Tags["Event"])
	}

	if err := parser.Err(); err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(events) != 100 {
		t.Errorf("Expected 100 games, got %d", len(events))
	}

	for i, event := range events {
		expected := fmt.Sprintf("Game%03d", i)
		if event != expected {
			t.Errorf("Game %d: got %q, want %q", i, event, expected)
		}
	}
}

func TestGamesFromReader_Stop(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString(fmt.Sprintf(`[Event "Game%03d"]
[Result "*"]

1. e4 *

`, i))
	}

	count := 0
	parser := GamesFromReader(&buf, 4)
	for range parser.Games {
		count++
		if count >= 10 {
			parser.Stop()
			break
		}
	}

	if count < 10 {
		t.Errorf("Expected at least 10 games before stop, got %d", count)
	}
}

