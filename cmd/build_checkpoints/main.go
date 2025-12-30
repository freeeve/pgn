// build_checkpoints is a tool for building position index checkpoint files.
//
// Usage:
//   build_checkpoints -depth 7                     # Output: checkpoints_depth7.csv.zst
//   build_checkpoints -depth 8 -cores 8            # Output: checkpoints_depth8.csv.zst
//   build_checkpoints -depth 9 -output custom.csv  # Uncompressed output
//   build_checkpoints -depth 9 -resume             # Resume from existing file
//
// The checkpoint files enable fast bidirectional position<->index mapping
// for all positions up to the specified depth. Files ending in .zst are
// automatically compressed/decompressed.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pgn "github.com/freeeve/pgn/v2"
)

const checkpointInterval = 1 << 20 // 1,048,576 positions

// Known perft sums (total nodes at each depth) for the starting position.
// These are the sums of perft(0) through perft(depth).
// Used for progress bar calculation.
var perftSums = map[int]uint64{
	0:  1,
	1:  21,
	2:  421,
	3:  9_323,
	4:  206_604,
	5:  5_072_213,
	6:  124_132_537,
	7:  3_320_034_397,
	8:  88_319_013_353,
	9:  2_527_849_247_520,
	10: 71_880_708_959_937,
}

func main() {
	// Command-line flags
	depth := flag.Int("depth", 6, "Maximum depth to enumerate (default: 6)")
	output := flag.String("output", "", "Output file (default: checkpoints_depth{N}.csv.zst)")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of CPU cores to use (default: all)")
	singleThreaded := flag.Bool("single", false, "Use single-threaded enumeration (for testing)")
	startFEN := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Starting position FEN")
	resume := flag.Bool("resume", false, "Resume from existing checkpoint file")
	boardInterval := flag.Int("board-interval", 60, "Seconds between board diagram displays (0 to disable)")

	flag.Parse()

	// Generate default output filename if not specified
	if *output == "" {
		*output = fmt.Sprintf("checkpoints_depth%d.csv.zst", *depth)
	}

	if *depth < 1 || *depth > 12 {
		fmt.Fprintf(os.Stderr, "Error: -depth must be between 1 and 12\n")
		os.Exit(1)
	}

	if *cores < 1 {
		fmt.Fprintf(os.Stderr, "Error: -cores must be at least 1\n")
		os.Exit(1)
	}

	// Set GOMAXPROCS to control parallelism
	runtime.GOMAXPROCS(*cores)

	fmt.Printf("Position Index Checkpoint Builder\n")
	fmt.Printf("==================================\n")
	fmt.Printf("Depth:        %d\n", *depth)
	fmt.Printf("Output:       %s\n", *output)
	fmt.Printf("Cores:        %d\n", *cores)
	fmt.Printf("Mode:         ")
	if *singleThreaded {
		fmt.Printf("single-threaded\n")
	} else {
		fmt.Printf("parallel\n")
	}
	fmt.Printf("Starting FEN: %s\n", *startFEN)

	// Get expected total for progress bar
	expectedTotal, hasExpected := perftSums[*depth]
	if hasExpected {
		fmt.Printf("Expected:     %s positions\n", formatNumber(expectedTotal))
	}
	fmt.Printf("\n")

	// Parse starting position
	start, err := pgn.NewGame(*startFEN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid FEN: %v\n", err)
		os.Exit(1)
	}

	// Check for resume (informational only for now)
	if *resume {
		if resumeIndex := loadResumePoint(*output); resumeIndex > 0 {
			fmt.Printf("Found existing checkpoint at index %s\n", formatNumber(resumeIndex))
			fmt.Printf("Note: Full resume not yet implemented - will restart enumeration\n\n")
		}
	}

	// Create enumerator
	enum := pgn.NewPositionEnumeratorDFS(start)

	// Open output file for incremental writes (crash recovery)
	partialFile := *output + ".partial"
	outFile, err := os.Create(partialFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}

	// Write header
	fmt.Fprintf(outFile, "# maxDepth=%d\n", *depth)
	fmt.Fprintf(outFile, "# Partial checkpoint file - will be finalized on completion\n")
	fmt.Fprintf(outFile, "index,depth,fen\n")

	// Thread-safe writer for incremental checkpoints
	var writeMu sync.Mutex
	var checkpointCount atomic.Uint64

	// Create progress tracker
	var positionCount atomic.Uint64
	var phase atomic.Int32 // 0=init, 1=counting, 2=enumerating

	// Track current position for board display
	var currentPosMu sync.Mutex
	var currentPos *pgn.GameState
	var currentPosDepth int

	// Progress display goroutine
	done := make(chan struct{})
	startTime := time.Now()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		// Use exponential moving average for smoother rate
		var smoothRate float64
		lastBoardTime := time.Now()
		boardPrinted := false
		const boardLines = 12 // title + header + 8 ranks + footer + blank

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				current := positionCount.Load()
				currentPhase := phase.Load()
				now := time.Now()
				elapsed := now.Sub(startTime)

				// Calculate overall rate (more stable than interval-based)
				var rate float64
				if elapsed.Seconds() > 0 {
					rate = float64(current) / elapsed.Seconds()
				}

				// Smooth the rate with EMA (alpha = 0.3)
				if smoothRate == 0 {
					smoothRate = rate
				} else {
					smoothRate = 0.3*rate + 0.7*smoothRate
				}

				// Calculate progress
				var progress float64
				var eta string
				if hasExpected && expectedTotal > 0 {
					progress = float64(current) / float64(expectedTotal) * 100
					remaining := expectedTotal - current
					if smoothRate > 0 {
						etaSecs := float64(remaining) / smoothRate
						eta = formatDuration(time.Duration(etaSecs * float64(time.Second)))
					} else {
						eta = "calculating..."
					}
				}

				// Build progress bar
				progressBar := buildProgressBar(progress, 40)

				// Phase indicator
				phaseStr := ""
				if currentPhase == 1 {
					phaseStr = " [counting]"
				}

				// Show board periodically (if enabled)
				if *boardInterval > 0 && now.Sub(lastBoardTime) >= time.Duration(*boardInterval)*time.Second {
					currentPosMu.Lock()
					if currentPos != nil {
						// Move cursor up to overwrite previous board if we printed one
						if boardPrinted {
							fmt.Printf("\r\033[K\033[%dA", boardLines)
						} else {
							fmt.Printf("\r\033[K\n")
						}
						fmt.Printf("Current position (depth %d, index %s):\n", currentPosDepth, formatNumber(current))
						fmt.Print(currentPos.String())
						fmt.Printf("\n")
						boardPrinted = true
					}
					currentPosMu.Unlock()
					lastBoardTime = now
				}

				// Clear line and print status
				fmt.Printf("\r\033[K%s %5.1f%% | %s | %.1fM/s | ETA: %s | Elapsed: %s%s",
					progressBar,
					progress,
					formatNumber(current),
					smoothRate/1e6,
					eta,
					formatDuration(elapsed),
					phaseStr)
			}
		}
	}()

	// Progress callback - updates counter and writes checkpoints incrementally
	var lastCheckpointNum atomic.Uint64
	progressCallback := func(index uint64, pos *pgn.GameState, posDepth int) bool {
		// Batch progress updates: add 65536 every 65536 positions to reduce contention
		if index&0xFFFF == 0 {
			positionCount.Add(0x10000)
		}

		// Sample position for board display (every ~64K positions)
		if index&0xFFFF == 0 {
			currentPosMu.Lock()
			posCopy := pos.Copy()
			currentPos = posCopy
			currentPosDepth = posDepth
			currentPosMu.Unlock()
		}

		// Check if we should write a checkpoint (every ~1M positions)
		checkpointNum := index / checkpointInterval

		// Try to claim this checkpoint number
		for {
			lastNum := lastCheckpointNum.Load()
			if checkpointNum <= lastNum {
				break // Already written or past this point
			}
			// Try to claim it
			if lastCheckpointNum.CompareAndSwap(lastNum, checkpointNum) {
				writeMu.Lock()
				fen := pos.ToFEN()
				fmt.Fprintf(outFile, "%d,%d,%s\n", index, posDepth, fen)
				count := checkpointCount.Add(1)

				// Sync every 10 checkpoints for crash safety
				if count%10 == 0 {
					outFile.Sync()
				}
				writeMu.Unlock()
				break
			}
			// Someone else got it, try again with updated value
		}

		return true // Continue enumeration
	}

	// Enumerate positions
	if !*singleThreaded {
		fmt.Printf("Enumerating positions (counting subtrees first)...\n")
	} else {
		fmt.Printf("Enumerating positions...\n")
	}
	phase.Store(2) // Enumeration phase

	if *singleThreaded {
		enum.EnumerateDFS(*depth, progressCallback)
	} else {
		enum.EnumerateDFSParallel(*depth, progressCallback)
	}

	close(done)

	// Close partial file
	outFile.Sync()
	outFile.Close()

	// Final statistics
	elapsed := time.Since(startTime)
	totalPositions := enum.CurrentIndexDFS()
	checkpoints := enum.GetCheckpointsDFS()
	numCheckpoints := len(checkpoints)
	incrementalCount := checkpointCount.Load()

	// Clear progress line and print final stats
	fmt.Printf("\r\033[K") // Clear line
	fmt.Printf("\n")
	fmt.Printf("Enumeration complete!\n")
	fmt.Printf("  Positions:   %s\n", formatNumber(totalPositions))
	fmt.Printf("  Checkpoints: %d (wrote %d incrementally to %s)\n", numCheckpoints, incrementalCount, partialFile)
	fmt.Printf("  Time:        %s\n", formatDuration(elapsed))
	fmt.Printf("  Rate:        %.1f M positions/sec\n\n", float64(totalPositions)/elapsed.Seconds()/1e6)

	fmt.Printf("Saving final checkpoints to %s...\n", *output)
	saveStart := time.Now()

	if err := enum.SaveCheckpointsCSV(*output, *depth); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving checkpoints: %v\n", err)
		os.Exit(1)
	}

	saveTime := time.Since(saveStart)

	// Get file size
	info, err := os.Stat(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved in %s\n", formatDuration(saveTime))
	fmt.Printf("File size: %.2f MB\n\n", float64(info.Size())/1e6)

	// Verify by loading back
	fmt.Printf("Verifying...\n")
	verify := pgn.NewPositionEnumeratorDFS(start)
	count, err := verify.LoadCheckpointsCSV(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checkpoints: %v\n", err)
		os.Exit(1)
	}

	if count != numCheckpoints {
		fmt.Fprintf(os.Stderr, "Verification failed: expected %d checkpoints, loaded %d\n", numCheckpoints, count)
		os.Exit(1)
	}

	fmt.Printf("Verified: loaded %d checkpoints\n\n", count)

	// Clean up partial file on success
	os.Remove(partialFile)

	fmt.Printf("Success! Checkpoint file ready for use.\n")
	fmt.Printf("(Removed partial file %s)\n", partialFile)
}

// loadResumePoint reads the last checkpoint index from an existing file.
func loadResumePoint(filename string) uint64 {
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()

	var lastIndex uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "index") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			var idx uint64
			fmt.Sscanf(parts[0], "%d", &idx)
			if idx > lastIndex {
				lastIndex = idx
			}
		}
	}
	return lastIndex
}

// buildProgressBar creates an ASCII progress bar.
func buildProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return "[" + bar + "]"
}

// formatNumber formats a number with thousand separators.
func formatNumber(n uint64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
		if len(s) > remainder {
			result.WriteString(",")
		}
	}

	for i := remainder; i < len(s); i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(s[i : i+3])
	}

	return result.String()
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
