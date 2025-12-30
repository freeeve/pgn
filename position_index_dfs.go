// position_index_dfs.go provides memory-efficient DFS position enumeration.
//
// DFS uses O(depth) memory instead of BFS's O(width), making it vastly more
// efficient for deep enumeration. At depth 40, BFS might use GBs while DFS
// uses only a few KB.

package pgn

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/klauspost/compress/zstd"
)

const (
	CheckpointIntervalDFS = 1 << 20 // 1,048,576

	// Max depth for terminal detection (prevent infinite loops)
	MaxDepthDFS = 500 // ~250 moves, well beyond normal games
)

// CheckpointDFS stores minimal info for DFS restart.
type CheckpointDFS struct {
	Index uint64    // Position index
	Depth int       // Current ply depth
	State GameState // Position state (for quick restart)
	Stack []Mv      // Move stack to reach this position (enables DFS resumption)
}

// positionsEqual checks if two positions are identical (including move counters).
// Used for position-to-index lookup.
func positionsEqual(a, b *GameState) bool {
	if a.SideToMove != b.SideToMove {
		return false
	}
	if a.Castle != b.Castle {
		return false
	}
	if a.EP != b.EP {
		return false
	}
	if a.Halfmove != b.Halfmove {
		return false
	}
	if a.Fullmove != b.Fullmove {
		return false
	}
	for i := 0; i < v2PieceCount; i++ {
		if a.pieces[i] != b.pieces[i] {
			return false
		}
	}
	return true
}

// boardsEqual checks if two positions have the same board state.
// Ignores halfmove and fullmove counters (useful for finding transpositions).
func boardsEqual(a, b *GameState) bool {
	if a.SideToMove != b.SideToMove {
		return false
	}
	if a.Castle != b.Castle {
		return false
	}
	if a.EP != b.EP {
		return false
	}
	for i := 0; i < v2PieceCount; i++ {
		if a.pieces[i] != b.pieces[i] {
			return false
		}
	}
	return true
}

// positionID computes a hash for repetition detection.
// Excludes halfmove/fullmove counters (so same board state = same ID).
func positionID(pos *GameState) uint64 {
	// Simple FNV-1a hash combining all position components
	const offset64 = 14695981039346656037
	const prime64 = 1099511628211

	h := uint64(offset64)

	// Hash all piece bitboards
	for i := 0; i < v2PieceCount; i++ {
		h ^= uint64(pos.pieces[i])
		h *= prime64
	}

	// Hash side to move
	h ^= uint64(pos.SideToMove)
	h *= prime64

	// Hash castling rights
	h ^= uint64(pos.Castle)
	h *= prime64

	// Hash EP square
	h ^= uint64(pos.EP)
	h *= prime64

	// Note: we do NOT hash halfmove or fullmove
	// This makes the ID suitable for repetition detection

	return h
}

// PositionEnumeratorDFS performs depth-first enumeration.
type PositionEnumeratorDFS struct {
	startPos     GameState
	checkpoints  []*CheckpointDFS
	currentIdx   uint64
	currentStack []Mv // Current move stack during enumeration
}

// NewPositionEnumeratorDFS creates a DFS enumerator.
func NewPositionEnumeratorDFS(startPos *GameState) *PositionEnumeratorDFS {
	var start GameState
	if startPos == nil {
		s, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		start = *s
	} else {
		start = *startPos
	}

	return &PositionEnumeratorDFS{
		startPos:    start,
		checkpoints: make([]*CheckpointDFS, 0, 64),
		currentIdx:  0,
	}
}

// EnumerateDFS performs depth-first enumeration with make/unmake.
// This is MUCH more memory efficient than BFS - only O(depth) memory.
//
// The enumeration order is deterministic based on move generation order.
// Terminal conditions:
//   - Checkmate/Stalemate (no legal moves)
//   - 50-move rule (halfmove >= 100)
//   - Threefold repetition (same position 3x on path)
//   - Max depth reached (safety limit)
//   - Callback returns false (early stop)
//
// The callback should return true to continue, false to stop enumeration.
func (e *PositionEnumeratorDFS) EnumerateDFS(
	maxDepth int,
	callback func(index uint64, pos *GameState, depth int) bool,
) {
	if maxDepth <= 0 {
		maxDepth = MaxDepthDFS
	}

	repCount := make(map[uint64]uint8, 100) // Track repetitions on current path
	e.currentIdx = 0
	e.currentStack = make([]Mv, 0, maxDepth) // Initialize move stack

	e.enumerateDFSRecursive(&e.startPos, 0, maxDepth, repCount, callback)
}

// enumerateDFSRecursive is the recursive DFS worker.
// Uses make/unmake pattern for minimal allocations.
// Returns false if enumeration should stop (callback returned false).
func (e *PositionEnumeratorDFS) enumerateDFSRecursive(
	pos *GameState,
	depth int,
	maxDepth int,
	repCount map[uint64]uint8,
	callback func(uint64, *GameState, int) bool,
) bool {
	// Visit this position
	if callback != nil {
		if !callback(e.currentIdx, pos, depth) {
			return false // Early stop requested
		}
	}

	// Store checkpoint if needed
	if e.currentIdx%CheckpointIntervalDFS == 0 {
		// Copy the current move stack for resumption
		stackCopy := make([]Mv, len(e.currentStack))
		copy(stackCopy, e.currentStack)

		ckpt := &CheckpointDFS{
			Index: e.currentIdx,
			Depth: depth,
			State: *pos, // Copy state
			Stack: stackCopy,
		}
		e.checkpoints = append(e.checkpoints, ckpt)
	}

	e.currentIdx++

	// Terminal conditions
	if depth >= maxDepth {
		return true
	}

	// 50-move rule: stop if halfmove >= 100
	if pos.Halfmove >= 100 {
		return true
	}

	// Threefold repetition: only check if we're deep enough for cycles to be possible
	// (need at least 8 plies = 4 moves each side to create a repetition)
	if depth >= 8 {
		posID := positionID(pos)
		if repCount[posID] >= 2 {
			return true // Would be 3rd occurrence = draw
		}

		// Track this position on the path
		repCount[posID]++
		defer func() { repCount[posID]-- }()
	}

	// Generate pseudo-legal moves using pooled buffers (zero allocations)
	pseudo := getMoveList()
	genPseudoMovesTo(pseudo, pos)

	// Track if any legal move exists (for checkmate/stalemate detection)
	hasLegal := false

	for _, mv := range *pseudo {
		// Make move
		undo := MakeMove(pos, mv)

		// Check if move is legal (our king not in check)
		ourKing := kingSquare(pos, pos.SideToMove^1) // We switched sides after MakeMove
		inCheck := squareAttacked(pos, ourKing, pos.SideToMove)

		if !inCheck {
			hasLegal = true

			// Push move onto stack before recursing
			e.currentStack = append(e.currentStack, mv)

			// Legal move - recurse
			if !e.enumerateDFSRecursive(pos, depth+1, maxDepth, repCount, callback) {
				UnmakeMove(pos, mv, undo)
				e.currentStack = e.currentStack[:len(e.currentStack)-1] // Pop
				releaseMoves(pseudo)
				return false
			}

			// Pop move from stack after returning
			e.currentStack = e.currentStack[:len(e.currentStack)-1]
		}

		// Unmake move
		UnmakeMove(pos, mv, undo)
	}

	releaseMoves(pseudo)

	// Terminal: no legal moves (checkmate or stalemate)
	// This is fine - we've already visited this position above
	_ = hasLegal

	return true
}

// countDFS counts positions in a subtree without storing them.
// Used for the first pass of parallel enumeration.
func countDFS(pos *GameState, depth, maxDepth int, repCount map[uint64]uint8) uint64 {
	// Count this position
	count := uint64(1)

	// Terminal conditions
	if depth >= maxDepth {
		return count
	}

	if pos.Halfmove >= 100 {
		return count
	}

	// Threefold repetition check
	if depth >= 8 {
		posID := positionID(pos)
		if repCount[posID] >= 2 {
			return count
		}
		repCount[posID]++
		defer func() { repCount[posID]-- }()
	}

	// Generate pseudo-legal moves using pooled buffers
	pseudo := getMoveList()
	genPseudoMovesTo(pseudo, pos)

	for _, mv := range *pseudo {
		undo := MakeMove(pos, mv)

		// Check if move is legal
		ourKing := kingSquare(pos, pos.SideToMove^1)
		inCheck := squareAttacked(pos, ourKing, pos.SideToMove)

		if !inCheck {
			count += countDFS(pos, depth+1, maxDepth, repCount)
		}

		UnmakeMove(pos, mv, undo)
	}

	releaseMoves(pseudo)
	return count
}

// EnumerateDFSParallel performs two-pass parallel enumeration with deterministic ordering.
//
// Pass 1: Count subtree sizes for each root move (parallel)
// Pass 2: Enumerate each subtree with correct index offsets (parallel)
//
// This maintains deterministic position-to-index mapping while using all CPU cores.
func (e *PositionEnumeratorDFS) EnumerateDFSParallel(
	maxDepth int,
	callback func(index uint64, pos *GameState, depth int) bool,
) {
	if maxDepth <= 0 {
		maxDepth = MaxDepthDFS
	}

	// Generate root moves (deterministic order)
	moves := getMoveList()
	genPseudoMovesTo(moves, &e.startPos)

	var rootMoves []Mv
	for _, mv := range *moves {
		undo := MakeMove(&e.startPos, mv)
		ourKing := kingSquare(&e.startPos, e.startPos.SideToMove^1)
		inCheck := squareAttacked(&e.startPos, ourKing, e.startPos.SideToMove)
		UnmakeMove(&e.startPos, mv, undo)

		if !inCheck {
			rootMoves = append(rootMoves, mv)
		}
	}
	releaseMoves(moves)

	// Pass 1: Count subtree sizes (parallel)
	subtreeSizes := make([]uint64, len(rootMoves))
	var wg sync.WaitGroup

	for i, mv := range rootMoves {
		wg.Add(1)
		go func(idx int, move Mv) {
			defer wg.Done()

			pos := e.startPos
			undo := MakeMove(&pos, move)
			repCount := make(map[uint64]uint8, 100)
			subtreeSizes[idx] = countDFS(&pos, 1, maxDepth, repCount)
			UnmakeMove(&pos, move, undo)
		}(i, mv)
	}
	wg.Wait()

	// Calculate index offsets for each subtree
	offsets := make([]uint64, len(rootMoves))
	offsets[0] = 1 // Root position is index 0
	for i := 1; i < len(rootMoves); i++ {
		offsets[i] = offsets[i-1] + subtreeSizes[i-1]
	}

	// Visit root position first (index 0)
	if callback != nil {
		if !callback(0, &e.startPos, 0) {
			return // Early stop
		}
	}

	// Store root checkpoint
	e.checkpoints = append(e.checkpoints, &CheckpointDFS{
		Index: 0,
		Depth: 0,
		State: e.startPos,
	})

	// Pass 2: Enumerate subtrees in parallel
	type subtreeResult struct {
		checkpoints []*CheckpointDFS
	}

	results := make([]subtreeResult, len(rootMoves))
	var stopRequested atomic.Bool

	for i, mv := range rootMoves {
		wg.Add(1)
		go func(idx int, move Mv, startIdx uint64) {
			defer wg.Done()

			if stopRequested.Load() {
				return
			}

			pos := e.startPos
			undo := MakeMove(&pos, move)

			// Create sub-enumerator for this subtree
			subEnum := &PositionEnumeratorDFS{
				startPos:    pos,
				checkpoints: make([]*CheckpointDFS, 0, 100),
				currentIdx:  startIdx,
			}

			repCount := make(map[uint64]uint8, 100)
			stopped := !subEnum.enumerateDFSRecursive(&pos, 1, maxDepth, repCount, callback)

			if stopped {
				stopRequested.Store(true)
			}

			results[idx].checkpoints = subEnum.checkpoints
			UnmakeMove(&pos, move, undo)
		}(i, mv, offsets[i])
	}
	wg.Wait()

	// Merge checkpoints from all subtrees
	for _, result := range results {
		e.checkpoints = append(e.checkpoints, result.checkpoints...)
	}

	// Sort checkpoints by index (should already be sorted, but ensure it)
	sort.Slice(e.checkpoints, func(i, j int) bool {
		return e.checkpoints[i].Index < e.checkpoints[j].Index
	})

	// Set final index count
	if len(rootMoves) > 0 {
		lastIdx := len(rootMoves) - 1
		e.currentIdx = offsets[lastIdx] + subtreeSizes[lastIdx]
	} else {
		e.currentIdx = 1 // Just the root position
	}
}

// GetCheckpointsDFS returns all stored checkpoints.
func (e *PositionEnumeratorDFS) GetCheckpointsDFS() []*CheckpointDFS {
	return e.checkpoints
}

// GetCheckpointForIndexDFS finds the checkpoint at or before the given index.
func (e *PositionEnumeratorDFS) GetCheckpointForIndexDFS(targetIdx uint64) *CheckpointDFS {
	if len(e.checkpoints) == 0 {
		return nil
	}

	var best *CheckpointDFS
	for _, ckpt := range e.checkpoints {
		if ckpt.Index <= targetIdx && (best == nil || ckpt.Index > best.Index) {
			best = ckpt
		}
	}
	return best
}

// PositionAtIndexDFS returns the position at a specific index.
// This replays the DFS from the start until reaching the target index.
func (e *PositionEnumeratorDFS) PositionAtIndexDFS(targetIdx uint64, maxDepth int) (*GameState, bool) {
	if maxDepth <= 0 {
		maxDepth = MaxDepthDFS
	}

	// Special case: target is the start position
	if targetIdx == 0 {
		result := e.startPos
		return &result, true
	}

	// Full DFS replay until we hit the target index
	var result *GameState
	found := false
	currentIdx := uint64(0)
	repCount := make(map[uint64]uint8, 100)

	var searchDFS func(*GameState, int)
	searchDFS = func(pos *GameState, depth int) {
		if found {
			return
		}

		// Check if this is the target
		if currentIdx == targetIdx {
			result = &GameState{}
			*result = *pos
			found = true
			return
		}
		currentIdx++

		if depth >= maxDepth {
			return
		}
		if pos.Halfmove >= 100 {
			return
		}

		// Threefold repetition
		if depth >= 8 {
			posID := positionID(pos)
			if repCount[posID] >= 2 {
				return
			}
			repCount[posID]++
			defer func() { repCount[posID]-- }()
		}

		// Use same move generation as enumeration
		pseudo := getMoveList()
		genPseudoMovesTo(pseudo, pos)

		for _, mv := range *pseudo {
			if found {
				break
			}

			undo := MakeMove(pos, mv)
			ourKing := kingSquare(pos, pos.SideToMove^1)
			inCheck := squareAttacked(pos, ourKing, pos.SideToMove)

			if !inCheck {
				searchDFS(pos, depth+1)
			}

			UnmakeMove(pos, mv, undo)
		}
		releaseMoves(pseudo)
	}

	state := e.startPos
	searchDFS(&state, 0)
	return result, found
}

// IndexOfBoardDFS searches for a board state (ignoring move counters) and returns its first index.
// This is faster than IndexOfPositionDFS when you only care about the board configuration.
func (e *PositionEnumeratorDFS) IndexOfBoardDFS(target *GameState, maxDepth int) (uint64, bool) {
	return e.indexOfPositionDFSInternal(target, maxDepth, true)
}

// IndexOfPositionDFS searches for a position and returns its index.
// Requires exact match including halfmove and fullmove counters.
func (e *PositionEnumeratorDFS) IndexOfPositionDFS(target *GameState, maxDepth int) (uint64, bool) {
	return e.indexOfPositionDFSInternal(target, maxDepth, false)
}

// indexOfPositionDFSInternal is the internal implementation for position lookup.
// When stacks are available, searches all gaps using stack-based DFS resumption.
// Otherwise falls back to full DFS scan.
func (e *PositionEnumeratorDFS) indexOfPositionDFSInternal(target *GameState, maxDepth int, boardOnly bool) (uint64, bool) {
	if target == nil {
		return 0, false
	}

	if maxDepth <= 0 {
		maxDepth = MaxDepthDFS
	}

	// Get all checkpoint gaps to search
	candidateGaps := e.allGaps()

	// If we have checkpoints with stacks, search all gaps
	if len(e.checkpoints) > 0 && e.checkpoints[0].Stack != nil {
		return e.searchCandidateGaps(candidateGaps, maxDepth, target, boardOnly)
	}

	// Fall back to full DFS if no stacks available
	return e.fullDFSSearch(target, maxDepth, boardOnly, candidateGaps)
}

// searchCandidateGaps searches candidate gaps in parallel using stack-based resumption.
// Returns the lowest index found across all candidate gaps.
func (e *PositionEnumeratorDFS) searchCandidateGaps(
	candidateGaps []int,
	maxDepth int,
	target *GameState,
	boardOnly bool,
) (uint64, bool) {
	if len(candidateGaps) == 0 {
		return 0, false
	}

	// For a single gap, no need for goroutines
	if len(candidateGaps) == 1 {
		gapIdx := candidateGaps[0]
		if gapIdx < 0 || gapIdx >= len(e.checkpoints) {
			return 0, false
		}
		ckpt := e.checkpoints[gapIdx]
		var endIdx uint64
		if gapIdx+1 < len(e.checkpoints) {
			endIdx = e.checkpoints[gapIdx+1].Index
		} else {
			endIdx = e.currentIdx
		}
		return e.searchGapWithStack(ckpt, endIdx, maxDepth, target, boardOnly)
	}

	// Search all candidate gaps in parallel
	type result struct {
		gapIdx   int
		foundIdx uint64
		found    bool
	}

	results := make(chan result, len(candidateGaps))
	var wg sync.WaitGroup

	for _, gapIdx := range candidateGaps {
		if gapIdx < 0 || gapIdx >= len(e.checkpoints) {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			ckpt := e.checkpoints[idx]

			// Determine end of this gap
			var endIdx uint64
			if idx+1 < len(e.checkpoints) {
				endIdx = e.checkpoints[idx+1].Index
			} else {
				endIdx = e.currentIdx
			}

			// Search this gap
			foundIdx, found := e.searchGapWithStack(ckpt, endIdx, maxDepth, target, boardOnly)
			results <- result{gapIdx: idx, foundIdx: foundIdx, found: found}
		}(gapIdx)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and find the lowest index
	var bestIdx uint64
	bestFound := false

	for r := range results {
		if r.found {
			if !bestFound || r.foundIdx < bestIdx {
				bestIdx = r.foundIdx
				bestFound = true
			}
		}
	}

	return bestIdx, bestFound
}

// fullDFSSearch performs a full DFS when stacks aren't available.
func (e *PositionEnumeratorDFS) fullDFSSearch(
	target *GameState,
	maxDepth int,
	boardOnly bool,
	candidateGaps []int,
) (uint64, bool) {
	// Convert candidate gaps to a set for O(1) lookup
	candidateSet := make(map[int]bool)
	for _, g := range candidateGaps {
		candidateSet[g] = true
	}

	currentIdx := uint64(0)
	found := false
	var foundIdx uint64
	repCount := make(map[uint64]uint8, 100)

	var searchDFS func(*GameState, int)
	searchDFS = func(pos *GameState, depth int) {
		if found {
			return
		}

		// Determine which gap this position is in
		gapIdx := int(currentIdx / CheckpointIntervalDFS)

		// Check if this gap is a candidate (or if we have no filter info)
		if len(candidateSet) == 0 || candidateSet[gapIdx] {
			// Check if matches target
			var matches bool
			if boardOnly {
				matches = boardsEqual(pos, target)
			} else {
				matches = positionsEqual(pos, target)
			}
			if matches {
				foundIdx = currentIdx
				found = true
				return
			}
		}
		currentIdx++

		if depth >= maxDepth {
			return
		}
		if pos.Halfmove >= 100 {
			return
		}

		// Threefold repetition
		if depth >= 8 {
			posID := positionID(pos)
			if repCount[posID] >= 2 {
				return
			}
			repCount[posID]++
			defer func() { repCount[posID]-- }()
		}

		// Generate moves
		pseudo := getMoveList()
		genPseudoMovesTo(pseudo, pos)

		for _, mv := range *pseudo {
			if found {
				break
			}

			undo := MakeMove(pos, mv)
			ourKing := kingSquare(pos, pos.SideToMove^1)
			inCheck := squareAttacked(pos, ourKing, pos.SideToMove)

			if !inCheck {
				searchDFS(pos, depth+1)
			}

			UnmakeMove(pos, mv, undo)
		}
		releaseMoves(pseudo)
	}

	state := e.startPos
	searchDFS(&state, 0)
	return foundIdx, found
}

// allGaps returns indices of all checkpoint gaps.
func (e *PositionEnumeratorDFS) allGaps() []int {
	if len(e.checkpoints) == 0 {
		return nil
	}
	result := make([]int, len(e.checkpoints))
	for i := range result {
		result[i] = i
	}
	return result
}

// searchGapWithStack searches within a checkpoint gap using the stack to resume DFS.
// This allows searching from deep checkpoints by continuing the traversal from where we left off.
func (e *PositionEnumeratorDFS) searchGapWithStack(
	ckpt *CheckpointDFS,
	endIdx uint64,
	maxDepth int,
	target *GameState,
	boardOnly bool,
) (uint64, bool) {
	if ckpt.Stack == nil {
		// No stack - can't resume properly, fall back to forward search only
		return e.searchGapForward(ckpt, endIdx, maxDepth, target, boardOnly)
	}

	// Replay the stack to rebuild position and path state
	pos := e.startPos
	repCount := make(map[uint64]uint8, 100)

	for i, mv := range ckpt.Stack {
		// Track repetitions as we replay
		if i >= 8 {
			posID := positionID(&pos)
			repCount[posID]++
		}
		MakeMove(&pos, mv)
	}

	// Now continue the DFS from the checkpoint, exploring remaining siblings
	currentIdx := ckpt.Index
	found := false
	var foundIdx uint64

	// First check the checkpoint position itself
	var matches bool
	if boardOnly {
		matches = boardsEqual(&pos, target)
	} else {
		matches = positionsEqual(&pos, target)
	}
	if matches {
		return currentIdx, true
	}
	currentIdx++

	// Continue enumeration from checkpoint position
	var continueDFS func(*GameState, int, []Mv, int)
	continueDFS = func(state *GameState, depth int, stack []Mv, skipUntilIdx int) {
		if found || currentIdx >= endIdx {
			return
		}

		if depth >= maxDepth {
			return
		}
		if state.Halfmove >= 100 {
			return
		}

		// Threefold repetition
		if depth >= 8 {
			posID := positionID(state)
			if repCount[posID] >= 2 {
				return
			}
			repCount[posID]++
			defer func() { repCount[posID]-- }()
		}

		// Generate moves
		pseudo := getMoveList()
		genPseudoMovesTo(pseudo, state)

		// Determine which move index to start from
		startMoveIdx := 0
		if skipUntilIdx >= 0 && skipUntilIdx < len(stack) {
			// We're resuming - find the move in the stack and start AFTER it
			targetMv := stack[skipUntilIdx]
			for i, mv := range *pseudo {
				if mv == targetMv {
					startMoveIdx = i + 1 // Start from the NEXT move
					break
				}
			}
		}

		for i := startMoveIdx; i < len(*pseudo); i++ {
			if found || currentIdx >= endIdx {
				break
			}

			mv := (*pseudo)[i]
			undo := MakeMove(state, mv)

			// Check if move is legal
			ourKing := kingSquare(state, state.SideToMove^1)
			inCheck := squareAttacked(state, ourKing, state.SideToMove)

			if !inCheck {
				// Visit this position
				if boardOnly {
					matches = boardsEqual(state, target)
				} else {
					matches = positionsEqual(state, target)
				}
				if matches {
					foundIdx = currentIdx
					found = true
					UnmakeMove(state, mv, undo)
					break
				}
				currentIdx++

				// Recurse (no more skipping at deeper levels)
				continueDFS(state, depth+1, nil, -1)
			}

			UnmakeMove(state, mv, undo)
		}

		releaseMoves(pseudo)
	}

	// Start from checkpoint depth, using stack to know where to resume
	continueDFS(&pos, ckpt.Depth, ckpt.Stack, ckpt.Depth)

	// If not found yet, we need to also backtrack up the stack
	if !found && currentIdx < endIdx {
		// Backtrack through the stack levels
		for level := len(ckpt.Stack) - 1; level >= 0 && !found && currentIdx < endIdx; level-- {
			// Unmake moves to get to this level
			tempPos := e.startPos
			tempRepCount := make(map[uint64]uint8, 100)
			for i := 0; i < level; i++ {
				if i >= 8 {
					posID := positionID(&tempPos)
					tempRepCount[posID]++
				}
				MakeMove(&tempPos, ckpt.Stack[i])
			}

			// Continue from this level
			repCount = tempRepCount
			continueDFS(&tempPos, level, ckpt.Stack, level)
		}
	}

	return foundIdx, found
}

// searchGapForward searches forward only from the checkpoint (no stack resumption).
func (e *PositionEnumeratorDFS) searchGapForward(
	ckpt *CheckpointDFS,
	endIdx uint64,
	maxDepth int,
	target *GameState,
	boardOnly bool,
) (uint64, bool) {
	currentIdx := ckpt.Index
	found := false
	var foundIdx uint64
	repCount := make(map[uint64]uint8, 100)

	var searchDFS func(*GameState, int)
	searchDFS = func(pos *GameState, depth int) {
		if found || currentIdx >= endIdx {
			return
		}

		// Check if matches target
		var matches bool
		if boardOnly {
			matches = boardsEqual(pos, target)
		} else {
			matches = positionsEqual(pos, target)
		}
		if matches {
			foundIdx = currentIdx
			found = true
			return
		}
		currentIdx++

		if depth >= maxDepth {
			return
		}
		if pos.Halfmove >= 100 {
			return
		}

		// Threefold repetition
		if depth >= 8 {
			posID := positionID(pos)
			if repCount[posID] >= 2 {
				return
			}
			repCount[posID]++
			defer func() { repCount[posID]-- }()
		}

		pseudo := getMoveList()
		genPseudoMovesTo(pseudo, pos)

		for _, mv := range *pseudo {
			if found || currentIdx >= endIdx {
				break
			}

			undo := MakeMove(pos, mv)
			ourKing := kingSquare(pos, pos.SideToMove^1)
			inCheck := squareAttacked(pos, ourKing, pos.SideToMove)

			if !inCheck {
				searchDFS(pos, depth+1)
			}

			UnmakeMove(pos, mv, undo)
		}

		releaseMoves(pseudo)
	}

	state := ckpt.State
	searchDFS(&state, ckpt.Depth)
	return foundIdx, found
}

// CurrentIndexDFS returns the current enumeration index.
func (e *PositionEnumeratorDFS) CurrentIndexDFS() uint64 {
	return e.currentIdx
}

// ContinueFromCheckpoint continues enumeration from a checkpoint.
// This allows extending enumeration depth or resuming after interruption.
func (e *PositionEnumeratorDFS) ContinueFromCheckpoint(
	ckptIdx int,
	maxDepth int,
	callback func(uint64, *GameState, int) bool,
) error {
	if ckptIdx < 0 || ckptIdx >= len(e.checkpoints) {
		return nil
	}

	ckpt := e.checkpoints[ckptIdx]
	e.currentIdx = ckpt.Index

	state := ckpt.State
	repCount := make(map[uint64]uint8, 100)

	e.enumerateDFSRecursive(&state, ckpt.Depth, maxDepth, repCount, callback)
	return nil
}

// isZstdFile returns true if the filename indicates zstd compression.
func isZstdFile(filename string) bool {
	return strings.HasSuffix(filename, ".zstd") || strings.HasSuffix(filename, ".zst")
}

// stackToUCI converts a move stack to UCI notation string (e.g., "e2e4,e7e5,g1f3").
func stackToUCI(stack []Mv) string {
	if len(stack) == 0 {
		return ""
	}
	parts := make([]string, len(stack))
	for i, mv := range stack {
		parts[i] = mv.String()
	}
	return strings.Join(parts, ",")
}

// uciToStack parses a UCI notation string back to a move stack.
func uciToStack(uci string) ([]Mv, error) {
	if uci == "" {
		return nil, nil
	}
	parts := strings.Split(uci, ",")
	stack := make([]Mv, 0, len(parts))
	for _, part := range parts {
		mv, err := ParseUCI(part)
		if err != nil {
			return nil, fmt.Errorf("invalid UCI move %q: %w", part, err)
		}
		stack = append(stack, mv)
	}
	return stack, nil
}

// SaveCheckpointsCSV writes checkpoints to a CSV file with metadata.
// Format: index,depth,fen,stack with maxDepth metadata header.
// The stack field contains comma-separated UCI moves.
// If filename ends with .zstd or .zst, the output is compressed.
func (e *PositionEnumeratorDFS) SaveCheckpointsCSV(filename string, maxDepth int) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint file: %w", err)
	}
	defer file.Close()

	var writer *csv.Writer
	var zw *zstd.Encoder

	if isZstdFile(filename) {
		zw, err = zstd.NewWriter(file)
		if err != nil {
			return fmt.Errorf("failed to create zstd encoder: %w", err)
		}
		defer zw.Close()
		writer = csv.NewWriter(zw)
	} else {
		writer = csv.NewWriter(file)
	}
	defer writer.Flush()

	// Write header with metadata
	if err := writer.Write([]string{"# maxDepth=" + strconv.Itoa(maxDepth)}); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Write CSV header
	if err := writer.Write([]string{"index", "depth", "fen", "stack"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each checkpoint
	for _, ckpt := range e.checkpoints {
		record := []string{
			strconv.FormatUint(ckpt.Index, 10),
			strconv.Itoa(ckpt.Depth),
			ckpt.State.ToFEN(),
			stackToUCI(ckpt.Stack),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write checkpoint: %w", err)
		}
	}

	return nil
}

// LoadCheckpointsCSV loads checkpoints from a CSV file.
// Returns the number of checkpoints loaded.
// If filename ends with .zstd or .zst, the input is decompressed.
func (e *PositionEnumeratorDFS) LoadCheckpointsCSV(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer file.Close()

	var csvReader io.Reader = file

	if isZstdFile(filename) {
		zr, err := zstd.NewReader(file)
		if err != nil {
			return 0, fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		defer zr.Close()
		csvReader = zr
	}

	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = -1 // Allow variable number of fields (for metadata lines)
	reader.Comment = '#'         // Skip lines starting with #

	// Skip header lines
	headerSkipped := false

	// Read checkpoints
	e.checkpoints = make([]*CheckpointDFS, 0)
	count := 0
	var maxIdx uint64

	for {
		record, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		// Skip header line (first non-comment line)
		if !headerSkipped {
			headerSkipped = true
			continue
		}

		// Handle formats:
		// Old format: index,fen (2 fields)
		// Medium format: index,depth,fen (3 fields)
		// Stack format: index,depth,fen,stack (4 fields)
		// Legacy format: index,depth,fen,stack,bloom (5 fields) - bloom ignored
		var index uint64
		var depth int
		var fen string
		var stack []Mv

		switch len(record) {
		case 2:
			// Old format: index,fen
			index, err = strconv.ParseUint(record[0], 10, 64)
			if err != nil {
				return count, fmt.Errorf("invalid index in checkpoint: %w", err)
			}
			depth = 0
			fen = record[1]

		case 3:
			// Medium format: index,depth,fen
			index, err = strconv.ParseUint(record[0], 10, 64)
			if err != nil {
				return count, fmt.Errorf("invalid index in checkpoint: %w", err)
			}
			depth, err = strconv.Atoi(record[1])
			if err != nil {
				return count, fmt.Errorf("invalid depth in checkpoint: %w", err)
			}
			fen = record[2]

		case 4, 5:
			// Stack format: index,depth,fen,stack
			// Legacy format: index,depth,fen,stack,bloom (bloom ignored)
			index, err = strconv.ParseUint(record[0], 10, 64)
			if err != nil {
				return count, fmt.Errorf("invalid index in checkpoint: %w", err)
			}
			depth, err = strconv.Atoi(record[1])
			if err != nil {
				return count, fmt.Errorf("invalid depth in checkpoint: %w", err)
			}
			fen = record[2]

			// Parse stack if present
			if record[3] != "" {
				stack, err = uciToStack(record[3])
				if err != nil {
					// Non-fatal: just skip the stack
					stack = nil
				}
			}

		default:
			continue // Skip malformed lines
		}

		state, err := NewGame(fen)
		if err != nil {
			return count, fmt.Errorf("invalid FEN in checkpoint: %w", err)
		}

		ckpt := &CheckpointDFS{
			Index: index,
			Depth: depth,
			State: *state,
			Stack: stack,
		}

		e.checkpoints = append(e.checkpoints, ckpt)
		count++

		if index > maxIdx {
			maxIdx = index
		}
	}

	// Set currentIdx to allow gap searches to work
	if maxIdx > 0 {
		e.currentIdx = maxIdx + CheckpointIntervalDFS
	}

	return count, nil
}

// AppendCheckpointCSV appends a single checkpoint to a CSV file.
// This is useful for incremental saves during long enumerations.
func AppendCheckpointCSV(filename string, index uint64, fen string) error {
	// Check if file exists
	fileExists := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if new file
	if !fileExists {
		if err := writer.Write([]string{"index", "fen"}); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write checkpoint
	record := []string{
		strconv.FormatUint(index, 10),
		fen,
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return nil
}
