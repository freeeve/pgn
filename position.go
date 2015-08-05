package pgn

import "fmt"

// Position represents one or more locations on the board.  Each location has
// its own bit, in rank-major ascending order. Parts of the API which accept or
// return a Position may indicate whether more than one set bit is expected.
//
// In addition to the square constants, File and Rank constants can be binary
// AND'd together to refer to a single square, such that FileC & Rank3 == C3
type Position uint64

const (
	A1 Position = 1 << iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8

	NoPosition Position = 0
)

const (
	Rank1 Position = 0xff << (8 * iota)
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

const (
	FileA Position = 0x0101010101010101 << iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

// PositionFromOrd returns the position representing a single board square.
// file and rank are expected to be in the range 0..7 (inclusive).
func PositionFromOrd(file, rank int) Position {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoPosition
	}
	return Position(1) << uint(rank*8+file)
}

// ParsePosition accepts a one or two byte string, and returns a position
// representing a file, a rank, or a fully specific square. Inputs are treated
// case-insensitively.
func ParsePosition(s string) (Position, error) {
	p, ok := parsePosition(s)
	if !ok {
		return 0, fmt.Errorf("pgn: invalid position string: %s", s)
	}
	return p, nil
}

func parsePosition(s string) (p Position, ok bool) {
	if len(s) == 0 || len(s) > 2 {
		return 0, false
	}
	rankonly := false

	// assign a whole rank or file to start with (or if only one byte provided)
	switch b := s[0]; b {
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H':
		b += 'a' - 'A'
		fallthrough
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h':
		p = Position(1 << (b - 'a')).File()
	case '1', '2', '3', '4', '5', '6', '7', '8':
		p = Position(1 << (8 * (b - '1'))).Rank()
		rankonly = true
	default:
		return 0, false
	}

	if len(s) == 1 {
		return p, true
	} else if rankonly {
		return 0, false
	}

	// now refine with a rank
	switch b := s[1]; b {
	case '1', '2', '3', '4', '5', '6', '7', '8':
		p &= Position(1 << (8 * (b - '1'))).Rank()
	default:
		return 0, false
	}

	return p, true
}

// String returns an appropriate string representation for p. If p does not
// represent exactly onboard square, the output should not be relied upon.
func (p Position) String() string {
	if p == NoPosition {
		return "-"
	} else if !p.IsSquare() {
		return fmt.Sprintf("<%x>", uint64(p))
	}
	return string([]byte{p.FileSym(), p.RankSym()})
}

// IsSquare returns true if and only if p represents a single board square.
func (p Position) IsSquare() bool {
	return 0 == p&(p-1) && p != 0
}

// File fills all bits in every file where any bit is set. If only checking if a position
// corresponds to some file, use `p & FileX != 0` where X is a-f.
func (p Position) File() Position {
	p |= p>>8 | p>>16 | p>>24 | p>>32 | p>>40 | p>>48 | p>>56
	p &= 0xff
	p |= p<<8 | p<<16 | p<<24 | p<<32 | p<<40 | p<<48 | p<<56
	return p
}

// Rank fills all bits in every rank where any bit is set. If only checking if a position
// corresponds to some rank, use `p & RankX != 0` where X is 1-8.
func (p Position) Rank() Position {
	// following math will 0xff every byte that isn't 0x00
	const m = ^Position(0) / 255
	p |= p>>1 | p>>2 | p>>3 | p>>4 | p>>5 | p>>6 | p>>7
	return (m & p) * 255
}

// FileOrd returns 0-7 for positions that correspond to a single square, -1 otherwise.
func (p Position) FileOrd() int {
	if !p.IsSquare() {
		return -1
	}
	switch p.File() {
	case FileA:
		return 0
	case FileB:
		return 1
	case FileC:
		return 2
	case FileD:
		return 3
	case FileE:
		return 4
	case FileF:
		return 5
	case FileG:
		return 6
	case FileH:
		return 7
	}
	panic("unreachable")
}

// FileSym returns the human-readable byte value that most appropriately
// represents the File of p. If p represents more than one file, the output
// should not be relied upon.
func (p Position) FileSym() byte {
	ord := p.FileOrd()
	if ord < 0 {
		return '-'
	}
	return byte('a' + ord)
}

// RankOrd returns 0-7 for positions that correspond to a single square, -1 otherwise.
func (p Position) RankOrd() int {
	if !p.IsSquare() {
		return -1
	}
	switch p.Rank() {
	case Rank1:
		return 0
	case Rank2:
		return 1
	case Rank3:
		return 2
	case Rank4:
		return 3
	case Rank5:
		return 4
	case Rank6:
		return 5
	case Rank7:
		return 6
	case Rank8:
		return 7
	}
	panic("unreachable")
}

// RankSym returns the human-readable byte value that most appropriately
// represents the Rank of p. If p represents more than one rank, the output
// should not be relied upon.
func (p Position) RankSym() byte {
	ord := p.RankOrd()
	if ord < 0 {
		return '-'
	}
	return byte('1' + ord)
}
