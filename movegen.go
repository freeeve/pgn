// movegen.go provides legal move generation using bitboards and magic bitboards.
//
// GenerateLegalMoves returns all legal moves from a position by generating
// pseudo-legal moves and filtering out those that leave the king in check.
// GenerateLegalMovesPooled provides a pooled version for reduced allocations.

package pgn

import (
	"math/bits"
	"sync"
)

// maxMoves is the maximum possible legal moves in any chess position.
const maxMoves = 256

var moveListPool = sync.Pool{
	New: func() interface{} {
		s := make([]Mv, 0, maxMoves)
		return &s
	},
}

func getMoveList() *[]Mv {
	return moveListPool.Get().(*[]Mv)
}

func releaseMoves(moves *[]Mv) {
	*moves = (*moves)[:0]
	moveListPool.Put(moves)
}

// generateMovesPooled is the internal pooled version for perft.
func generateMovesPooled(pos *GameState) *[]Mv {
	if pos == nil {
		return nil
	}

	pseudo := getMoveList()
	genPseudoMovesTo(pseudo, pos)

	legal := getMoveList()
	for _, mv := range *pseudo {
		undo := MakeMove(pos, mv)
		kSq := kingSquare(pos, pos.SideToMove^1)
		inCheck := squareAttacked(pos, kSq, pos.SideToMove)
		UnmakeMove(pos, mv, undo)
		if !inCheck {
			*legal = append(*legal, mv)
		}
	}
	releaseMoves(pseudo)

	return legal
}

// GenerateLegalMoves returns all legal moves from the position.
func GenerateLegalMoves(pos *GameState) []Mv {
	if pos == nil {
		return nil
	}

	pseudo := generatePseudoMoves(pos)
	legal := make([]Mv, 0, len(pseudo))
	for _, mv := range pseudo {
		undo := MakeMove(pos, mv)
		kSq := kingSquare(pos, pos.SideToMove^1)
		inCheck := squareAttacked(pos, kSq, pos.SideToMove)
		UnmakeMove(pos, mv, undo)
		if !inCheck {
			legal = append(legal, mv)
		}
	}

	return legal
}

// WithLegalMoves generates legal moves using pooled slices and calls fn with them.
// The moves slice is automatically released after fn returns.
// This is more efficient than GenerateLegalMoves for repeated calls.
func WithLegalMoves(pos *GameState, fn func([]Mv)) {
	if pos == nil {
		fn(nil)
		return
	}

	pseudo := getMoveList()
	genPseudoMovesTo(pseudo, pos)

	legal := getMoveList()
	for _, mv := range *pseudo {
		undo := MakeMove(pos, mv)
		kSq := kingSquare(pos, pos.SideToMove^1)
		inCheck := squareAttacked(pos, kSq, pos.SideToMove)
		UnmakeMove(pos, mv, undo)
		if !inCheck {
			*legal = append(*legal, mv)
		}
	}
	releaseMoves(pseudo)

	fn(*legal)
	releaseMoves(legal)
}

// generatePseudoMoves returns pseudo-legal moves without king-safety filtering.
func generatePseudoMoves(pos *GameState) []Mv {
	if pos == nil {
		return nil
	}
	pseudo := make([]Mv, 0, 64)
	genPseudoMovesTo(&pseudo, pos)
	return pseudo
}

func genPseudoMovesTo(pseudo *[]Mv, pos *GameState) {
	enemyKing := pos.pieces[v2BKing]
	if pos.SideToMove == White {
		genWhitePawnMovesTo(pseudo, pos, enemyKing)
		genKnightMovesTo(pseudo, pos, v2WKnight, enemyKing)
		genBishopMovesTo(pseudo, pos, v2WBishop, enemyKing)
		genRookMovesTo(pseudo, pos, v2WRook, enemyKing)
		genQueenMovesTo(pseudo, pos, v2WQueen, enemyKing)
		genKingMovesTo(pseudo, pos, v2WKing, enemyKing)
		genCastlesTo(pseudo, pos, true)
	} else {
		enemyKing = pos.pieces[v2WKing]
		genBlackPawnMovesTo(pseudo, pos, enemyKing)
		genKnightMovesTo(pseudo, pos, v2BKnight, enemyKing)
		genBishopMovesTo(pseudo, pos, v2BBishop, enemyKing)
		genRookMovesTo(pseudo, pos, v2BRook, enemyKing)
		genQueenMovesTo(pseudo, pos, v2BQueen, enemyKing)
		genKingMovesTo(pseudo, pos, v2BKing, enemyKing)
		genCastlesTo(pseudo, pos, false)
	}
}

func genWhitePawnMovesTo(moves *[]Mv, pos *GameState, enemyKing Bitboard) {
	pawns := pos.pieces[v2WPawn]
	empty := ^pos.occAll

	single := (pawns << 8) & empty
	addPawnPushes(moves, single, false, false)

	rank2 := Bitboard(0x000000000000FF00)
	double := ((pawns & rank2) << 8) & empty
	double = (double << 8) & empty
	addPawnPushes(moves, double, true, false)

	leftCap := ((pawns & ^fileMask(0)) << 7) & pos.occ[Black] & ^enemyKing
	rightCap := ((pawns & ^fileMask(7)) << 9) & pos.occ[Black] & ^enemyKing
	addPawnCaptures(moves, leftCap, false)
	addPawnCaptures(moves, rightCap, true)

	if pos.EP >= 0 && pos.EP < 64 {
		epMask := Bitboard(1) << uint(pos.EP)
		leftEP := ((pawns & ^fileMask(0)) << 7) & epMask
		rightEP := ((pawns & ^fileMask(7)) << 9) & epMask
		if leftEP != 0 {
			to := Square(pos.EP)
			from := to - 7
			*moves = append(*moves, Mv{From: from, To: to, Flags: 2})
		}
		if rightEP != 0 {
			to := Square(pos.EP)
			from := to - 9
			*moves = append(*moves, Mv{From: from, To: to, Flags: 2})
		}
	}
}

func genBlackPawnMovesTo(moves *[]Mv, pos *GameState, enemyKing Bitboard) {
	pawns := pos.pieces[v2BPawn]
	empty := ^pos.occAll

	single := (pawns >> 8) & empty
	addPawnPushesDown(moves, single, false, false)

	rank7 := Bitboard(0x00FF000000000000)
	double := ((pawns & rank7) >> 8) & empty
	double = (double >> 8) & empty
	addPawnPushesDown(moves, double, true, false)

	leftCap := ((pawns & ^fileMask(0)) >> 9) & pos.occ[White] & ^enemyKing
	rightCap := ((pawns & ^fileMask(7)) >> 7) & pos.occ[White] & ^enemyKing
	addPawnCapturesDown(moves, leftCap, false)
	addPawnCapturesDown(moves, rightCap, true)

	if pos.EP >= 0 && pos.EP < 64 {
		epMask := Bitboard(1) << uint(pos.EP)
		leftEP := ((pawns & ^fileMask(0)) >> 9) & epMask
		rightEP := ((pawns & ^fileMask(7)) >> 7) & epMask
		if leftEP != 0 {
			to := Square(pos.EP)
			from := to + 9
			*moves = append(*moves, Mv{From: from, To: to, Flags: 2})
		}
		if rightEP != 0 {
			to := Square(pos.EP)
			from := to + 7
			*moves = append(*moves, Mv{From: from, To: to, Flags: 2})
		}
	}
}

func addPawnPushes(moves *[]Mv, bb Bitboard, isDouble bool, _ bool) {
	for bb != 0 {
		to := Square(bitscanForward(bb))
		from := to - 8
		if isDouble {
			from = to - 16
		}
		flag := uint16(0)
		if isDouble {
			flag = 1
		}
		if to >= 56 {
			for _, promo := range []PromoPiece{PromoQueen, PromoRook, PromoBishop, PromoKnight} {
				*moves = append(*moves, Mv{From: from, To: to, Flags: flag, Promo: promo})
			}
		} else {
			*moves = append(*moves, Mv{From: from, To: to, Flags: flag})
		}
		bb &= bb - 1
	}
}

func addPawnPushesDown(moves *[]Mv, bb Bitboard, isDouble bool, _ bool) {
	for bb != 0 {
		to := Square(bitscanForward(bb))
		from := to + 8
		if isDouble {
			from = to + 16
		}
		flag := uint16(0)
		if isDouble {
			flag = 1
		}
		if to <= 7 {
			for _, promo := range []PromoPiece{PromoQueen, PromoRook, PromoBishop, PromoKnight} {
				*moves = append(*moves, Mv{From: from, To: to, Flags: flag, Promo: promo})
			}
		} else {
			*moves = append(*moves, Mv{From: from, To: to, Flags: flag})
		}
		bb &= bb - 1
	}
}

func addPawnCaptures(moves *[]Mv, bb Bitboard, right bool) {
	for bb != 0 {
		to := Square(bitscanForward(bb))
		var from Square
		if right {
			from = to - 9
		} else {
			from = to - 7
		}
		if to >= 56 {
			for _, promo := range []PromoPiece{PromoQueen, PromoRook, PromoBishop, PromoKnight} {
				*moves = append(*moves, Mv{From: from, To: to, Promo: promo})
			}
		} else {
			*moves = append(*moves, Mv{From: from, To: to})
		}
		bb &= bb - 1
	}
}

func addPawnCapturesDown(moves *[]Mv, bb Bitboard, right bool) {
	for bb != 0 {
		to := Square(bitscanForward(bb))
		var from Square
		if right {
			from = to + 7
		} else {
			from = to + 9
		}
		if to <= 7 {
			for _, promo := range []PromoPiece{PromoQueen, PromoRook, PromoBishop, PromoKnight} {
				*moves = append(*moves, Mv{From: from, To: to, Promo: promo})
			}
		} else {
			*moves = append(*moves, Mv{From: from, To: to})
		}
		bb &= bb - 1
	}
}

func genBishopMovesTo(moves *[]Mv, pos *GameState, idx int, enemyKing Bitboard) {
	bishops := pos.pieces[idx]
	us := White
	if idx >= v2BPawn {
		us = Black
	}
	for bishops != 0 {
		from := Square(bitscanForward(bishops))
		att := bishopAttacks(from, pos.occAll) & ^pos.occ[us] & ^enemyKing
		for att != 0 {
			to := Square(bitscanForward(att))
			*moves = append(*moves, Mv{From: from, To: to})
			att &= att - 1
		}
		bishops &= bishops - 1
	}
}

func genRookMovesTo(moves *[]Mv, pos *GameState, idx int, enemyKing Bitboard) {
	rooks := pos.pieces[idx]
	us := White
	if idx >= v2BPawn {
		us = Black
	}
	for rooks != 0 {
		from := Square(bitscanForward(rooks))
		att := rookAttacks(from, pos.occAll) & ^pos.occ[us] & ^enemyKing
		for att != 0 {
			to := Square(bitscanForward(att))
			*moves = append(*moves, Mv{From: from, To: to})
			att &= att - 1
		}
		rooks &= rooks - 1
	}
}

func genQueenMovesTo(moves *[]Mv, pos *GameState, idx int, enemyKing Bitboard) {
	queens := pos.pieces[idx]
	us := White
	if idx >= v2BPawn {
		us = Black
	}
	for queens != 0 {
		from := Square(bitscanForward(queens))
		att := queenAttacks(from, pos.occAll) & ^pos.occ[us] & ^enemyKing
		for att != 0 {
			to := Square(bitscanForward(att))
			*moves = append(*moves, Mv{From: from, To: to})
			att &= att - 1
		}
		queens &= queens - 1
	}
}

func genKnightMovesTo(moves *[]Mv, pos *GameState, idx int, enemyKing Bitboard) {
	knights := pos.pieces[idx]
	us := White
	if idx >= v2BPawn {
		us = Black
	}
	ourOcc := pos.occ[us]
	for knights != 0 {
		from := Square(bitscanForward(knights))
		att := knightAttacks(from) & ^ourOcc & ^enemyKing
		for att != 0 {
			to := Square(bitscanForward(att))
			*moves = append(*moves, Mv{From: from, To: to})
			att &= att - 1
		}
		knights &= knights - 1
	}
}

func genKingMovesTo(moves *[]Mv, pos *GameState, idx int, enemyKing Bitboard) {
	kings := pos.pieces[idx]
	if kings == 0 {
		return
	}
	us := White
	if idx >= v2BPawn {
		us = Black
	}
	ourOcc := pos.occ[us]
	from := Square(bitscanForward(kings))
	att := kingAttacks(from) & ^ourOcc & ^enemyKing
	for att != 0 {
		to := Square(bitscanForward(att))
		*moves = append(*moves, Mv{From: from, To: to})
		att &= att - 1
	}
}

func genCastlesTo(moves *[]Mv, pos *GameState, white bool) {
	if white {
		if pos.Castle&(1<<0) != 0 &&
			pos.pieces[v2WKing]&(1<<4) != 0 &&
			pos.pieces[v2WRook]&(1<<7) != 0 &&
			(pos.occAll&(1<<5|1<<6)) == 0 &&
			!squareAttacked(pos, 4, Black) &&
			!squareAttacked(pos, 5, Black) &&
			!squareAttacked(pos, 6, Black) {
			*moves = append(*moves, Mv{From: 4, To: 6, Flags: 4})
		}
		if pos.Castle&(1<<1) != 0 &&
			pos.pieces[v2WKing]&(1<<4) != 0 &&
			pos.pieces[v2WRook]&(1<<0) != 0 &&
			(pos.occAll&(1<<1|1<<2|1<<3)) == 0 &&
			!squareAttacked(pos, 4, Black) &&
			!squareAttacked(pos, 3, Black) &&
			!squareAttacked(pos, 2, Black) {
			*moves = append(*moves, Mv{From: 4, To: 2, Flags: 4})
		}
	} else {
		if pos.Castle&(1<<2) != 0 &&
			pos.pieces[v2BKing]&(1<<60) != 0 &&
			pos.pieces[v2BRook]&(1<<63) != 0 &&
			(pos.occAll&(1<<61|1<<62)) == 0 &&
			!squareAttacked(pos, 60, White) &&
			!squareAttacked(pos, 61, White) &&
			!squareAttacked(pos, 62, White) {
			*moves = append(*moves, Mv{From: 60, To: 62, Flags: 4})
		}
		if pos.Castle&(1<<3) != 0 &&
			pos.pieces[v2BKing]&(1<<60) != 0 &&
			pos.pieces[v2BRook]&(1<<56) != 0 &&
			(pos.occAll&(1<<57|1<<58|1<<59)) == 0 &&
			!squareAttacked(pos, 60, White) &&
			!squareAttacked(pos, 59, White) &&
			!squareAttacked(pos, 58, White) {
			*moves = append(*moves, Mv{From: 60, To: 58, Flags: 4})
		}
	}
}

// squareAttacked returns true if sq is attacked by the given color.
func squareAttacked(pos *GameState, sq Square, by Color) bool {
	if sq < 0 || sq > 63 {
		return false
	}
	if by == White {
		att := Bitboard(0)
		if sq-7 >= 0 {
			att |= (Bitboard(1) << uint(sq-7)) & ^fileMask(0)
		}
		if sq-9 >= 0 {
			att |= (Bitboard(1) << uint(sq-9)) & ^fileMask(7)
		}
		if pos.pieces[v2WPawn]&att != 0 {
			return true
		}
	} else {
		att := Bitboard(0)
		if sq+7 < 64 {
			att |= (Bitboard(1) << uint(sq+7)) & ^fileMask(7)
		}
		if sq+9 < 64 {
			att |= (Bitboard(1) << uint(sq+9)) & ^fileMask(0)
		}
		if pos.pieces[v2BPawn]&att != 0 {
			return true
		}
	}

	if by == White {
		if knightAttacks(sq)&pos.pieces[v2WKnight] != 0 {
			return true
		}
	} else {
		if knightAttacks(sq)&pos.pieces[v2BKnight] != 0 {
			return true
		}
	}

	if by == White {
		if kingAttacks(sq)&pos.pieces[v2WKing] != 0 {
			return true
		}
	} else {
		if kingAttacks(sq)&pos.pieces[v2BKing] != 0 {
			return true
		}
	}

	diagAtt := bishopAttacks(sq, pos.occAll)
	if by == White {
		if diagAtt&(pos.pieces[v2WBishop]|pos.pieces[v2WQueen]) != 0 {
			return true
		}
	} else {
		if diagAtt&(pos.pieces[v2BBishop]|pos.pieces[v2BQueen]) != 0 {
			return true
		}
	}

	orthoAtt := rookAttacks(sq, pos.occAll)
	if by == White {
		if orthoAtt&(pos.pieces[v2WRook]|pos.pieces[v2WQueen]) != 0 {
			return true
		}
	} else {
		if orthoAtt&(pos.pieces[v2BRook]|pos.pieces[v2BQueen]) != 0 {
			return true
		}
	}
	return false
}

func kingSquare(pos *GameState, color Color) Square {
	if color == White {
		if pos.pieces[v2WKing] != 0 {
			return Square(bitscanForward(pos.pieces[v2WKing]))
		}
	} else {
		if pos.pieces[v2BKing] != 0 {
			return Square(bitscanForward(pos.pieces[v2BKing]))
		}
	}
	return -1
}

func bitscanForward(bb Bitboard) int {
	return bits.TrailingZeros64(uint64(bb))
}
