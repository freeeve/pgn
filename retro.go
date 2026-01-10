// retro.go provides backwards move generation for finding parent positions.
//
// Given a chess position, GenerateParentPositions returns all candidate parent
// positions along with the forward move that would reach the current position.
// This is useful for searching position ancestry in a position store.
//
// The generated positions are candidates that should be validated by:
// 1. Looking them up in your position store
// 2. Verifying the forward move is legal using GenerateLegalMoves

package pgn

// ParentCandidate represents a candidate parent position with the connecting move.
type ParentCandidate struct {
	Position PackedPosition
	Move     Mv // The forward move from parent to current position
}

// GenerateParentPositions generates all candidate parent positions.
// For each candidate, Move indicates the forward move from parent to current position.
//
// This generates pseudo-legal retro-moves and should be validated by:
// 1. Looking up Position in your store
// 2. If found, verifying Move is in GenerateLegalMoves(parent)
func GenerateParentPositions(pos *GameState) []ParentCandidate {
	if pos == nil {
		return nil
	}

	var results []ParentCandidate

	// The side that made the previous move is opposite of current side to move
	mover := pos.SideToMove ^ 1

	// Generate retro-moves for each piece type
	if mover == White {
		results = genRetroWhitePawns(results, pos)
		results = genRetroKnights(results, pos, v2WKnight)
		results = genRetroBishops(results, pos, v2WBishop)
		results = genRetroRooks(results, pos, v2WRook)
		results = genRetroQueens(results, pos, v2WQueen)
		results = genRetroKing(results, pos, v2WKing)
		results = genRetroCastling(results, pos, true)
	} else {
		results = genRetroBlackPawns(results, pos)
		results = genRetroKnights(results, pos, v2BKnight)
		results = genRetroBishops(results, pos, v2BBishop)
		results = genRetroRooks(results, pos, v2BRook)
		results = genRetroQueens(results, pos, v2BQueen)
		results = genRetroKing(results, pos, v2BKing)
		results = genRetroCastling(results, pos, false)
	}

	// Handle en passant: if EP square is set, the previous move MUST have been
	// a double pawn push. This is handled in genRetro*Pawns.

	return results
}

// genRetroWhitePawns generates parent positions for white pawn moves.
func genRetroWhitePawns(results []ParentCandidate, pos *GameState) []ParentCandidate {
	pawns := pos.pieces[v2WPawn]

	for pawns != 0 {
		to := Square(bitscanForward(pawns))
		pawns &= pawns - 1

		toRank := to.Rank()

		// Pawns can't be on rank 1 (they'd be black pieces or promoted)
		if toRank == 0 {
			continue
		}

		// Single push: pawn came from one rank below
		if toRank >= 2 { // Can't single-push to rank 2 (would mean pawn was on rank 1)
			from := to - 8
			// The 'from' square must be empty in current position (pawn moved away)
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroMove(results, pos, from, to, v2WPawn, false, 0, NoPromo)
			}
		}

		// Double push: pawn came from rank 2 to rank 4
		if toRank == 3 {
			from := to - 16
			// Both intermediate and from square must be empty
			intermediate := to - 8
			if pos.occAll&(1<<uint(from)) == 0 && pos.occAll&(1<<uint(intermediate)) == 0 {
				// Double push sets EP square, which must match current EP
				// Actually in retro, if current position has EP set on this file,
				// the previous move was this double push
				results = addRetroMove(results, pos, from, to, v2WPawn, false, 1, NoPromo)
			}
		}

		// Captures: pawn came from diagonal
		// Left capture: from = to - 7 (unless on a-file)
		if to.File() > 0 && toRank >= 2 {
			from := to - 7
			if pos.occAll&(1<<uint(from)) == 0 {
				// There was a capture, try each black piece type
				results = addRetroCaptureVariants(results, pos, from, to, v2WPawn, Black, NoPromo)
			}
		}

		// Right capture: from = to - 9 (unless on h-file)
		if to.File() < 7 && toRank >= 2 {
			from := to - 9
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroCaptureVariants(results, pos, from, to, v2WPawn, Black, NoPromo)
			}
		}
	}

	// Un-promotions: Q/R/B/N on rank 8 could have been a promoted pawn
	rank8 := rankMasks[7]
	for _, pieceIdx := range []int{v2WQueen, v2WRook, v2WBishop, v2WKnight} {
		pieces := pos.pieces[pieceIdx] & rank8
		for pieces != 0 {
			to := Square(bitscanForward(pieces))
			pieces &= pieces - 1

			promo := pieceIdxToPromo(pieceIdx)

			// Promotion by push: from = to - 8
			from := to - 8
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroPromotion(results, pos, from, to, pieceIdx, promo)
			}

			// Promotion by capture left: from = to - 7
			if to.File() > 0 {
				from = to - 7
				if pos.occAll&(1<<uint(from)) == 0 {
					results = addRetroPromotionCapture(results, pos, from, to, pieceIdx, promo, Black)
				}
			}

			// Promotion by capture right: from = to - 9
			if to.File() < 7 {
				from = to - 9
				if pos.occAll&(1<<uint(from)) == 0 {
					results = addRetroPromotionCapture(results, pos, from, to, pieceIdx, promo, Black)
				}
			}
		}
	}

	// En passant captures: if current position has an EP square set and there's
	// a black pawn that could have been captured EP
	if pos.EP >= 0 {
		// Current EP square means Black just did double push in parent position
		// So we're looking for retro-EP where White captured
		// This is complex: in parent, black pawn was on EP-8 (rank 5),
		// and white pawn captured to EP square
		// But wait - if we have EP set, it means the CURRENT side to move can EP.
		// So if white to move, black did double push. We're generating retro for black.
		// Actually this case is handled by genRetroBlackPawns, not here.
	}

	return results
}

// genRetroBlackPawns generates parent positions for black pawn moves.
func genRetroBlackPawns(results []ParentCandidate, pos *GameState) []ParentCandidate {
	pawns := pos.pieces[v2BPawn]

	for pawns != 0 {
		to := Square(bitscanForward(pawns))
		pawns &= pawns - 1

		toRank := to.Rank()

		// Pawns can't be on rank 8 (they'd be promoted)
		if toRank == 7 {
			continue
		}

		// Single push: pawn came from one rank above
		if toRank <= 5 { // Can't single-push to rank 7 (would mean pawn was on rank 8)
			from := to + 8
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroMove(results, pos, from, to, v2BPawn, false, 0, NoPromo)
			}
		}

		// Double push: pawn came from rank 7 to rank 5
		if toRank == 4 {
			from := to + 16
			intermediate := to + 8
			if pos.occAll&(1<<uint(from)) == 0 && pos.occAll&(1<<uint(intermediate)) == 0 {
				results = addRetroMove(results, pos, from, to, v2BPawn, false, 1, NoPromo)
			}
		}

		// Captures
		// Left capture: from = to + 9 (unless on a-file)
		if to.File() > 0 && toRank <= 5 {
			from := to + 9
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroCaptureVariants(results, pos, from, to, v2BPawn, White, NoPromo)
			}
		}

		// Right capture: from = to + 7 (unless on h-file)
		if to.File() < 7 && toRank <= 5 {
			from := to + 7
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroCaptureVariants(results, pos, from, to, v2BPawn, White, NoPromo)
			}
		}
	}

	// Un-promotions: Q/R/B/N on rank 1 could have been a promoted pawn
	rank1 := rankMasks[0]
	for _, pieceIdx := range []int{v2BQueen, v2BRook, v2BBishop, v2BKnight} {
		pieces := pos.pieces[pieceIdx] & rank1
		for pieces != 0 {
			to := Square(bitscanForward(pieces))
			pieces &= pieces - 1

			promo := pieceIdxToPromo(pieceIdx)

			// Promotion by push: from = to + 8
			from := to + 8
			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroPromotion(results, pos, from, to, pieceIdx, promo)
			}

			// Promotion by capture left: from = to + 9
			if to.File() > 0 {
				from = to + 9
				if pos.occAll&(1<<uint(from)) == 0 {
					results = addRetroPromotionCapture(results, pos, from, to, pieceIdx, promo, White)
				}
			}

			// Promotion by capture right: from = to + 7
			if to.File() < 7 {
				from = to + 7
				if pos.occAll&(1<<uint(from)) == 0 {
					results = addRetroPromotionCapture(results, pos, from, to, pieceIdx, promo, White)
				}
			}
		}
	}

	return results
}

// genRetroKnights generates parent positions for knight moves.
func genRetroKnights(results []ParentCandidate, pos *GameState, pieceIdx int) []ParentCandidate {
	knights := pos.pieces[pieceIdx]
	capturedColor := White
	if pieceIdx < v2BPawn {
		capturedColor = Black
	}

	for knights != 0 {
		to := Square(bitscanForward(knights))
		knights &= knights - 1

		// Knight could have come from any of its attack squares
		froms := knightAttacks(to)
		for froms != 0 {
			from := Square(bitscanForward(froms))
			froms &= froms - 1

			if pos.occAll&(1<<uint(from)) == 0 {
				// Non-capture
				results = addRetroMove(results, pos, from, to, pieceIdx, false, 0, NoPromo)
				// Captures
				results = addRetroCaptureVariants(results, pos, from, to, pieceIdx, capturedColor, NoPromo)
			}
		}
	}

	return results
}

// genRetroBishops generates parent positions for bishop moves.
func genRetroBishops(results []ParentCandidate, pos *GameState, pieceIdx int) []ParentCandidate {
	bishops := pos.pieces[pieceIdx]
	capturedColor := White
	if pieceIdx < v2BPawn {
		capturedColor = Black
	}

	for bishops != 0 {
		to := Square(bitscanForward(bishops))
		bishops &= bishops - 1

		// Bishop could have come from diagonal squares (with clear path in current position)
		// We need to mask out the bishop's own square when computing attacks
		occWithoutPiece := pos.occAll &^ (1 << uint(to))
		froms := bishopAttacks(to, occWithoutPiece)

		for froms != 0 {
			from := Square(bitscanForward(froms))
			froms &= froms - 1

			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroMove(results, pos, from, to, pieceIdx, false, 0, NoPromo)
				results = addRetroCaptureVariants(results, pos, from, to, pieceIdx, capturedColor, NoPromo)
			}
		}
	}

	return results
}

// genRetroRooks generates parent positions for rook moves.
func genRetroRooks(results []ParentCandidate, pos *GameState, pieceIdx int) []ParentCandidate {
	rooks := pos.pieces[pieceIdx]
	capturedColor := White
	if pieceIdx < v2BPawn {
		capturedColor = Black
	}

	for rooks != 0 {
		to := Square(bitscanForward(rooks))
		rooks &= rooks - 1

		occWithoutPiece := pos.occAll &^ (1 << uint(to))
		froms := rookAttacks(to, occWithoutPiece)

		for froms != 0 {
			from := Square(bitscanForward(froms))
			froms &= froms - 1

			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroMove(results, pos, from, to, pieceIdx, false, 0, NoPromo)
				results = addRetroCaptureVariants(results, pos, from, to, pieceIdx, capturedColor, NoPromo)
			}
		}
	}

	return results
}

// genRetroQueens generates parent positions for queen moves.
func genRetroQueens(results []ParentCandidate, pos *GameState, pieceIdx int) []ParentCandidate {
	queens := pos.pieces[pieceIdx]
	capturedColor := White
	if pieceIdx < v2BPawn {
		capturedColor = Black
	}

	for queens != 0 {
		to := Square(bitscanForward(queens))
		queens &= queens - 1

		occWithoutPiece := pos.occAll &^ (1 << uint(to))
		froms := queenAttacks(to, occWithoutPiece)

		for froms != 0 {
			from := Square(bitscanForward(froms))
			froms &= froms - 1

			if pos.occAll&(1<<uint(from)) == 0 {
				results = addRetroMove(results, pos, from, to, pieceIdx, false, 0, NoPromo)
				results = addRetroCaptureVariants(results, pos, from, to, pieceIdx, capturedColor, NoPromo)
			}
		}
	}

	return results
}

// genRetroKing generates parent positions for king moves (non-castling).
func genRetroKing(results []ParentCandidate, pos *GameState, pieceIdx int) []ParentCandidate {
	kings := pos.pieces[pieceIdx]
	if kings == 0 {
		return results
	}

	capturedColor := White
	if pieceIdx < v2BPawn {
		capturedColor = Black
	}

	to := Square(bitscanForward(kings))
	froms := kingAttacks(to)

	for froms != 0 {
		from := Square(bitscanForward(froms))
		froms &= froms - 1

		if pos.occAll&(1<<uint(from)) == 0 {
			results = addRetroMove(results, pos, from, to, pieceIdx, false, 0, NoPromo)
			results = addRetroCaptureVariants(results, pos, from, to, pieceIdx, capturedColor, NoPromo)
		}
	}

	return results
}

// genRetroCastling generates parent positions for un-castling.
func genRetroCastling(results []ParentCandidate, pos *GameState, white bool) []ParentCandidate {
	if white {
		// White kingside: King on g1, Rook on f1 -> unmake to e1, h1
		if pos.pieces[v2WKing]&(1<<6) != 0 && pos.pieces[v2WRook]&(1<<5) != 0 {
			// Check if e1, h1 are empty (they should be after castling)
			if pos.occAll&(1<<4) == 0 && pos.occAll&(1<<7) == 0 {
				results = addRetroCastleMove(results, pos, 4, 6, true, true) // e1->g1 kingside
			}
		}
		// White queenside: King on c1, Rook on d1 -> unmake to e1, a1
		if pos.pieces[v2WKing]&(1<<2) != 0 && pos.pieces[v2WRook]&(1<<3) != 0 {
			if pos.occAll&(1<<4) == 0 && pos.occAll&(1<<0) == 0 {
				results = addRetroCastleMove(results, pos, 4, 2, true, false) // e1->c1 queenside
			}
		}
	} else {
		// Black kingside: King on g8, Rook on f8 -> unmake to e8, h8
		if pos.pieces[v2BKing]&(1<<62) != 0 && pos.pieces[v2BRook]&(1<<61) != 0 {
			if pos.occAll&(1<<60) == 0 && pos.occAll&(1<<63) == 0 {
				results = addRetroCastleMove(results, pos, 60, 62, false, true) // e8->g8 kingside
			}
		}
		// Black queenside: King on c8, Rook on d8 -> unmake to e8, a8
		if pos.pieces[v2BKing]&(1<<58) != 0 && pos.pieces[v2BRook]&(1<<59) != 0 {
			if pos.occAll&(1<<60) == 0 && pos.occAll&(1<<56) == 0 {
				results = addRetroCastleMove(results, pos, 60, 58, false, false) // e8->c8 queenside
			}
		}
	}

	return results
}

// addRetroMove adds a parent position for a non-capture move.
func addRetroMove(results []ParentCandidate, pos *GameState, from, to Square, pieceIdx int, _ bool, flags uint16, promo PromoPiece) []ParentCandidate {
	parent := buildParentPosition(pos, from, to, pieceIdx, -1, -1, flags, promo)
	move := Mv{From: from, To: to, Flags: flags, Promo: promo}
	return append(results, ParentCandidate{Position: parent, Move: move})
}

// addRetroCaptureVariants adds parent positions for all possible capture variants.
func addRetroCaptureVariants(results []ParentCandidate, pos *GameState, from, to Square, pieceIdx int, capturedColor Color, promo PromoPiece) []ParentCandidate {
	// Try each capturable piece type (not king)
	var captureTypes []int
	if capturedColor == White {
		captureTypes = []int{v2WPawn, v2WKnight, v2WBishop, v2WRook, v2WQueen}
	} else {
		captureTypes = []int{v2BPawn, v2BKnight, v2BBishop, v2BRook, v2BQueen}
	}

	for _, capType := range captureTypes {
		// Skip impossible captures (pawns on rank 1 or 8)
		if capType == v2WPawn && (to.Rank() == 0 || to.Rank() == 7) {
			continue
		}
		if capType == v2BPawn && (to.Rank() == 0 || to.Rank() == 7) {
			continue
		}

		parent := buildParentPosition(pos, from, to, pieceIdx, capType, to, 0, promo)
		move := Mv{From: from, To: to, Promo: promo}
		results = append(results, ParentCandidate{Position: parent, Move: move})
	}

	return results
}

// addRetroPromotion adds a parent position for a pawn promotion (push).
func addRetroPromotion(results []ParentCandidate, pos *GameState, from, to Square, promotedPieceIdx int, promo PromoPiece) []ParentCandidate {
	// In parent, there was a pawn at 'from', not the promoted piece at 'to'
	parent := buildParentPositionPromo(pos, from, to, promotedPieceIdx, promo)
	move := Mv{From: from, To: to, Promo: promo}
	return append(results, ParentCandidate{Position: parent, Move: move})
}

// addRetroPromotionCapture adds parent positions for promotion captures.
func addRetroPromotionCapture(results []ParentCandidate, pos *GameState, from, to Square, promotedPieceIdx int, promo PromoPiece, capturedColor Color) []ParentCandidate {
	var captureTypes []int
	if capturedColor == White {
		captureTypes = []int{v2WPawn, v2WKnight, v2WBishop, v2WRook, v2WQueen}
	} else {
		captureTypes = []int{v2BPawn, v2BKnight, v2BBishop, v2BRook, v2BQueen}
	}

	for _, capType := range captureTypes {
		// Pawns can't be on promotion ranks
		if capType == v2WPawn || capType == v2BPawn {
			continue // Pawns can't be on rank 1 or 8 to be captured
		}

		parent := buildParentPositionPromoCapture(pos, from, to, promotedPieceIdx, promo, capType)
		move := Mv{From: from, To: to, Promo: promo}
		results = append(results, ParentCandidate{Position: parent, Move: move})
	}

	return results
}

// addRetroCastleMove adds a parent position for un-castling.
func addRetroCastleMove(results []ParentCandidate, pos *GameState, kingFrom, kingTo Square, white, kingside bool) []ParentCandidate {
	parent := buildParentPositionCastle(pos, kingFrom, kingTo, white, kingside)
	move := Mv{From: kingFrom, To: kingTo, Flags: 4}
	return append(results, ParentCandidate{Position: parent, Move: move})
}

// buildParentPosition creates a PackedPosition for the parent state.
func buildParentPosition(pos *GameState, from, to Square, pieceIdx, capturedIdx int, capturedSq Square, flags uint16, promo PromoPiece) PackedPosition {
	// Create temporary game state for parent
	parent := *pos // shallow copy

	// Flip side to move
	parent.SideToMove ^= 1

	// Move piece from 'to' back to 'from'
	parent.pieces[pieceIdx] &^= 1 << uint(to)
	parent.pieces[pieceIdx] |= 1 << uint(from)

	// Restore captured piece if any
	if capturedIdx >= 0 && capturedSq >= 0 {
		parent.pieces[capturedIdx] |= 1 << uint(capturedSq)
	}

	// Rebuild occupancy
	parent.occ[White] = parent.pieces[v2WPawn] | parent.pieces[v2WKnight] | parent.pieces[v2WBishop] |
		parent.pieces[v2WRook] | parent.pieces[v2WQueen] | parent.pieces[v2WKing]
	parent.occ[Black] = parent.pieces[v2BPawn] | parent.pieces[v2BKnight] | parent.pieces[v2BBishop] |
		parent.pieces[v2BRook] | parent.pieces[v2BQueen] | parent.pieces[v2BKing]
	parent.occAll = parent.occ[White] | parent.occ[Black]

	// Handle EP square for double pawn push
	parent.EP = -1
	if flags == 1 {
		// Double pawn push - EP was set AFTER the move, so parent has no EP
		// The current position's EP (if any) was set by THIS move
	}

	// Restore potential castling rights
	// If king/rook moved FROM their starting squares, parent might have had those rights
	parent.Castle = expandCastlingRights(pos.Castle, from, pieceIdx)

	// Adjust fullmove counter
	if parent.SideToMove == Black {
		parent.Fullmove--
		if parent.Fullmove < 1 {
			parent.Fullmove = 1
		}
	}

	return parent.Pack()
}

// buildParentPositionPromo creates parent position for promotion (un-promote).
func buildParentPositionPromo(pos *GameState, from, to Square, promotedPieceIdx int, promo PromoPiece) PackedPosition {
	parent := *pos

	parent.SideToMove ^= 1

	// Remove promoted piece from 'to'
	parent.pieces[promotedPieceIdx] &^= 1 << uint(to)

	// Add pawn at 'from'
	pawnIdx := v2WPawn
	if promotedPieceIdx >= v2BPawn {
		pawnIdx = v2BPawn
	}
	parent.pieces[pawnIdx] |= 1 << uint(from)

	// Rebuild occupancy
	parent.occ[White] = parent.pieces[v2WPawn] | parent.pieces[v2WKnight] | parent.pieces[v2WBishop] |
		parent.pieces[v2WRook] | parent.pieces[v2WQueen] | parent.pieces[v2WKing]
	parent.occ[Black] = parent.pieces[v2BPawn] | parent.pieces[v2BKnight] | parent.pieces[v2BBishop] |
		parent.pieces[v2BRook] | parent.pieces[v2BQueen] | parent.pieces[v2BKing]
	parent.occAll = parent.occ[White] | parent.occ[Black]

	parent.EP = -1

	if parent.SideToMove == Black {
		parent.Fullmove--
		if parent.Fullmove < 1 {
			parent.Fullmove = 1
		}
	}

	return parent.Pack()
}

// buildParentPositionPromoCapture creates parent position for promotion with capture.
func buildParentPositionPromoCapture(pos *GameState, from, to Square, promotedPieceIdx int, promo PromoPiece, capturedIdx int) PackedPosition {
	parent := *pos

	parent.SideToMove ^= 1

	// Remove promoted piece from 'to'
	parent.pieces[promotedPieceIdx] &^= 1 << uint(to)

	// Add pawn at 'from'
	pawnIdx := v2WPawn
	if promotedPieceIdx >= v2BPawn {
		pawnIdx = v2BPawn
	}
	parent.pieces[pawnIdx] |= 1 << uint(from)

	// Restore captured piece at 'to'
	parent.pieces[capturedIdx] |= 1 << uint(to)

	// Rebuild occupancy
	parent.occ[White] = parent.pieces[v2WPawn] | parent.pieces[v2WKnight] | parent.pieces[v2WBishop] |
		parent.pieces[v2WRook] | parent.pieces[v2WQueen] | parent.pieces[v2WKing]
	parent.occ[Black] = parent.pieces[v2BPawn] | parent.pieces[v2BKnight] | parent.pieces[v2BBishop] |
		parent.pieces[v2BRook] | parent.pieces[v2BQueen] | parent.pieces[v2BKing]
	parent.occAll = parent.occ[White] | parent.occ[Black]

	parent.EP = -1

	// Restore castling rights if rook was captured
	parent.Castle = expandCastlingRightsCapture(pos.Castle, to, capturedIdx)

	if parent.SideToMove == Black {
		parent.Fullmove--
		if parent.Fullmove < 1 {
			parent.Fullmove = 1
		}
	}

	return parent.Pack()
}

// buildParentPositionCastle creates parent position for un-castling.
func buildParentPositionCastle(pos *GameState, kingFrom, kingTo Square, white, kingside bool) PackedPosition {
	parent := *pos

	parent.SideToMove ^= 1

	if white {
		// Move king from 'to' back to 'from' (e1)
		parent.pieces[v2WKing] &^= 1 << uint(kingTo)
		parent.pieces[v2WKing] |= 1 << uint(kingFrom)

		if kingside {
			// Rook from f1 back to h1
			parent.pieces[v2WRook] &^= 1 << 5
			parent.pieces[v2WRook] |= 1 << 7
		} else {
			// Rook from d1 back to a1
			parent.pieces[v2WRook] &^= 1 << 3
			parent.pieces[v2WRook] |= 1 << 0
		}
		// Restore BOTH white castling rights (king was on e1)
		// Also check if rooks are in their starting positions
		parent.Castle |= (1 << 0) // K - if h1 rook exists
		if parent.pieces[v2WRook]&(1<<0) != 0 {
			parent.Castle |= (1 << 1) // Q - if a1 rook exists
		}
	} else {
		parent.pieces[v2BKing] &^= 1 << uint(kingTo)
		parent.pieces[v2BKing] |= 1 << uint(kingFrom)

		if kingside {
			parent.pieces[v2BRook] &^= 1 << 61
			parent.pieces[v2BRook] |= 1 << 63
		} else {
			parent.pieces[v2BRook] &^= 1 << 59
			parent.pieces[v2BRook] |= 1 << 56
		}
		// Restore BOTH black castling rights (king was on e8)
		parent.Castle |= (1 << 2) // k - if h8 rook exists
		if parent.pieces[v2BRook]&(1<<56) != 0 {
			parent.Castle |= (1 << 3) // q - if a8 rook exists
		}
	}

	// Rebuild occupancy
	parent.occ[White] = parent.pieces[v2WPawn] | parent.pieces[v2WKnight] | parent.pieces[v2WBishop] |
		parent.pieces[v2WRook] | parent.pieces[v2WQueen] | parent.pieces[v2WKing]
	parent.occ[Black] = parent.pieces[v2BPawn] | parent.pieces[v2BKnight] | parent.pieces[v2BBishop] |
		parent.pieces[v2BRook] | parent.pieces[v2BQueen] | parent.pieces[v2BKing]
	parent.occAll = parent.occ[White] | parent.occ[Black]

	parent.EP = -1

	if parent.SideToMove == Black {
		parent.Fullmove--
		if parent.Fullmove < 1 {
			parent.Fullmove = 1
		}
	}

	return parent.Pack()
}

// expandCastlingRights potentially restores castling rights that were lost by the move.
func expandCastlingRights(castle uint8, from Square, pieceIdx int) uint8 {
	// If king moved from starting square, parent might have had castling rights
	switch pieceIdx {
	case v2WKing:
		if from == 4 {
			// We don't know if parent had rights, but we should try with rights
			// To reduce false positives, only add if rooks are in starting positions
			return castle | (1 << 0) | (1 << 1) // Add both K and Q
		}
	case v2BKing:
		if from == 60 {
			return castle | (1 << 2) | (1 << 3) // Add both k and q
		}
	case v2WRook:
		if from == 0 {
			return castle | (1 << 1) // Q
		}
		if from == 7 {
			return castle | (1 << 0) // K
		}
	case v2BRook:
		if from == 56 {
			return castle | (1 << 3) // q
		}
		if from == 63 {
			return castle | (1 << 2) // k
		}
	}
	return castle
}

// expandCastlingRightsCapture restores castling rights if rook was captured.
func expandCastlingRightsCapture(castle uint8, capturedSq Square, capturedIdx int) uint8 {
	if capturedIdx == v2WRook {
		if capturedSq == 0 {
			return castle | (1 << 1)
		}
		if capturedSq == 7 {
			return castle | (1 << 0)
		}
	}
	if capturedIdx == v2BRook {
		if capturedSq == 56 {
			return castle | (1 << 3)
		}
		if capturedSq == 63 {
			return castle | (1 << 2)
		}
	}
	return castle
}

// pieceIdxToPromo converts a piece index to promotion type.
func pieceIdxToPromo(idx int) PromoPiece {
	switch idx {
	case v2WQueen, v2BQueen:
		return PromoQueen
	case v2WRook, v2BRook:
		return PromoRook
	case v2WBishop, v2BBishop:
		return PromoBishop
	case v2WKnight, v2BKnight:
		return PromoKnight
	}
	return NoPromo
}

// ValidateParentMove checks if the forward move from parent to current is legal.
// Use this after looking up a parent position in your store.
func ValidateParentMove(parent *GameState, move Mv) bool {
	legalMoves := GenerateLegalMoves(parent)
	for _, m := range legalMoves {
		if m.From == move.From && m.To == move.To && m.Promo == move.Promo && m.Flags == move.Flags {
			return true
		}
	}
	return false
}

// ValidateParentMoveQuick is like ValidateParentMove but only checks from/to/promo.
// This is sufficient for most cases since flags are mainly internal.
func ValidateParentMoveQuick(parent *GameState, move Mv) bool {
	legalMoves := GenerateLegalMoves(parent)
	for _, m := range legalMoves {
		if m.From == move.From && m.To == move.To && m.Promo == move.Promo {
			return true
		}
	}
	return false
}
