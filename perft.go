// perft.go provides perft (performance test) functions for move generation.
//
// Perft counts all leaf nodes at a given depth from a position, useful for
// validating move generation correctness and measuring performance.

package pgn

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Perft counts all legal move sequences to a given depth from a position.
// It uses parallel processing and pooled allocations for best performance.
// Returns the number of leaf nodes (positions) at the given depth.
func Perft(pos *GameState, depth int) int64 {
	if depth <= 0 {
		return 1
	}
	moves := GenerateLegalMoves(pos)
	if depth == 1 {
		return int64(len(moves))
	}

	// For shallow depths or few moves, use single-threaded pooled version
	if depth <= 3 || len(moves) < 4 {
		return perftPooled(pos, depth)
	}

	// Parallel search at root level
	var total int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

	for _, mv := range moves {
		wg.Add(1)
		sem <- struct{}{}

		go func(m Mv) {
			defer wg.Done()
			defer func() { <-sem }()

			localPos := *pos
			undo := MakeMove(&localPos, m)
			count := perftPooled(&localPos, depth-1)
			UnmakeMove(&localPos, m, undo)

			atomic.AddInt64(&total, count)
		}(mv)
	}

	wg.Wait()
	return total
}

// perftPooled uses pooled move lists to reduce allocations.
func perftPooled(pos *GameState, depth int) int64 {
	if depth <= 0 {
		return 1
	}
	moves := generateMovesPooled(pos)
	if depth == 1 {
		count := int64(len(*moves))
		releaseMoves(moves)
		return count
	}
	var nodes int64
	for _, mv := range *moves {
		undo := MakeMove(pos, mv)
		nodes += perftPooled(pos, depth-1)
		UnmakeMove(pos, mv, undo)
	}
	releaseMoves(moves)
	return nodes
}

