// build_checkpoints is a tool for building position index checkpoint files.
//
// Usage:
//   build_checkpoints -depth 7 -output positions_d7.csv
//   build_checkpoints -depth 8 -cores 8 -output positions_d8.csv
//
// The checkpoint files enable fast bidirectional position<->index mapping
// for all positions up to the specified depth.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	pgn "github.com/freeeve/pgn/v2"
)

func main() {
	// Command-line flags
	depth := flag.Int("depth", 6, "Maximum depth to enumerate (default: 6)")
	output := flag.String("output", "", "Output CSV file (required)")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of CPU cores to use (default: all)")
	singleThreaded := flag.Bool("single", false, "Use single-threaded enumeration (for testing)")
	startFEN := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Starting position FEN")

	flag.Parse()

	// Validate inputs
	if *output == "" {
		fmt.Fprintf(os.Stderr, "Error: -output is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if *depth < 1 || *depth > 100 {
		fmt.Fprintf(os.Stderr, "Error: -depth must be between 1 and 100\n")
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
	fmt.Printf("\n")

	// Parse starting position
	start, err := pgn.NewGame(*startFEN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid FEN: %v\n", err)
		os.Exit(1)
	}

	// Create enumerator
	enum := pgn.NewPositionEnumeratorDFS(start)

	// Enumerate positions
	fmt.Printf("Enumerating positions...\n")
	startTime := time.Now()

	if *singleThreaded {
		enum.EnumerateDFS(*depth, nil)
	} else {
		enum.EnumerateDFSParallel(*depth, nil)
	}

	elapsed := time.Since(startTime)

	// Get statistics
	totalPositions := enum.CurrentIndexDFS()
	checkpoints := enum.GetCheckpointsDFS()
	numCheckpoints := len(checkpoints)

	fmt.Printf("\nEnumeration complete!\n")
	fmt.Printf("  Positions:   %d (%.2f billion)\n", totalPositions, float64(totalPositions)/1e9)
	fmt.Printf("  Checkpoints: %d\n", numCheckpoints)
	fmt.Printf("  Time:        %s\n", elapsed)
	fmt.Printf("  Rate:        %.1f M positions/sec\n\n", float64(totalPositions)/elapsed.Seconds()/1e6)

	// Save to CSV
	fmt.Printf("Saving checkpoints to %s...\n", *output)
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

	fmt.Printf("Saved in %s\n", saveTime)
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

	fmt.Printf("✓ Verified: loaded %d checkpoints\n\n", count)
	fmt.Printf("Success! Checkpoint file ready for use.\n")
}
