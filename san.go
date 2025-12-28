// san.go provides SAN (Standard Algebraic Notation) parsing using magic bitboards.
//
// ParseSAN and ParseSANBytes convert SAN move strings (e.g., "e4", "Nxf3", "O-O")
// to Mv structs. The parser uses magic bitboards to directly find attackers
// without generating all legal moves, providing fast disambiguation.

package pgn

import (
	"fmt"
	"math/bits"
)

// ParseSAN parses a SAN move string and returns the matching move.
func ParseSAN(gs *GameState, san string) (Mv, error) {
	return ParseSANBytes(gs, []byte(san))
}

// ParseSANBytes parses a SAN move from a byte slice without allocations.
func ParseSANBytes(gs *GameState, san []byte) (Mv, error) {
	n := len(san)

	for n > 0 {
		c := san[n-1]
		if c == '+' || c == '#' || c == '!' || c == '?' {
			n--
		} else {
			break
		}
	}
	if n == 0 {
		return Mv{}, errInvalidSAN
	}

	if san[0] == 'O' || san[0] == '0' {
		if n >= 3 && (san[1] == '-' || san[1] == 'O' || san[1] == '0') {
			if n >= 5 && san[2] == 'O' || san[2] == '0' || san[2] == '-' {
				return parseCastleQueenside(gs)
			}
			return parseCastleKingside(gs)
		}
	}

	if n < 2 {
		return Mv{}, errInvalidSAN
	}

	var promo PromoPiece
	if n >= 4 && san[n-2] == '=' {
		promo = parsePromo(san[n-1])
		n -= 2
	}

	toFile := int(san[n-2] - 'a')
	toRank := int(san[n-1] - '1')
	if toFile < 0 || toFile > 7 || toRank < 0 || toRank > 7 {
		return Mv{}, errInvalidSAN
	}
	toSq := Square(toRank*8 + toFile)

	var pieceType byte = 'P'
	var disambFile, disambRank int = -1, -1

	i := 0
	if n > 2 && san[0] >= 'A' && san[0] <= 'Z' && san[0] != 'O' {
		pieceType = san[0]
		i = 1
	}

	for i < n-2 {
		c := san[i]
		if c == 'x' {
			i++
			continue
		}
		if c >= 'a' && c <= 'h' {
			disambFile = int(c - 'a')
		} else if c >= '1' && c <= '8' {
			disambRank = int(c - '1')
		}
		i++
	}

	from, err := findAttacker(gs, toSq, pieceType, disambFile, disambRank)
	if err != nil {
		return Mv{}, err
	}

	var flags uint16
	if pieceType == 'P' {
		if gs.EP == toSq {
			flags = 2
		}
		fromRank := int(from / 8)
		if gs.SideToMove == White && fromRank == 1 && toRank == 3 {
			flags = 1
		} else if gs.SideToMove == Black && fromRank == 6 && toRank == 4 {
			flags = 1
		}
	}

	mv := Mv{From: from, To: toSq, Promo: promo, Flags: flags}

	if !isLegalMove(gs, mv) {
		return Mv{}, errIllegalMove
	}

	return mv, nil
}

var errInvalidSAN = fmt.Errorf("invalid SAN")
var errIllegalMove = fmt.Errorf("illegal move")

func parsePromo(c byte) PromoPiece {
	switch c {
	case 'Q':
		return PromoQueen
	case 'R':
		return PromoRook
	case 'B':
		return PromoBishop
	case 'N':
		return PromoKnight
	}
	return NoPromo
}

func parseCastleKingside(gs *GameState) (Mv, error) {
	if gs.SideToMove == White {
		if gs.Castle&1 == 0 {
			return Mv{}, fmt.Errorf("kingside castle not allowed")
		}
		mv := Mv{From: 4, To: 6, Flags: 4}
		if isLegalMove(gs, mv) {
			return mv, nil
		}
	} else {
		if gs.Castle&4 == 0 {
			return Mv{}, fmt.Errorf("kingside castle not allowed")
		}
		mv := Mv{From: 60, To: 62, Flags: 4}
		if isLegalMove(gs, mv) {
			return mv, nil
		}
	}
	return Mv{}, fmt.Errorf("illegal castle O-O")
}

func parseCastleQueenside(gs *GameState) (Mv, error) {
	if gs.SideToMove == White {
		if gs.Castle&2 == 0 {
			return Mv{}, fmt.Errorf("queenside castle not allowed")
		}
		mv := Mv{From: 4, To: 2, Flags: 4}
		if isLegalMove(gs, mv) {
			return mv, nil
		}
	} else {
		if gs.Castle&8 == 0 {
			return Mv{}, fmt.Errorf("queenside castle not allowed")
		}
		mv := Mv{From: 60, To: 58, Flags: 4}
		if isLegalMove(gs, mv) {
			return mv, nil
		}
	}
	return Mv{}, fmt.Errorf("illegal castle O-O-O")
}

// findAttacker finds the piece that can move to toSq using magic bitboards.
func findAttacker(gs *GameState, toSq Square, pieceType byte, disambFile, disambRank int) (Square, error) {
	toMask := Bitboard(1) << uint(toSq)
	us := gs.SideToMove

	var attackers Bitboard

	switch pieceType {
	case 'P':
		attackers = findPawnAttackers(gs, toSq, us)
	case 'N':
		knightBB := gs.pieces[v2WKnight]
		if us == Black {
			knightBB = gs.pieces[v2BKnight]
		}
		attackers = knightAttacks(toSq) & knightBB
	case 'B':
		bishopBB := gs.pieces[v2WBishop]
		if us == Black {
			bishopBB = gs.pieces[v2BBishop]
		}
		attackers = bishopAttacks(toSq, gs.occAll) & bishopBB
	case 'R':
		rookBB := gs.pieces[v2WRook]
		if us == Black {
			rookBB = gs.pieces[v2BRook]
		}
		attackers = rookAttacks(toSq, gs.occAll) & rookBB
	case 'Q':
		queenBB := gs.pieces[v2WQueen]
		if us == Black {
			queenBB = gs.pieces[v2BQueen]
		}
		attackers = queenAttacks(toSq, gs.occAll) & queenBB
	case 'K':
		kingBB := gs.pieces[v2WKing]
		if us == Black {
			kingBB = gs.pieces[v2BKing]
		}
		attackers = kingAttacks(toSq) & kingBB
	}

	attackers &^= toMask

	if disambFile >= 0 {
		attackers &= fileMasks[disambFile]
	}
	if disambRank >= 0 {
		attackers &= rankMasks[disambRank]
	}

	if attackers == 0 {
		return -1, fmt.Errorf("no piece can move to target")
	}

	count := bits.OnesCount64(uint64(attackers))
	if count > 1 {
		var legalFrom Square = -1
		legalCount := 0
		for attackers != 0 {
			from := Square(bits.TrailingZeros64(uint64(attackers)))
			attackers &^= Bitboard(1) << uint(from)

			mv := Mv{From: from, To: toSq, Promo: NoPromo}
			if isLegalMoveQuick(gs, mv) {
				legalFrom = from
				legalCount++
			}
		}
		if legalCount == 0 {
			return -1, fmt.Errorf("no legal move to target")
		}
		if legalCount > 1 {
			return -1, fmt.Errorf("ambiguous move")
		}
		return legalFrom, nil
	}

	return Square(bits.TrailingZeros64(uint64(attackers))), nil
}

func isLegalMoveQuick(gs *GameState, mv Mv) bool {
	undo := MakeMove(gs, mv)
	kingSq := kingSquare(gs, gs.SideToMove^1)
	inCheck := squareAttacked(gs, kingSq, gs.SideToMove)
	UnmakeMove(gs, mv, undo)
	return !inCheck
}

func findPawnAttackers(gs *GameState, toSq Square, us Color) Bitboard {
	toMask := Bitboard(1) << uint(toSq)
	toRank := toSq / 8
	toFile := toSq % 8

	var attackers Bitboard

	if us == White {
		pawns := gs.pieces[v2WPawn]
		if gs.occAll&toMask == 0 {
			singleFrom := toSq - 8
			if singleFrom >= 0 && pawns&(Bitboard(1)<<uint(singleFrom)) != 0 {
				attackers |= Bitboard(1) << uint(singleFrom)
			}
			if toRank == 3 {
				doubleFrom := toSq - 16
				middleSq := toSq - 8
				if doubleFrom >= 0 && gs.occAll&(Bitboard(1)<<uint(middleSq)) == 0 {
					if pawns&(Bitboard(1)<<uint(doubleFrom)) != 0 {
						attackers |= Bitboard(1) << uint(doubleFrom)
					}
				}
			}
		}
		if gs.occ[Black]&toMask != 0 || gs.EP == toSq {
			if toFile > 0 {
				capFrom := toSq - 9
				if capFrom >= 0 && pawns&(Bitboard(1)<<uint(capFrom)) != 0 {
					attackers |= Bitboard(1) << uint(capFrom)
				}
			}
			if toFile < 7 {
				capFrom := toSq - 7
				if capFrom >= 0 && pawns&(Bitboard(1)<<uint(capFrom)) != 0 {
					attackers |= Bitboard(1) << uint(capFrom)
				}
			}
		}
	} else {
		pawns := gs.pieces[v2BPawn]
		if gs.occAll&toMask == 0 {
			singleFrom := toSq + 8
			if singleFrom < 64 && pawns&(Bitboard(1)<<uint(singleFrom)) != 0 {
				attackers |= Bitboard(1) << uint(singleFrom)
			}
			if toRank == 4 {
				doubleFrom := toSq + 16
				middleSq := toSq + 8
				if doubleFrom < 64 && gs.occAll&(Bitboard(1)<<uint(middleSq)) == 0 {
					if pawns&(Bitboard(1)<<uint(doubleFrom)) != 0 {
						attackers |= Bitboard(1) << uint(doubleFrom)
					}
				}
			}
		}
		if gs.occ[White]&toMask != 0 || gs.EP == toSq {
			if toFile > 0 {
				capFrom := toSq + 7
				if capFrom < 64 && pawns&(Bitboard(1)<<uint(capFrom)) != 0 {
					attackers |= Bitboard(1) << uint(capFrom)
				}
			}
			if toFile < 7 {
				capFrom := toSq + 9
				if capFrom < 64 && pawns&(Bitboard(1)<<uint(capFrom)) != 0 {
					attackers |= Bitboard(1) << uint(capFrom)
				}
			}
		}
	}

	return attackers
}

func isLegalMove(gs *GameState, mv Mv) bool {
	undo := MakeMove(gs, mv)
	kingSq := kingSquare(gs, gs.SideToMove^1)
	inCheck := squareAttacked(gs, kingSq, gs.SideToMove)
	UnmakeMove(gs, mv, undo)
	return !inCheck
}
