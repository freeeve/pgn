// apply.go provides functions to apply and undo chess moves on a GameState.

package pgn

import "fmt"

// UndoInfo stores the state needed to unmake a move.
type UndoInfo struct {
	movedPiece    int
	capturedPiece int
	capturedSq    Square
	castle        uint8
	ep            Square
	halfmove      int
}

// MakeMove applies a move in-place and returns undo info for UnmakeMove.
// This is the fast path for perft and search where we need to undo moves.
func MakeMove(pos *GameState, m Mv) UndoInfo {
	undo := UndoInfo{
		movedPiece:    -1,
		capturedPiece: -1,
		capturedSq:    -1,
		castle:        pos.Castle,
		ep:            pos.EP,
		halfmove:      pos.Halfmove,
	}

	fromMask := Bitboard(1) << uint(m.From)
	toMask := Bitboard(1) << uint(m.To)

	// Find and remove moving piece
	for i := 0; i < v2PieceCount; i++ {
		if pos.pieces[i]&fromMask != 0 {
			undo.movedPiece = i
			pos.pieces[i] &^= fromMask
			break
		}
	}

	// Handle captures
	if m.Flags&2 != 0 && (undo.movedPiece == v2WPawn || undo.movedPiece == v2BPawn) {
		// En passant: captured pawn is on a different square than the target
		if undo.movedPiece == v2WPawn {
			undo.capturedSq = m.To - 8
		} else {
			undo.capturedSq = m.To + 8
		}
		capMask := Bitboard(1) << uint(undo.capturedSq)
		for i := 0; i < v2PieceCount; i++ {
			if pos.pieces[i]&capMask != 0 {
				undo.capturedPiece = i
				pos.pieces[i] &^= capMask
				break
			}
		}
	} else {
		for i := 0; i < v2PieceCount; i++ {
			if pos.pieces[i]&toMask != 0 {
				undo.capturedPiece = i
				undo.capturedSq = m.To
				pos.pieces[i] &^= toMask
				break
			}
		}
	}

	// Determine final piece (handle promotion)
	finalPiece := undo.movedPiece
	if m.Promo != NoPromo {
		if pos.SideToMove == White {
			switch m.Promo {
			case PromoQueen:
				finalPiece = v2WQueen
			case PromoRook:
				finalPiece = v2WRook
			case PromoBishop:
				finalPiece = v2WBishop
			case PromoKnight:
				finalPiece = v2WKnight
			}
		} else {
			switch m.Promo {
			case PromoQueen:
				finalPiece = v2BQueen
			case PromoRook:
				finalPiece = v2BRook
			case PromoBishop:
				finalPiece = v2BBishop
			case PromoKnight:
				finalPiece = v2BKnight
			}
		}
	}

	pos.pieces[finalPiece] |= toMask

	// Update castling rights
	switch undo.movedPiece {
	case v2WKing:
		pos.Castle &^= (1 << 0) | (1 << 1)
	case v2BKing:
		pos.Castle &^= (1 << 2) | (1 << 3)
	case v2WRook:
		if m.From == 0 {
			pos.Castle &^= 1 << 1
		} else if m.From == 7 {
			pos.Castle &^= 1 << 0
		}
	case v2BRook:
		if m.From == 56 {
			pos.Castle &^= 1 << 3
		} else if m.From == 63 {
			pos.Castle &^= 1 << 2
		}
	}

	// Clear castling rights if rook is captured
	if undo.capturedPiece == v2WRook {
		if m.To == 0 {
			pos.Castle &^= 1 << 1
		} else if m.To == 7 {
			pos.Castle &^= 1 << 0
		}
	} else if undo.capturedPiece == v2BRook {
		if m.To == 56 {
			pos.Castle &^= 1 << 3
		} else if m.To == 63 {
			pos.Castle &^= 1 << 2
		}
	}

	// Handle castling: move the rook
	if m.Flags&4 != 0 {
		switch {
		case m.From == 4 && m.To == 6:
			pos.pieces[v2WRook] &^= 1 << 7
			pos.pieces[v2WRook] |= 1 << 5
		case m.From == 4 && m.To == 2:
			pos.pieces[v2WRook] &^= 1 << 0
			pos.pieces[v2WRook] |= 1 << 3
		case m.From == 60 && m.To == 62:
			pos.pieces[v2BRook] &^= 1 << 63
			pos.pieces[v2BRook] |= 1 << 61
		case m.From == 60 && m.To == 58:
			pos.pieces[v2BRook] &^= 1 << 56
			pos.pieces[v2BRook] |= 1 << 59
		}
	}

	// Update occupancy incrementally
	us := pos.SideToMove
	them := White
	if us == White {
		them = Black
	}
	pos.occ[us] &^= fromMask
	pos.occ[us] |= toMask
	if undo.capturedPiece >= 0 {
		capMask := Bitboard(1) << uint(undo.capturedSq)
		pos.occ[them] &^= capMask
	}

	// Castling rook occupancy
	if m.Flags&4 != 0 {
		switch {
		case m.From == 4 && m.To == 6:
			pos.occ[White] &^= 1 << 7
			pos.occ[White] |= 1 << 5
		case m.From == 4 && m.To == 2:
			pos.occ[White] &^= 1 << 0
			pos.occ[White] |= 1 << 3
		case m.From == 60 && m.To == 62:
			pos.occ[Black] &^= 1 << 63
			pos.occ[Black] |= 1 << 61
		case m.From == 60 && m.To == 58:
			pos.occ[Black] &^= 1 << 56
			pos.occ[Black] |= 1 << 59
		}
	}
	pos.occAll = pos.occ[White] | pos.occ[Black]

	// Update EP square (only set for double pawn pushes)
	pos.EP = -1
	if undo.movedPiece == v2WPawn && m.Flags == 1 {
		pos.EP = m.To - 8
	} else if undo.movedPiece == v2BPawn && m.Flags == 1 {
		pos.EP = m.To + 8
	}

	// Update halfmove clock
	if undo.capturedPiece >= 0 || undo.movedPiece == v2WPawn || undo.movedPiece == v2BPawn {
		pos.Halfmove = 0
	} else {
		pos.Halfmove++
	}

	// Toggle side to move
	if pos.SideToMove == White {
		pos.SideToMove = Black
	} else {
		pos.SideToMove = White
		pos.Fullmove++
	}

	return undo
}

// UnmakeMove reverses a move using the undo info from MakeMove.
func UnmakeMove(pos *GameState, m Mv, undo UndoInfo) {
	// Toggle side back
	if pos.SideToMove == White {
		pos.SideToMove = Black
		pos.Fullmove--
	} else {
		pos.SideToMove = White
	}

	fromMask := Bitboard(1) << uint(m.From)
	toMask := Bitboard(1) << uint(m.To)

	// Remove piece from target (could be promoted piece)
	for i := 0; i < v2PieceCount; i++ {
		if pos.pieces[i]&toMask != 0 {
			pos.pieces[i] &^= toMask
			break
		}
	}

	// Restore moving piece to origin
	pos.pieces[undo.movedPiece] |= fromMask

	// Restore captured piece
	if undo.capturedPiece >= 0 {
		capMask := Bitboard(1) << uint(undo.capturedSq)
		pos.pieces[undo.capturedPiece] |= capMask
	}

	// Undo castling rook move
	if m.Flags&4 != 0 {
		switch {
		case m.From == 4 && m.To == 6:
			pos.pieces[v2WRook] |= 1 << 7
			pos.pieces[v2WRook] &^= 1 << 5
		case m.From == 4 && m.To == 2:
			pos.pieces[v2WRook] |= 1 << 0
			pos.pieces[v2WRook] &^= 1 << 3
		case m.From == 60 && m.To == 62:
			pos.pieces[v2BRook] |= 1 << 63
			pos.pieces[v2BRook] &^= 1 << 61
		case m.From == 60 && m.To == 58:
			pos.pieces[v2BRook] |= 1 << 56
			pos.pieces[v2BRook] &^= 1 << 59
		}
	}

	// Restore state
	pos.Castle = undo.castle
	pos.EP = undo.ep
	pos.Halfmove = undo.halfmove

	// Rebuild occupancy
	pos.occ[White] = pos.pieces[v2WPawn] | pos.pieces[v2WKnight] | pos.pieces[v2WBishop] |
		pos.pieces[v2WRook] | pos.pieces[v2WQueen] | pos.pieces[v2WKing]
	pos.occ[Black] = pos.pieces[v2BPawn] | pos.pieces[v2BKnight] | pos.pieces[v2BBishop] |
		pos.pieces[v2BRook] | pos.pieces[v2BQueen] | pos.pieces[v2BKing]
	pos.occAll = pos.occ[White] | pos.occ[Black]
}

// ApplyMove applies a move to the position without undo capability.
// This is simpler but slower than MakeMove/UnmakeMove for search.
func ApplyMove(pos *GameState, m Mv) error {
	if pos == nil {
		return fmt.Errorf("nil position")
	}
	fromMask := Bitboard(1) << uint(m.From)
	toMask := Bitboard(1) << uint(m.To)

	// Find moving piece
	var movingIdx int = -1
	for i := 0; i < v2PieceCount; i++ {
		if pos.pieces[i]&fromMask != 0 {
			movingIdx = i
			pos.pieces[i] &^= fromMask
			break
		}
	}
	if movingIdx == -1 {
		return fmt.Errorf("no piece on from square")
	}

	capturedIdx := -1

	// Handle en passant capture
	if m.Flags&2 != 0 && (movingIdx == v2WPawn || movingIdx == v2BPawn) {
		if movingIdx == v2WPawn {
			capSq := Square(int(m.To) - 8)
			capMask := Bitboard(1) << uint(capSq)
			for i := v2BPawn; i < v2PieceCount; i++ {
				if pos.pieces[i]&capMask != 0 {
					pos.pieces[i] &^= capMask
					capturedIdx = i
					break
				}
			}
		} else {
			capSq := Square(int(m.To) + 8)
			capMask := Bitboard(1) << uint(capSq)
			for i := 0; i < v2BPawn; i++ {
				if pos.pieces[i]&capMask != 0 {
					pos.pieces[i] &^= capMask
					capturedIdx = i
					break
				}
			}
		}
	} else {
		// Normal capture
		for i := 0; i < v2PieceCount; i++ {
			if pos.pieces[i]&toMask != 0 {
				pos.pieces[i] &^= toMask
				capturedIdx = i
				break
			}
		}
	}

	// Handle promotion
	if m.Promo != NoPromo {
		switch pos.SideToMove {
		case White:
			switch m.Promo {
			case PromoQueen:
				movingIdx = v2WQueen
			case PromoRook:
				movingIdx = v2WRook
			case PromoBishop:
				movingIdx = v2WBishop
			case PromoKnight:
				movingIdx = v2WKnight
			}
		case Black:
			switch m.Promo {
			case PromoQueen:
				movingIdx = v2BQueen
			case PromoRook:
				movingIdx = v2BRook
			case PromoBishop:
				movingIdx = v2BBishop
			case PromoKnight:
				movingIdx = v2BKnight
			}
		}
	}

	pos.pieces[movingIdx] |= toMask

	// Update castling rights
	switch movingIdx {
	case v2WKing:
		pos.Castle &^= (1 << 0) | (1 << 1)
	case v2BKing:
		pos.Castle &^= (1 << 2) | (1 << 3)
	case v2WRook:
		if m.From == 0 {
			pos.Castle &^= 1 << 1
		}
		if m.From == 7 {
			pos.Castle &^= 1 << 0
		}
	case v2BRook:
		if m.From == 56 {
			pos.Castle &^= 1 << 3
		}
		if m.From == 63 {
			pos.Castle &^= 1 << 2
		}
	}

	// Clear castling rights if rook captured
	if capturedIdx == v2WRook {
		if m.To == 0 {
			pos.Castle &^= 1 << 1
		}
		if m.To == 7 {
			pos.Castle &^= 1 << 0
		}
	}
	if capturedIdx == v2BRook {
		if m.To == 56 {
			pos.Castle &^= 1 << 3
		}
		if m.To == 63 {
			pos.Castle &^= 1 << 2
		}
	}

	// Handle castling: move the rook
	if m.Flags&4 != 0 {
		switch {
		case m.From == 4 && m.To == 6:
			pos.pieces[v2WRook] &^= 1 << 7
			pos.pieces[v2WRook] |= 1 << 5
			pos.Castle &^= (1<<0 | 1<<1)
		case m.From == 4 && m.To == 2:
			pos.pieces[v2WRook] &^= 1 << 0
			pos.pieces[v2WRook] |= 1 << 3
			pos.Castle &^= (1<<0 | 1<<1)
		case m.From == 60 && m.To == 62:
			pos.pieces[v2BRook] &^= 1 << 63
			pos.pieces[v2BRook] |= 1 << 61
			pos.Castle &^= (1<<2 | 1<<3)
		case m.From == 60 && m.To == 58:
			pos.pieces[v2BRook] &^= 1 << 56
			pos.pieces[v2BRook] |= 1 << 59
			pos.Castle &^= (1<<2 | 1<<3)
		}
	}

	// Recompute occupancy
	pos.occ[White] = 0
	pos.occ[Black] = 0
	for i := 0; i < v2PieceCount; i++ {
		if i < v2BPawn {
			pos.occ[White] |= pos.pieces[i]
		} else {
			pos.occ[Black] |= pos.pieces[i]
		}
	}
	pos.occAll = pos.occ[White] | pos.occ[Black]

	// Toggle side to move
	if pos.SideToMove == White {
		pos.SideToMove = Black
	} else {
		pos.SideToMove = White
		pos.Fullmove++
	}

	// Update EP square
	pos.EP = -1
	if movingIdx == v2WPawn && m.Flags == 1 {
		pos.EP = m.To - 8
	} else if movingIdx == v2BPawn && m.Flags == 1 {
		pos.EP = m.To + 8
	}

	// Update halfmove clock
	if m.Promo != NoPromo || movingIdx == v2WPawn || movingIdx == v2BPawn || (toMask&pos.occ[pos.SideToMove]) != 0 {
		pos.Halfmove = 0
	} else {
		pos.Halfmove++
	}

	return nil
}
