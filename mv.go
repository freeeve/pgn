// mv.go defines the compact move type used by the engine.

package pgn

import "fmt"

// Mv represents a chess move with from/to squares and optional promotion.
type Mv struct {
	From  Square
	To    Square
	Promo PromoPiece
	Flags uint16
}

// String returns the move in UCI notation (e.g., "e2e4", "e7e8q").
func (m Mv) String() string {
	s := m.From.String() + m.To.String()
	switch m.Promo {
	case PromoQueen:
		s += "q"
	case PromoRook:
		s += "r"
	case PromoBishop:
		s += "b"
	case PromoKnight:
		s += "n"
	}
	return s
}

// PromoPiece represents the piece type for pawn promotion.
type PromoPiece int

const (
	NoPromo PromoPiece = iota
	PromoQueen
	PromoRook
	PromoBishop
	PromoKnight
)

// ParseUCI parses a move in UCI notation (e.g., "e2e4", "e7e8q").
// Returns an error if the string is not a valid UCI move format.
// Note: This only parses the format; it does not validate legality.
func ParseUCI(uci string) (Mv, error) {
	if len(uci) < 4 || len(uci) > 5 {
		return Mv{}, fmt.Errorf("invalid UCI move length: %q", uci)
	}

	from, err := ParseSquare(uci[0:2])
	if err != nil {
		return Mv{}, fmt.Errorf("invalid from square in UCI move: %w", err)
	}

	to, err := ParseSquare(uci[2:4])
	if err != nil {
		return Mv{}, fmt.Errorf("invalid to square in UCI move: %w", err)
	}

	mv := Mv{From: from, To: to}

	if len(uci) == 5 {
		switch uci[4] {
		case 'q', 'Q':
			mv.Promo = PromoQueen
		case 'r', 'R':
			mv.Promo = PromoRook
		case 'b', 'B':
			mv.Promo = PromoBishop
		case 'n', 'N':
			mv.Promo = PromoKnight
		default:
			return Mv{}, fmt.Errorf("invalid promotion piece: %c", uci[4])
		}
	}

	return mv, nil
}
