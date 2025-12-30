package pgn

import (
	"fmt"
	"testing"
	"time"
)

// TestDFSScalability measures DFS performance at increasing depths
// and projects feasibility of reaching depth 15-30.
// Skipped by default as it takes a long time (runs to depth 8).
func TestDFSScalability(t *testing.T) {
	t.Skip("Skipping long-running scalability test")
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	type depthStats struct {
		depth     int
		positions uint64
		duration  time.Duration
		memAlloc  uint64
		avgBranch float64
	}

	var stats []depthStats

	// Test depths 1-8 (go beyond if time permits)
	maxTestDepth := 8
	for depth := 1; depth <= maxTestDepth; depth++ {
		enum := NewPositionEnumeratorDFS(start)

		var count uint64
		startTime := time.Now()

		enum.EnumerateDFS(depth, func(index uint64, pos *GameState, d int) bool {
			count++
		return true
	})

		duration := time.Since(startTime)

		// Calculate average branching factor
		var avgBranch float64
		if depth > 0 && len(stats) > 0 {
			prevCount := stats[len(stats)-1].positions
			if prevCount > 0 {
				avgBranch = float64(count) / float64(prevCount)
			}
		}

		stat := depthStats{
			depth:     depth,
			positions: count,
			duration:  duration,
			avgBranch: avgBranch,
		}
		stats = append(stats, stat)

		t.Logf("Depth %2d: %10d positions in %12s (%.1fx growth)",
			depth, count, duration, avgBranch)

		// Stop if taking too long
		if duration > 30*time.Second {
			t.Logf("Stopping at depth %d (taking too long)", depth)
			break
		}
	}

	// Project time for deeper depths
	t.Log("\n=== Projections ===")
	if len(stats) >= 3 {
		// Use last 3 depths to estimate average branching factor
		var totalBranch float64
		var branchCount float64
		for i := len(stats) - 3; i < len(stats); i++ {
			if stats[i].avgBranch > 0 {
				totalBranch += stats[i].avgBranch
				branchCount++
			}
		}
		avgBranchingFactor := totalBranch / branchCount

		// Use last measurement to project
		lastStat := stats[len(stats)-1]
		posPerSec := float64(lastStat.positions) / lastStat.duration.Seconds()

		t.Logf("Estimated average branching factor: %.2f", avgBranchingFactor)
		t.Logf("Current enumeration rate: %.2f M positions/sec", posPerSec/1e6)

		// Project for depths 10, 15, 20, 30
		projections := []int{10, 15, 20, 30}
		lastDepth := lastStat.depth
		lastPositions := float64(lastStat.positions)

		for _, targetDepth := range projections {
			if targetDepth <= lastDepth {
				continue
			}

			// Estimate positions at target depth
			depthDiff := targetDepth - lastDepth
			estimatedPositions := lastPositions
			for i := 0; i < depthDiff; i++ {
				estimatedPositions *= avgBranchingFactor
			}

			// Estimate time
			estimatedSeconds := estimatedPositions / posPerSec
			estimatedDuration := time.Duration(estimatedSeconds * float64(time.Second))

			t.Logf("Depth %2d: ~%.2e positions, estimated time: %s",
				targetDepth,
				estimatedPositions,
				formatDuration(estimatedDuration))
		}
	}
}

// TestReachDeepPosition tests reaching a SINGLE position at deep depth.
// This is much faster than enumerating ALL positions.
func TestReachDeepPosition(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Test: Can we quickly reach a position at depth 30?
	// Just follow the first legal move at each step
	pos := *start
	stack := make([]Mv, 0, 50)

	targetDepth := 30
	startTime := time.Now()

	for depth := 0; depth < targetDepth; depth++ {
		moves := GenerateLegalMoves(&pos)
		if len(moves) == 0 {
			t.Logf("Game ended at depth %d (checkmate/stalemate)", depth)
			break
		}

		// Take first move
		mv := moves[0]
		undo := MakeMove(&pos, mv)
		stack = append(stack, mv)

		// We could unmake if needed, but we're just going forward
		_ = undo
	}

	duration := time.Since(startTime)

	t.Logf("Reached depth %d in %s (following first move each time)",
		len(stack), duration)
	t.Logf("Final position: %s", pos.ToFEN())

	// This should be nearly instantaneous (<1ms)
	if duration > 10*time.Millisecond {
		t.Errorf("Reaching depth %d took too long: %s", targetDepth, duration)
	}
}

// TestEnumerateToFirstNPositions tests enumerating the first N positions
// regardless of depth. This is useful for checkpoint testing.
// Skipped by default as it takes a long time (runs to depth 100).
func TestEnumerateToFirstNPositions(t *testing.T) {
	t.Skip("Skipping long-running enumeration test")
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// How long to enumerate first 1M positions? First 10M?
	targets := []uint64{
		100_000,
		1_000_000,
		10_000_000,
	}

	for _, targetCount := range targets {
		enum := NewPositionEnumeratorDFS(start)

		var count uint64
		startTime := time.Now()

		// Use a very high maxDepth, but stop after reaching target count
		enum.EnumerateDFS(100, func(index uint64, pos *GameState, depth int) bool {
			count++
			if count >= targetCount {
				// Note: can't easily stop enumeration early from callback
				// This is a limitation we could fix
			}
			return true
		})

		duration := time.Since(startTime)
		posPerSec := float64(count) / duration.Seconds()

		t.Logf("Enumerated %d positions in %s (%.2f M positions/sec)",
			count, duration, posPerSec/1e6)

		// For checkpoint interval of 2^20 = 1,048,576
		if targetCount >= CheckpointIntervalDFS {
			checkpointCount := count / CheckpointIntervalDFS
			t.Logf("  → Would create %d checkpoints", checkpointCount)
		}

		// Stop if taking too long
		if duration > 60*time.Second {
			t.Logf("Stopping (taking too long)")
			break
		}
	}
}

// TestDFSMemoryUsage tests memory characteristics at various depths
// Skipped by default as it takes a long time (runs to depth 10).
func TestDFSMemoryUsage(t *testing.T) {
	t.Skip("Skipping long-running memory test")
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Test: what's the max depth reached during enumeration?
	depths := []int{5, 7, 10}

	for _, maxDepth := range depths {
		enum := NewPositionEnumeratorDFS(start)

		var count uint64
		var maxDepthReached int

		startTime := time.Now()
		enum.EnumerateDFS(maxDepth, func(index uint64, pos *GameState, depth int) bool {
			count++
			if depth > maxDepthReached {
				maxDepthReached = depth
			}
			return true
		})
		duration := time.Since(startTime)

		// Stack size = maxDepth * sizeof(GameState + undo info)
		// GameState ≈ 169 bytes, UndoInfo ≈ 32 bytes
		estimatedStackBytes := maxDepthReached * (169 + 32)

		t.Logf("MaxDepth=%d: enumerated %d positions in %s",
			maxDepth, count, duration)
		t.Logf("  Max depth reached: %d", maxDepthReached)
		t.Logf("  Estimated stack size: %d bytes (%.2f KB)",
			estimatedStackBytes, float64(estimatedStackBytes)/1024)

		if duration > 60*time.Second {
			break
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.String()
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%.1f days", d.Hours()/24)
	}
	return fmt.Sprintf("%.1f years", d.Hours()/(24*365))
}
