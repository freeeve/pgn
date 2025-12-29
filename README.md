# pgn

A high-performance PGN (Portable Game Notation) parser for Go.

**v2.0** - Complete rewrite with bitboard engine, parallel parsing, and zstd support.

[![Go Reference](https://pkg.go.dev/badge/github.com/freeeve/pgn/v2.svg)](https://pkg.go.dev/github.com/freeeve/pgn/v2)

## Installation

```bash
go get github.com/freeeve/pgn/v2
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/freeeve/pgn/v2"
)

func main() {
    // Parse any PGN file (handles .zst compression automatically)
    for game := range pgn.Games("games.pgn").Games {
        fmt.Printf("%s vs %s: %s\n",
            game.Tags["White"],
            game.Tags["Black"],
            game.Tags["Result"])
        
        // Replay the game
        gs := pgn.NewStartingPosition()
        for _, move := range game.Moves {
            pgn.MakeMove(gs, move)
        }
        fmt.Println(gs.ToFEN())
    }
}
```

## Features

- **Fast parallel parsing** - 240+ MB/s, 315K games/sec on Apple M3 Max
- **Streaming** - Parse files of any size with constant memory
- **Zstd support** - Automatic compression/decompression of `.zst` files
- **Bitboard engine** - Efficient move generation and validation
- **FEN support** - Parse and generate FEN strings
- **Packed positions** - 34-byte compact position encoding with base64
- **Position indexing** - Enumerate all positions with deterministic indices
- **Game state queries** - IsCheckmate, IsStalemate, IsSquareAttacked, KingSquare

## API

### Parsing PGN Files

```go
// Parse from file (parallel, handles .zst)
parser := pgn.Games("lichess_games.pgn.zst")
for game := range parser.Games {
    // game.Tags is map[string]string
    // game.Moves is []pgn.Mv
}
if err := parser.Err(); err != nil {
    log.Fatal(err)
}

// Parse from reader
parser := pgn.GamesFromReader(reader)
for game := range parser.Games {
    // ...
}

// Stop early
parser := pgn.Games("huge.pgn")
count := 0
for game := range parser.Games {
    count++
    if count >= 1000 {
        parser.Stop()
        break
    }
}
```

### Working with Positions

```go
// Create starting position
gs := pgn.NewStartingPosition()

// Parse from FEN
gs, err := pgn.NewGame("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")

// Generate FEN
fen := gs.ToFEN()

// Get piece at square
piece := gs.PieceAt(pgn.SqE4) // Returns 'P', 'n', etc. or 0 for empty

// Position status
inCheck := gs.IsInCheck()
isMate := gs.IsCheckmate()
isStalemate := gs.IsStalemate()

// Find king location
whiteKingSq := gs.KingSquare(pgn.White)

// Check if square is attacked
attacked := gs.IsSquareAttacked(pgn.SqE4, pgn.Black)

// Generate legal moves
moves := pgn.GenerateLegalMoves(gs)
```

### Making Moves

```go
gs := pgn.NewStartingPosition()

// Make move (modifies gs in place, returns undo info)
undo := pgn.MakeMove(gs, move)

// Unmake move (restore previous position)
pgn.UnmakeMove(gs, move, undo)

// Parse SAN notation
move, err := pgn.ParseSAN(gs, "e4")
move, err := pgn.ParseSAN(gs, "Nxf7+")
move, err := pgn.ParseSAN(gs, "O-O-O")

// Parse UCI notation
move, err := pgn.ParseUCI("e2e4")
move, err := pgn.ParseUCI("e7e8q") // promotion
```

### Packed Positions

Compact 34-byte encoding for storage/transmission:

```go
// Pack a position
gs := pgn.NewStartingPosition()
packed := gs.Pack()

// Get base64 string (46 chars, URL-safe)
key := packed.String() // "JFM2QhEREREAAAAAAAAAAAAAAAAAAAAAd3d3d4q5nKge_w"

// Parse from base64
packed, err := pgn.ParsePackedPosition(key)

// Unpack to GameState
gs := packed.Unpack()

// Convert to FEN
fen := packed.ToFEN()
```

### Position Indexing

Enumerate all legal chess positions up to a given depth with deterministic indexing:

```go
// Create enumerator from starting position
start := pgn.NewStartingPosition()
enum := pgn.NewPositionEnumeratorDFS(start)

// Enumerate all positions up to depth 5 (~5M positions)
enum.EnumerateDFS(5, func(index uint64, pos *pgn.GameState, depth int) bool {
    // Each position has a unique, deterministic index
    fmt.Printf("Index %d (depth %d): %s\n", index, depth, pos.ToFEN())
    return true // continue enumeration
})

// Save checkpoints for fast position lookup (supports .zstd compression)
enum.SaveCheckpointsCSV("checkpoints_depth5.csv.zstd", 5)

// Load checkpoints and lookup positions by index
enum2 := pgn.NewPositionEnumeratorDFS(start)
enum2.LoadCheckpointsCSV("checkpoints_depth5.csv.zstd")
pos, found := enum2.PositionAtIndexDFS(1000000, 5)

// Find index of a position
idx, found := enum2.IndexOfPositionDFS(somePosition, 5)
```

**Build checkpoint files** with the included tool:

```bash
# Build and run (outputs checkpoints_depth7.csv.zstd by default)
go run ./cmd/build_checkpoints -depth 7

# Custom output, more cores
go run ./cmd/build_checkpoints -depth 8 -cores 16 -output my_checkpoints.csv.zstd
```

## Benchmarks

**Test Machine:** Apple M3 Max (16 cores: 12P + 4E), 128 GB RAM, macOS

### PGN Parsing

Benchmark file: Lichess January 2013 rated games (89 MB, ~121K games)

```
$ go test -bench=BenchmarkParsePGN -benchmem -count=3 -benchtime=5s

goos: darwin
goarch: arm64
pkg: github.com/freeeve/pgn/bench
cpu: Apple M3 Max
BenchmarkParsePGN-16    18    380901891 ns/op    243.7 MB/sec    318539 games/sec
BenchmarkParsePGN-16    15    396615986 ns/op    234.0 MB/sec    305918 games/sec
BenchmarkParsePGN-16    18    383425106 ns/op    242.1 MB/sec    316443 games/sec
```

| Metric | Result |
|--------|--------|
| Throughput | ~240 MB/s |
| Games/sec | ~315K games/sec |
| Parallelism | 16 workers (auto-detected) |

### Perft (Move Generation)

```
$ go test -bench='BenchmarkPerft_Startpos_D6' -benchtime=5s -count=3

BenchmarkPerft_Startpos_D6-16    5    1104828250 ns/op    107763677 nodes/sec
BenchmarkPerft_Startpos_D6-16    5    1020441675 ns/op    116675315 nodes/sec
BenchmarkPerft_Startpos_D6-16    5    1082567617 ns/op    109979604 nodes/sec
```

| Depth | Nodes | Throughput |
|-------|-------|------------|
| 6 | 119,060,324 | **~110M nodes/sec** |

### Move Parsing & Application

20-move Ruy Lopez opening line (ParseSAN + MakeMove):

```
$ go test -bench=BenchmarkMakeMovesRuyLopez -benchmem -count=3

BenchmarkMakeMovesRuyLopez-16    492103    2253 ns/op    0 B/op    0 allocs/op
BenchmarkMakeMovesRuyLopez-16    528484    2179 ns/op    0 B/op    0 allocs/op
BenchmarkMakeMovesRuyLopez-16    534524    2354 ns/op    0 B/op    0 allocs/op
```

| Metric | Result |
|--------|--------|
| 20 moves (parse + apply) | ~2.2 µs |
| Per move | **~110 ns** |
| Allocations | 0 |

## Migration from v1

### Key Changes

| v1 | v2 |
|----|-----|
| `pgn.NewPGNScanner(r)` | `pgn.Games(path)` or `pgn.GamesFromReader(r)` |
| `scanner.Next()` / `scanner.Scan()` | `for game := range parser.Games` |
| `game.Moves` is `[]Move` | `game.Moves` is `[]Mv` |
| `pgn.NewBoard()` | `pgn.NewStartingPosition()` |
| `board.MakeMove(move)` | `pgn.MakeMove(gs, move)` |
| `board.String()` (FEN) | `gs.ToFEN()` |
| `move.From`, `move.To` (Position bitmask) | `move.From`, `move.To` (Square 0-63) |

### Migration Examples

**v1 - Sequential scanning:**
```go
f, _ := os.Open("games.pgn")
ps := pgn.NewPGNScanner(f)
for ps.Next() {
    game, _ := ps.Scan()
    b := pgn.NewBoard()
    for _, move := range game.Moves {
        b.MakeMove(move)
    }
    fmt.Println(b) // FEN
}
```

**v2 - Parallel streaming:**
```go
for game := range pgn.Games("games.pgn").Games {
    gs := pgn.NewStartingPosition()
    for _, move := range game.Moves {
        pgn.MakeMove(gs, move)
    }
    fmt.Println(gs.ToFEN())
}
```

### Square Utilities

```go
// Parse algebraic notation
sq, err := pgn.ParseSquare("e4") // returns SqE4

// Create from file/rank (0-indexed)
sq := pgn.MakeSquare(4, 3) // file=e (4), rank=4 (3) = e4

// Get file and rank
file := sq.File() // 0-7 for a-h
rank := sq.Rank() // 0-7 for 1-8

// String representation
str := sq.String() // "e4"
```

## License

MIT License - see LICENSE file.
