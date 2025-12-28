package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/freeeve/pgn/v2"
)

func main() {
	path := "lichess_db_standard_rated_2018-01.pgn.zst"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fmt.Printf("Parsing %s with %d workers...\n", path, runtime.NumCPU()-1)
	start := time.Now()

	var count int
	var totalMoves int
	parser := pgn.Games(path)

	for game := range parser.Games {
		count++
		totalMoves += len(game.Moves)
		if count%1000000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("  %dM games, %dM moves, %.1f games/sec\n",
				count/1000000, totalMoves/1000000, float64(count)/elapsed.Seconds())
		}
	}

	if err := parser.Err(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("\nDone!\n")
	fmt.Printf("  Games: %d\n", count)
	fmt.Printf("  Moves: %d\n", totalMoves)
	fmt.Printf("  Time: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Speed: %.0f games/sec\n", float64(count)/elapsed.Seconds())
}
