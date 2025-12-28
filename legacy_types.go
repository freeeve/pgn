// legacy_types.go provides backward-compatible types for migration.
//
// These types (Position, Piece, Move) use the old bitmask representation.
// New code should use Square, byte, and Mv respectively.

package pgn

import "fmt"

// Position is a legacy type representing a square as a bitmask.
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

// Piece is a legacy type representing a chess piece as a character.
type Piece byte

const (
	NoPiece     Piece = ' '
	BlackPawn   Piece = 'p'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackRook   Piece = 'r'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
	WhitePawn   Piece = 'P'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteRook   Piece = 'R'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
)

func (p Piece) String() string {
	return string(p)
}

// Move is the legacy move type using bitmask positions.
type Move struct {
	From    Position
	To      Position
	Promote Piece
	San     string
}

func (m Move) String() string {
	from := positionToAlg(m.From)
	to := positionToAlg(m.To)
	if m.Promote == NoPiece {
		return from + to
	}
	return from + to + string(m.Promote)
}

var NilMove = Move{From: NoPosition, To: NoPosition}

func positionToAlg(p Position) string {
	if p == NoPosition {
		return "-"
	}
	for sq := 0; sq < 64; sq++ {
		if p == (1 << sq) {
			file := byte('a' + (sq % 8))
			rank := byte('1' + (sq / 8))
			return string([]byte{file, rank})
		}
	}
	return "-"
}

// SquareToPosition converts a Square to a Position bitmask.
func SquareToPosition(sq Square) Position {
	if sq < 0 || sq > 63 {
		return NoPosition
	}
	return Position(1) << uint(sq)
}

// PositionToSquare converts a Position bitmask to a Square.
func PositionToSquare(p Position) Square {
	if p == NoPosition {
		return SqNone
	}
	for sq := 0; sq < 64; sq++ {
		if p == (1 << sq) {
			return Square(sq)
		}
	}
	return SqNone
}

// MvToMove converts a Mv to a legacy Move.
func MvToMove(mv Mv) Move {
	var promo Piece = NoPiece
	switch mv.Promo {
	case PromoQueen:
		promo = WhiteQueen
	case PromoRook:
		promo = WhiteRook
	case PromoBishop:
		promo = WhiteBishop
	case PromoKnight:
		promo = WhiteKnight
	}
	return Move{
		From:    SquareToPosition(mv.From),
		To:      SquareToPosition(mv.To),
		Promote: promo,
	}
}

// MoveToMv converts a legacy Move to Mv.
func MoveToMv(m Move) Mv {
	var promo PromoPiece = NoPromo
	switch m.Promote {
	case WhiteQueen, BlackQueen:
		promo = PromoQueen
	case WhiteRook, BlackRook:
		promo = PromoRook
	case WhiteBishop, BlackBishop:
		promo = PromoBishop
	case WhiteKnight, BlackKnight:
		promo = PromoKnight
	}
	return Mv{
		From:  PositionToSquare(m.From),
		To:    PositionToSquare(m.To),
		Promo: promo,
	}
}

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
