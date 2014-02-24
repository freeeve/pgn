pgn
===

a pgn parser for golang (still in development; attempting to stabilize API soon--feedback welcome)

[![Build Status](https://travis-ci.org/wfreeman/pgn.png?branch=master)](https://travis-ci.org/wfreeman/pgn)
[![Coverage Status](https://coveralls.io/repos/wfreeman/pgn/badge.png?branch=master)](https://coveralls.io/r/wfreeman/pgn?branch=master)

Normal go install... `go get github.com/wfreeman/pgn`

## minimum viable snippet

```Go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wfreeman/pgn"
)

func main() {
	f, err := os.Open("polgar.pgn")
	if err != nil {
		log.Fatal(err)
	}
	ps := pgn.NewPGNScanner(f)
	// while there's more to read in the file
	for ps.Next() {
		// scan the next game
		game, err := ps.Scan()
		if err != nil {
			log.Fatal(err)
		}

		// print out tags
		fmt.Println(game.Tags)

		// make a new board so we can get FEN positions
		b := pgn.NewBoard()
		for _, move := range game.Moves {
			// make the move on the board
			b.MakeMove(move)
			// print out FEN for each move in the game
			fmt.Println(b)
		}
	}
}
```

produces output like this for each game in the pgn file:

```
map[Event:Women's Chess Cup
    Site:Dresden GER 
    Round:7.1 
    Black:Polgar,Z 
    Result:1/2-1/2 
    Date:2006.07.08 
    White:Paehtz,E 
    WhiteElo:2438 
    BlackElo:2577 
    ECO:B35]
rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1
rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2
rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2
r1bqkbnr/pp1ppppp/2n5/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3
r1bqkbnr/pp1ppppp/2n5/2p5/3PP3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 3
r1bqkbnr/pp1ppppp/2n5/8/3pP3/5N2/PPP2PPP/RNBQKB1R w KQkq - 0 4
r1bqkbnr/pp1ppppp/2n5/8/3NP3/8/PPP2PPP/RNBQKB1R b KQkq - 0 4
r1bqkbnr/pp1ppp1p/2n3p1/8/3NP3/8/PPP2PPP/RNBQKB1R w KQkq - 0 5
r1bqkbnr/pp1ppp1p/2n3p1/8/3NP3/2N5/PPP2PPP/R1BQKB1R b KQkq - 1 5
r1bqk1nr/pp1pppbp/2n3p1/8/3NP3/2N5/PPP2PPP/R1BQKB1R w KQkq - 2 6
r1bqk1nr/pp1pppbp/2n3p1/8/3NP3/2N1B3/PPP2PPP/R2QKB1R b KQkq - 3 6
r1bqk2r/pp1pppbp/2n2np1/8/3NP3/2N1B3/PPP2PPP/R2QKB1R w KQkq - 4 7
r1bqk2r/pp1pppbp/2n2np1/8/2BNP3/2N1B3/PPP2PPP/R2QK2R b KQkq - 5 7
r1bq1rk1/pp1pppbp/2n2np1/8/2BNP3/2N1B3/PPP2PPP/R2QK2R w KQkq - 6 8
r1bq1rk1/pp1pppbp/2n2np1/8/3NP3/1BN1B3/PPP2PPP/R2QK2R b KQkq - 7 8
r1bq1rk1/1p1pppbp/2n2np1/p7/3NP3/1BN1B3/PPP2PPP/R2QK2R w KQkq - 0 9
r1bq1rk1/1p1pppbp/2n2np1/p7/3NP3/1BN1B3/PPP2PPP/R2Q1RK1 b KQkq - 1 9
r1bq1rk1/1p2ppbp/2np1np1/p7/3NP3/1BN1B3/PPP2PPP/R2Q1RK1 w KQkq - 0 10
r1bq1rk1/1p2ppbp/2np1np1/p7/3NP3/1BN1BP2/PPP3PP/R2Q1RK1 b KQkq - 0 10
r2q1rk1/1p1bppbp/2np1np1/p7/3NP3/1BN1BP2/PPP3PP/R2Q1RK1 w KQkq - 1 11
r2q1rk1/1p1bppbp/2np1np1/p7/P2NP3/1BN1BP2/1PP3PP/R2Q1RK1 b KQkq - 0 11
r2q1rk1/1p1bppbp/3p1np1/p7/P2nP3/1BN1BP2/1PP3PP/R2Q1RK1 w KQkq - 0 12
r2q1rk1/1p1bppbp/3p1np1/p7/P2BP3/1BN2P2/1PP3PP/R2Q1RK1 b KQkq - 0 12
r2q1rk1/1p2ppbp/2bp1np1/p7/P2BP3/1BN2P2/1PP3PP/R2Q1RK1 w KQkq - 1 13
r2q1rk1/1p2ppbp/2bp1np1/p7/P2BP3/1BN2P2/1PPQ2PP/R4RK1 b KQkq - 2 13
r2q1rk1/1p1nppbp/2bp2p1/p7/P2BP3/1BN2P2/1PPQ2PP/R4RK1 w KQkq - 3 14
r2q1rk1/1p1nppbp/2bp2p1/p7/P3P3/1BN1BP2/1PPQ2PP/R4RK1 b KQkq - 4 14
r2q1rk1/1p2ppbp/2bp2p1/p1n5/P3P3/1BN1BP2/1PPQ2PP/R4RK1 w KQkq - 5 15
r2q1rk1/1p2ppbp/2bp2p1/p1n5/P1B1P3/2N1BP2/1PPQ2PP/R4RK1 b KQkq - 6 15
r4rk1/1p2ppbp/1qbp2p1/p1n5/P1B1P3/2N1BP2/1PPQ2PP/R4RK1 w KQkq - 7 16
r4rk1/1p2ppbp/1qbp2p1/p1n5/P1B1P3/1PN1BP2/2PQ2PP/R4RK1 b KQkq - 0 16
r4rk1/1p2ppbp/2bp2p1/p1n5/PqB1P3/1PN1BP2/2PQ2PP/R4RK1 w KQkq - 1 17
r4rk1/1p2ppbp/2bp2p1/p1n5/PqBBP3/1PN2P2/2PQ2PP/R4RK1 b KQkq - 2 17
r4rk1/1p2pp1p/2bp2p1/p1n5/PqBbP3/1PN2P2/2PQ2PP/R4RK1 w KQkq - 0 18
r4rk1/1p2pp1p/2bp2p1/p1n5/PqBQP3/1PN2P2/2P3PP/R4RK1 b KQkq - 0 18
r4rk1/1p1npp1p/2bp2p1/p7/PqBQP3/1PN2P2/2P3PP/R4RK1 w KQkq - 1 19
r4rk1/1p1npp1p/2bp2p1/p7/PqBQP3/1PN2P2/2P3PP/3R1RK1 b KQkq - 2 19
2r2rk1/1p1npp1p/2bp2p1/p7/PqBQP3/1PN2P2/2P3PP/3R1RK1 w KQkq - 3 20
2r2rk1/1p1npp1p/2bp2p1/p7/PqBQP3/1P3P2/2P1N1PP/3R1RK1 b KQkq - 4 20
2r2rk1/1p1npp1p/3p2p1/p7/bqBQP3/1P3P2/2P1N1PP/3R1RK1 w KQkq - 0 21
2r2rk1/1p1npp1p/3p2p1/p7/bqBQP3/1PP2P2/4N1PP/3R1RK1 b KQkq - 0 21
2r2rk1/1p1npp1p/1q1p2p1/p7/b1BQP3/1PP2P2/4N1PP/3R1RK1 w KQkq - 1 22
2r2rk1/1p1npB1p/1q1p2p1/p7/b2QP3/1PP2P2/4N1PP/3R1RK1 b KQkq - 0 22
2r2r2/1p1npk1p/1q1p2p1/p7/b2QP3/1PP2P2/4N1PP/3R1RK1 w KQkq - 0 23
2r2r2/1p1npk1p/1q1p2p1/p7/P2QP3/2P2P2/4N1PP/3R1RK1 b KQkq - 0 23
5r2/1p1npk1p/1q1p2p1/p7/P1rQP3/2P2P2/4N1PP/3R1RK1 w KQkq - 1 24
5r2/1p1npk1p/1Q1p2p1/p7/P1r1P3/2P2P2/4N1PP/3R1RK1 b KQkq - 0 24
5r2/1p2pk1p/1n1p2p1/p7/P1r1P3/2P2P2/4N1PP/3R1RK1 w KQkq - 0 25
5r2/1p2pk1p/1n1p2p1/p7/P1rRP3/2P2P2/4N1PP/5RK1 b KQkq - 1 25
2r5/1p2pk1p/1n1p2p1/p7/P1rRP3/2P2P2/4N1PP/5RK1 w KQkq - 2 26
2r5/1p2pk1p/1n1p2p1/p7/P1rRP3/2P2P2/4N1PP/R5K1 b KQkq - 3 26
2r5/1p2pk1p/1n1p2p1/p7/P2rP3/2P2P2/4N1PP/R5K1 w KQkq - 0 27
2r5/1p2pk1p/1n1p2p1/p7/P2PP3/5P2/4N1PP/R5K1 b KQkq - 0 27
8/1p2pk1p/1n1p2p1/p7/P1rPP3/5P2/4N1PP/R5K1 w KQkq - 1 28
8/1p2pk1p/1n1p2p1/p7/P1rPP3/5P2/4N1PP/1R4K1 b KQkq - 2 28
8/1p2pk1p/1n1p2p1/p7/Pr1PP3/5P2/4N1PP/1R4K1 w KQkq - 3 29
8/1p2pk1p/1n1p2p1/p7/PR1PP3/5P2/4N1PP/6K1 b KQkq - 0 29
8/1p2pk1p/1n1p2p1/8/Pp1PP3/5P2/4N1PP/6K1 w KQkq - 0 30
8/1p2pk1p/1n1p2p1/8/Pp1PP3/5P2/6PP/2N3K1 b KQkq - 1 30
8/1p2pk1p/3p2p1/8/np1PP3/5P2/6PP/2N3K1 w KQkq - 0 31
8/1p2pk1p/3p2p1/8/np1PP3/1N3P2/6PP/6K1 b KQkq - 1 31
8/1p3k1p/3p2p1/4p3/np1PP3/1N3P2/6PP/6K1 w KQkq - 0 32
8/1p3k1p/3p2p1/4p3/np1PP3/1N3P2/5KPP/8 b KQkq - 1 32
8/1p3k1p/3p2p1/8/np1pP3/1N3P2/5KPP/8 w KQkq - 0 33
8/1p3k1p/3p2p1/8/np1NP3/5P2/5KPP/8 b KQkq - 0 33
8/1p3k1p/3p2p1/2n5/1p1NP3/5P2/5KPP/8 w KQkq - 1 34
8/1p3k1p/3p2p1/2n5/1p1NP3/4KP2/6PP/8 b KQkq - 2 34
8/1p5p/3p1kp1/2n5/1p1NP3/4KP2/6PP/8 w KQkq - 3 35
8/1p5p/3p1kp1/1Nn5/1p2P3/4KP2/6PP/8 b KQkq - 4 35
8/1p5p/3p1kp1/1Nn5/4P3/1p2KP2/6PP/8 w KQkq - 0 36
8/1p5p/3p1kp1/1Nn5/3KP3/1p3P2/6PP/8 b KQkq - 1 36
8/1p5p/3p1kp1/1Nn5/3KP3/5P2/1p4PP/8 w KQkq - 0 37
8/1p5p/3p1kp1/2n5/3KP3/N4P2/1p4PP/8 b KQkq - 1 37
8/7p/3p1kp1/1pn5/3KP3/N4P2/1p4PP/8 w KQkq - 0 38
8/7p/3p1kp1/1pn5/4P3/N1K2P2/1p4PP/8 b KQkq - 1 38
8/7p/3p1kp1/1p6/n3P3/N1K2P2/1p4PP/8 w KQkq - 2 39
8/7p/3p1kp1/1p6/n3P3/N4P2/1pK3PP/8 b KQkq - 3 39
8/7p/3p1kp1/8/np2P3/N4P2/1pK3PP/8 w KQkq - 0 40
8/7p/3p1kp1/8/npN1P3/5P2/1pK3PP/8 b KQkq - 1 40
8/7p/3pk1p1/8/npN1P3/5P2/1pK3PP/8 w KQkq - 2 41
8/7p/3pk1p1/8/np2P3/4NP2/1pK3PP/8 b KQkq - 3 41
8/7p/3pk3/6p1/np2P3/4NP2/1pK3PP/8 w KQkq - 0 42
8/7p/3pk3/6p1/np2P3/4NPP1/1pK4P/8 b KQkq - 0 42
8/7p/3p4/4k1p1/np2P3/4NPP1/1pK4P/8 w KQkq - 1 43
8/7p/3p4/4k1p1/npN1P3/5PP1/1pK4P/8 b KQkq - 2 43
8/7p/3p4/6p1/npNkP3/5PP1/1pK4P/8 w KQkq - 3 44
8/7p/3N4/6p1/np1kP3/5PP1/1pK4P/8 b KQkq - 0 44
8/7p/3N4/6p1/np2P3/4kPP1/1pK4P/8 w KQkq - 1 45
8/7p/3N4/4P1p1/np6/4kPP1/1pK4P/8 b KQkq - 0 45
8/7p/3N4/2n1P1p1/1p6/4kPP1/1pK4P/8 w KQkq - 1 46
8/5N1p/8/2n1P1p1/1p6/4kPP1/1pK4P/8 b KQkq - 2 46
8/5N1p/4n3/4P1p1/1p6/4kPP1/1pK4P/8 w KQkq - 3 47
8/5N1p/4n3/4P1p1/1p5P/4kPP1/1pK5/8 b KQkq - 0 47
8/5N1p/4n3/4P3/1p5p/4kPP1/1pK5/8 w KQkq - 0 48
8/5N1p/4n3/4P3/1p5P/4kP2/1pK5/8 b KQkq - 0 48
8/5N1p/4n3/4P3/1p3k1P/5P2/1pK5/8 w KQkq - 1 49
8/5N1p/4n3/4P3/1p3k1P/5P2/1K6/8 b KQkq - 0 49
8/5N1p/4n3/4Pk2/1p5P/5P2/1K6/8 w KQkq - 1 50
8/5N1p/4n3/4Pk2/1p5P/1K3P2/8/8 b KQkq - 2 50
8/5N1p/8/4Pk2/1p1n3P/1K3P2/8/8 w KQkq - 3 51
8/5N1p/8/4Pk2/1K1n3P/5P2/8/8 b KQkq - 0 51
8/5N1p/8/4Pk2/1K5P/5n2/8/8 w KQkq - 0 52
8/5N1p/8/2K1Pk2/7P/5n2/8/8 b KQkq - 1 52
8/5N1p/8/2K1nk2/7P/8/8/8 w KQkq - 0 53
8/7p/8/2K1Nk2/7P/8/8/8 b KQkq - 0 53
8/7p/8/2K1k3/7P/8/8/8 w KQkq - 0 54
8/7p/8/2K1k2P/8/8/8/8 b KQkq - 0 54
8/7p/8/2K2k1P/8/8/8/8 w KQkq - 1 55
8/7p/8/5k1P/3K4/8/8/8 b KQkq - 2 55
8/7p/8/6kP/3K4/8/8/8 w KQkq - 3 56
8/7p/8/6kP/8/4K3/8/8 b KQkq - 4 56
8/7p/8/7k/8/4K3/8/8 w KQkq - 0 57
8/7p/8/7k/8/8/5K2/8 b KQkq - 1 57
```

## license

MIT License, see LICENSE file.
