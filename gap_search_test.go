package pgn

import (
	"testing"
)

func TestGapSearchFromCheckpoint(t *testing.T) {
	start := NewStartingPosition()
	enum := NewPositionEnumeratorDFS(start)

	// Track a position at a specific index
	var targetPos *GameState
	var targetIdx uint64 = 1_500_000
	var targetDepth int

	enum.EnumerateDFS(6, func(index uint64, pos *GameState, depth int) bool {
		if index == targetIdx {
			targetPos = pos.Copy()
			targetDepth = depth
		}
		return true
	})

	t.Logf("Target at index %d, depth %d", targetIdx, targetDepth)
	t.Logf("Target FEN: %s", targetPos.ToFEN())

	// Find the checkpoint for this gap
	gapIdx := int(targetIdx / CheckpointIntervalDFS)
	ckpt := enum.GetCheckpointsDFS()[gapIdx]
	t.Logf("Checkpoint %d: index=%d, depth=%d", gapIdx, ckpt.Index, ckpt.Depth)
	t.Logf("Checkpoint FEN: %s", ckpt.State.ToFEN())
	t.Logf("Stack length: %d", len(ckpt.Stack))

	// Now test the stack-based lookup
	foundIdx, found := enum.IndexOfBoardDFS(targetPos, 6)
	if !found {
		t.Errorf("Stack-based search failed to find position at index %d", targetIdx)
	} else if foundIdx != targetIdx {
		// Position might be found at an earlier index due to transposition
		t.Logf("Found position at index %d (target was %d)", foundIdx, targetIdx)
	} else {
		t.Logf("SUCCESS: Found position at correct index %d", foundIdx)
	}
}

func TestStackBasedGapSearch(t *testing.T) {
	start := NewStartingPosition()
	enum := NewPositionEnumeratorDFS(start)

	// Test positions at various indices across multiple gaps
	testIndices := []uint64{
		500_000,   // Within first gap
		1_048_576, // Exactly at checkpoint boundary
		1_500_000, // In second gap (requires backtracking)
		2_097_152, // At third checkpoint
		2_500_000, // In third gap
	}

	// First enumerate to build checkpoints with stacks
	targetPositions := make(map[uint64]*GameState)
	enum.EnumerateDFS(6, func(index uint64, pos *GameState, depth int) bool {
		for _, target := range testIndices {
			if index == target {
				targetPositions[index] = pos.Copy()
			}
		}
		return true
	})

	t.Logf("Enumerated %d positions, %d checkpoints", enum.CurrentIndexDFS(), len(enum.GetCheckpointsDFS()))

	// Verify stacks are recorded
	for i, ckpt := range enum.GetCheckpointsDFS() {
		if ckpt.Stack == nil {
			t.Errorf("Checkpoint %d has no stack", i)
		} else {
			t.Logf("Checkpoint %d at index %d, depth %d, stack length %d",
				i, ckpt.Index, ckpt.Depth, len(ckpt.Stack))
		}
		if i >= 3 {
			break // Just show first few
		}
	}

	// Test that each position can be found
	for targetIdx, pos := range targetPositions {
		foundIdx, found := enum.IndexOfBoardDFS(pos, 6)
		if !found {
			t.Errorf("Failed to find position at index %d", targetIdx)
		} else if foundIdx > targetIdx {
			t.Errorf("Found position at later index %d (expected %d)", foundIdx, targetIdx)
		} else {
			t.Logf("Position at %d found at index %d (transposition possible if earlier)", targetIdx, foundIdx)
		}
	}
}
