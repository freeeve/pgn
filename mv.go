// mv.go defines the compact move type used by the engine.

package pgn

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
