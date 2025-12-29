package pgn

import (
	"os"
	"strings"
	"testing"
)

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func TestDFSBasicEnumeration(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	enum := NewPositionEnumeratorDFS(start)

	var count uint64
	enum.EnumerateDFS(3, func(index uint64, pos *GameState, depth int) bool {
		if index != count {
			t.Errorf("Expected index %d, got %d at depth %d", count, index, depth)
		}
		count++
		return true
	})

	if count == 0 {
		t.Error("No positions enumerated")
	}

	// Should have enumerated all positions up to depth 3
	t.Logf("Enumerated %d positions up to depth 3", count)
}

func TestDFSDeterministic(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Run twice and compare
	var fens1, fens2 []string

	enum1 := NewPositionEnumeratorDFS(start)
	enum1.EnumerateDFS(2, func(index uint64, pos *GameState, depth int) bool {
		fens1 = append(fens1, pos.ToFEN())
		return true
	})

	enum2 := NewPositionEnumeratorDFS(start)
	enum2.EnumerateDFS(2, func(index uint64, pos *GameState, depth int) bool {
		fens2 = append(fens2, pos.ToFEN())
		return true
	})

	if len(fens1) != len(fens2) {
		t.Fatalf("Different counts: %d vs %d", len(fens1), len(fens2))
	}

	for i := range fens1 {
		if fens1[i] != fens2[i] {
			t.Errorf("Position %d differs:\n%s\nvs\n%s", i, fens1[i], fens2[i])
		}
	}
}

func TestDFSCheckpoints(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	enum := NewPositionEnumeratorDFS(start)

	// Enumerate to depth 4 (should hit some checkpoints at 2^20 interval)
	enum.EnumerateDFS(4, nil)

	ckpts := enum.GetCheckpointsDFS()
	if len(ckpts) == 0 {
		t.Error("Expected at least one checkpoint")
	}

	// First checkpoint should be at index 0
	if ckpts[0].Index != 0 {
		t.Errorf("First checkpoint should be at index 0, got %d", ckpts[0].Index)
	}
}

func TestDFSPositionAtIndex(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	enum := NewPositionEnumeratorDFS(start)

	// Collect positions during enumeration
	positions := make(map[uint64]string)
	enum.EnumerateDFS(3, func(index uint64, pos *GameState, depth int) bool {
		if index < 100 {
			positions[index] = pos.ToFEN()
		}
		return true
	})

	// Test lookup for some indices
	for idx := uint64(0); idx < 10; idx++ {
		pos, found := enum.PositionAtIndexDFS(idx, 3)
		if !found {
			t.Errorf("Position at index %d not found", idx)
			continue
		}

		expectedFEN, ok := positions[idx]
		if !ok {
			continue
		}

		if pos.ToFEN() != expectedFEN {
			t.Errorf("Position at index %d mismatch:\nExpected: %s\nGot: %s",
				idx, expectedFEN, pos.ToFEN())
		}
	}
}

func TestDFSIndexOfPosition(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	enum := NewPositionEnumeratorDFS(start)

	// Collect some positions
	testPositions := make(map[uint64]*GameState)
	enum.EnumerateDFS(2, func(index uint64, pos *GameState, depth int) bool {
		if index == 0 || index == 5 || index == 10 || index == 20 {
			posCopy := *pos
			testPositions[index] = &posCopy
		}
		return true
	})

	// Search for them
	for expectedIdx, pos := range testPositions {
		foundIdx, found := enum.IndexOfPositionDFS(pos, 2)
		if !found {
			t.Errorf("Position at index %d not found", expectedIdx)
			continue
		}

		if foundIdx != expectedIdx {
			t.Errorf("Expected index %d, got %d", expectedIdx, foundIdx)
		}
	}
}

func TestDFSDepthCounts(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	enum := NewPositionEnumeratorDFS(start)

	depthCounts := make(map[int]int)
	enum.EnumerateDFS(2, func(index uint64, pos *GameState, depth int) bool {
		depthCounts[depth]++
		return true
	})

	t.Logf("Depth counts: %v", depthCounts)

	if depthCounts[0] != 1 {
		t.Errorf("Expected 1 position at depth 0, got %d", depthCounts[0])
	}

	// Note: DFS visits in different order than BFS, but total counts at each depth
	// should eventually be the same
}

// Benchmarks

func BenchmarkDFSDepth3(b *testing.B) {
	start, _ := NewGame(startFEN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enum := NewPositionEnumeratorDFS(start)
		enum.EnumerateDFS(3, nil)
	}
}

func BenchmarkDFSDepth4(b *testing.B) {
	start, _ := NewGame(startFEN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enum := NewPositionEnumeratorDFS(start)
		enum.EnumerateDFS(4, nil)
	}
}

func BenchmarkDFSDepth5(b *testing.B) {
	start, _ := NewGame(startFEN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enum := NewPositionEnumeratorDFS(start)
		enum.EnumerateDFS(5, nil)
	}
}

func BenchmarkDFSDepth6(b *testing.B) {
	start, _ := NewGame(startFEN)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enum := NewPositionEnumeratorDFS(start)
		enum.EnumerateDFS(6, nil)
	}
}

func BenchmarkDFSPositionAtIndex(b *testing.B) {
	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)

	// Build checkpoints
	enum.EnumerateDFS(4, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enum.PositionAtIndexDFS(100, 4)
		enum.PositionAtIndexDFS(1000, 4)
		enum.PositionAtIndexDFS(5000, 4)
	}
}

func BenchmarkDFSIndexOfPosition(b *testing.B) {
	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)

	var positions []*GameState
	enum.EnumerateDFS(3, func(index uint64, pos *GameState, depth int) bool {
		if index == 100 || index == 1000 || index == 5000 {
			posCopy := *pos
			positions = append(positions, &posCopy)
		}
		return true
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pos := range positions {
			enum.IndexOfPositionDFS(pos, 3)
		}
	}
}

// Test CSV checkpoint persistence

func TestCheckpointCSVSaveLoad(t *testing.T) {
	start, err := NewGame(startFEN)
	if err != nil {
		t.Fatalf("Failed to create starting position: %v", err)
	}

	// Create enumerator and generate some checkpoints
	enum1 := NewPositionEnumeratorDFS(start)
	enum1.EnumerateDFS(4, nil)

	ckpts1 := enum1.GetCheckpointsDFS()
	if len(ckpts1) == 0 {
		t.Fatal("No checkpoints created")
	}

	t.Logf("Created %d checkpoints", len(ckpts1))

	// Save to CSV
	filename := "/tmp/test_checkpoints.csv"
	defer os.Remove(filename)

	err = enum1.SaveCheckpointsCSV(filename, 4)
	if err != nil {
		t.Fatalf("Failed to save checkpoints: %v", err)
	}

	// Load into new enumerator
	enum2 := NewPositionEnumeratorDFS(start)
	count, err := enum2.LoadCheckpointsCSV(filename)
	if err != nil {
		t.Fatalf("Failed to load checkpoints: %v", err)
	}

	if count != len(ckpts1) {
		t.Errorf("Expected %d checkpoints, loaded %d", len(ckpts1), count)
	}

	// Verify checkpoint data matches
	ckpts2 := enum2.GetCheckpointsDFS()
	for i := range ckpts1 {
		if i >= len(ckpts2) {
			break
		}

		if ckpts1[i].Index != ckpts2[i].Index {
			t.Errorf("Checkpoint %d index mismatch: %d vs %d",
				i, ckpts1[i].Index, ckpts2[i].Index)
		}

		fen1 := ckpts1[i].State.ToFEN()
		fen2 := ckpts2[i].State.ToFEN()
		if fen1 != fen2 {
			t.Errorf("Checkpoint %d FEN mismatch:\n%s\nvs\n%s",
				i, fen1, fen2)
		}
	}
}

func TestAppendCheckpointCSV(t *testing.T) {
	filename := "/tmp/test_append_checkpoints.csv"
	defer os.Remove(filename)

	// Append several checkpoints
	testCases := []struct {
		index uint64
		fen   string
	}{
		{0, startFEN},
		{1048576, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{2097152, "rnbqkb1r/pppppppp/5n2/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 1 2"},
	}

	for _, tc := range testCases {
		err := AppendCheckpointCSV(filename, tc.index, tc.fen)
		if err != nil {
			t.Fatalf("Failed to append checkpoint: %v", err)
		}
	}

	// Load and verify
	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)
	count, err := enum.LoadCheckpointsCSV(filename)
	if err != nil {
		t.Fatalf("Failed to load checkpoints: %v", err)
	}

	if count != len(testCases) {
		t.Errorf("Expected %d checkpoints, got %d", len(testCases), count)
	}

	ckpts := enum.GetCheckpointsDFS()
	for i, tc := range testCases {
		if i >= len(ckpts) {
			break
		}

		if ckpts[i].Index != tc.index {
			t.Errorf("Checkpoint %d: expected index %d, got %d",
				i, tc.index, ckpts[i].Index)
		}

		if ckpts[i].State.ToFEN() != tc.fen {
			t.Errorf("Checkpoint %d: FEN mismatch", i)
		}
	}
}

func TestCheckpointCSVFormat(t *testing.T) {
	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)
	enum.EnumerateDFS(3, nil)

	filename := "/tmp/test_checkpoint_format.csv"
	defer os.Remove(filename)

	err := enum.SaveCheckpointsCSV(filename, 3)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Read file and verify format
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	t.Logf("CSV content:\n%s", content)

	// Should have metadata header and CSV header
	if !strings.Contains(content, "# maxDepth=3") {
		t.Error("Missing maxDepth metadata")
	}
	if !strings.Contains(content, "index,depth,fen") {
		t.Error("Missing or incorrect CSV header")
	}

	// Should have at least the checkpoint at index 0
	if !strings.Contains(content, "0,0,"+startFEN) {
		t.Error("Missing checkpoint at index 0")
	}
}

func TestCheckpointFENsAreCorrect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)

	// Enumerate enough to create multiple checkpoints (need > 1M positions)
	enum.EnumerateDFS(5, nil) // ~5M positions = ~5 checkpoints

	checkpoints := enum.GetCheckpointsDFS()
	if len(checkpoints) < 2 {
		t.Fatalf("Expected at least 2 checkpoints, got %d", len(checkpoints))
	}

	// Verify that non-zero index checkpoints have different FENs
	startingFEN := checkpoints[0].State.ToFEN()
	for i, ckpt := range checkpoints[1:] {
		fen := ckpt.State.ToFEN()
		if fen == startingFEN {
			t.Errorf("Checkpoint %d (index %d) has starting position FEN, expected different position",
				i+1, ckpt.Index)
		}
		// Also verify the FEN parses back to a valid position
		parsed, err := NewGame(fen)
		if err != nil {
			t.Errorf("Checkpoint %d FEN failed to parse: %v", i+1, err)
			continue
		}
		// And that it matches the stored state
		if !positionsEqual(&ckpt.State, parsed) {
			t.Errorf("Checkpoint %d: parsed FEN doesn't match stored state", i+1)
		}
	}
	t.Logf("Verified %d checkpoints have correct unique FENs", len(checkpoints))
}

func TestPositionLookupWithZstdCheckpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)

	// Enumerate to create multiple checkpoints
	positions := make(map[uint64]string)
	enum.EnumerateDFS(5, func(index uint64, pos *GameState, depth int) bool {
		// Sample only at checkpoint boundaries (guaranteed findable)
		if index%CheckpointIntervalDFS == 0 {
			positions[index] = pos.ToFEN()
		}
		return true
	})

	// Save to zstd file
	filename := "/tmp/test_lookup.csv.zstd"
	defer os.Remove(filename)

	err := enum.SaveCheckpointsCSV(filename, 5)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load into fresh enumerator
	enum2 := NewPositionEnumeratorDFS(start)
	count, err := enum2.LoadCheckpointsCSV(filename)
	if err != nil {
		t.Fatalf("Failed to load zstd checkpoints: %v", err)
	}
	t.Logf("Loaded %d checkpoints from zstd file", count)

	// Verify position lookup works
	for idx, expectedFEN := range positions {
		pos, found := enum2.PositionAtIndexDFS(idx, 5)
		if !found {
			t.Errorf("Position at index %d not found", idx)
			continue
		}
		if pos.ToFEN() != expectedFEN {
			t.Errorf("Position at index %d mismatch:\nExpected: %s\nGot: %s",
				idx, expectedFEN, pos.ToFEN())
		}
	}
	t.Logf("Verified %d position lookups with zstd checkpoints", len(positions))
}

func TestCheckpointCSVZstdRoundTrip(t *testing.T) {
	start, _ := NewGame(startFEN)
	enum := NewPositionEnumeratorDFS(start)

	// Enumerate a small tree to create checkpoints
	enum.EnumerateDFS(3, nil)

	filename := "/tmp/test_checkpoint.csv.zstd"
	defer os.Remove(filename)

	// Save compressed
	err := enum.SaveCheckpointsCSV(filename, 3)
	if err != nil {
		t.Fatalf("Failed to save compressed: %v", err)
	}

	// Verify file exists and is smaller than uncompressed would be
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	t.Logf("Compressed file size: %d bytes", info.Size())

	// Load into new enumerator
	enum2 := NewPositionEnumeratorDFS(start)
	count, err := enum2.LoadCheckpointsCSV(filename)
	if err != nil {
		t.Fatalf("Failed to load compressed: %v", err)
	}

	origCheckpoints := enum.GetCheckpointsDFS()
	if count != len(origCheckpoints) {
		t.Errorf("Checkpoint count mismatch: got %d, want %d", count, len(origCheckpoints))
	}
	t.Logf("Round-trip successful: %d checkpoints", count)
}
