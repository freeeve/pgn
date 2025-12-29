// Package pgn provides PGN parsing and chess position handling using bitboards.
//
// The core types are:
//   - GameState: bitboard-based position representation
//   - Mv: compact move encoding
//   - PGNScanner: streaming PGN parser
//   - ParsePGNParallel: high-performance parallel parser
package pgn

import "fmt"

// Bitboard represents a set of squares as a 64-bit integer where each bit
// corresponds to a square (bit 0 = a1, bit 63 = h8).
type Bitboard uint64

// Square is a board square index from 0-63, where a1=0, b1=1, ..., h8=63.
type Square int

// Square constants for convenient reference.
const (
	SqA1 Square = iota
	SqB1
	SqC1
	SqD1
	SqE1
	SqF1
	SqG1
	SqH1
	SqA2
	SqB2
	SqC2
	SqD2
	SqE2
	SqF2
	SqG2
	SqH2
	SqA3
	SqB3
	SqC3
	SqD3
	SqE3
	SqF3
	SqG3
	SqH3
	SqA4
	SqB4
	SqC4
	SqD4
	SqE4
	SqF4
	SqG4
	SqH4
	SqA5
	SqB5
	SqC5
	SqD5
	SqE5
	SqF5
	SqG5
	SqH5
	SqA6
	SqB6
	SqC6
	SqD6
	SqE6
	SqF6
	SqG6
	SqH6
	SqA7
	SqB7
	SqC7
	SqD7
	SqE7
	SqF7
	SqG7
	SqH7
	SqA8
	SqB8
	SqC8
	SqD8
	SqE8
	SqF8
	SqG8
	SqH8
	SqNone Square = -1
)

// File returns the file (0-7 for a-h) of the square.
func (sq Square) File() int { return int(sq % 8) }

// Rank returns the rank (0-7 for 1-8) of the square.
func (sq Square) Rank() int { return int(sq / 8) }

// String returns the algebraic notation of the square (e.g., "e4").
func (sq Square) String() string {
	if sq < 0 || sq > 63 {
		return "-"
	}
	return string([]byte{byte('a' + sq%8), byte('1' + sq/8)})
}

// Color represents the side to move (White or Black).
type Color int

const (
	White Color = iota
	Black
)

// String returns "w" for White, "b" for Black.
func (c Color) String() string {
	if c == White {
		return "w"
	}
	return "b"
}

// fileMask returns a bitboard with all squares on the given file set.
func fileMask(file int) Bitboard { return fileMasks[file] }

// rankMask returns a bitboard with all squares on the given rank set.
func rankMask(rank int) Bitboard { return rankMasks[rank] }

// File masks (vertical lines a-h)
var fileMasks = [8]Bitboard{
	0x0101010101010101,
	0x0202020202020202,
	0x0404040404040404,
	0x0808080808080808,
	0x1010101010101010,
	0x2020202020202020,
	0x4040404040404040,
	0x8080808080808080,
}

// Rank masks (horizontal lines 1-8)
var rankMasks = [8]Bitboard{
	0x00000000000000FF,
	0x000000000000FF00,
	0x0000000000FF0000,
	0x00000000FF000000,
	0x000000FF00000000,
	0x0000FF0000000000,
	0x00FF000000000000,
	0xFF00000000000000,
}

var knightAttacksV2 [64]Bitboard
var kingAttacksV2 [64]Bitboard

func init() {
	for sq := 0; sq < 64; sq++ {
		knightAttacksV2[sq] = maskKnightV2(Square(sq))
		kingAttacksV2[sq] = maskKingV2(Square(sq))
	}
}

func maskKnightV2(sq Square) Bitboard {
	r := sq / 8
	f := sq % 8
	var bb Bitboard
	deltas := [8][2]int{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
	for _, d := range deltas {
		r2 := int(r) + d[0]
		f2 := int(f) + d[1]
		if r2 >= 0 && r2 < 8 && f2 >= 0 && f2 < 8 {
			bb |= 1 << uint(r2*8+f2)
		}
	}
	return bb
}

func maskKingV2(sq Square) Bitboard {
	r := sq / 8
	f := sq % 8
	var bb Bitboard
	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			if dr == 0 && df == 0 {
				continue
			}
			r2 := int(r) + dr
			f2 := int(f) + df
			if r2 >= 0 && r2 < 8 && f2 >= 0 && f2 < 8 {
				bb |= 1 << uint(r2*8+f2)
			}
		}
	}
	return bb
}

// knightAttacks returns the precomputed knight attack bitboard from a square.
func knightAttacks(sq Square) Bitboard { return knightAttacksV2[sq] }

// kingAttacks returns the precomputed king attack bitboard from a square.
func kingAttacks(sq Square) Bitboard { return kingAttacksV2[sq] }

// ParseSquare parses algebraic notation (e.g., "e4") to a Square.
func ParseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return SqNone, fmt.Errorf("invalid square: %q", s)
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return SqNone, fmt.Errorf("invalid square: %q", s)
	}
	return Square(rank*8 + file), nil
}

// MakeSquare creates a Square from file (0-7) and rank (0-7).
func MakeSquare(file, rank int) Square {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return SqNone
	}
	return Square(rank*8 + file)
}
