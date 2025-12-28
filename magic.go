// magic.go implements magic bitboards for efficient slider (bishop, rook, queen) attacks.
//
// Magic bitboards use a precomputed lookup table indexed by a "magic" multiplication
// and shift of the occupied squares. This provides O(1) attack generation for sliders.
// The magic numbers are well-known constants from the Chess Programming Wiki.

package pgn

import "math/bits"

var (
	bishopMagics [64]uint64
	rookMagics   [64]uint64
	bishopMasks  [64]Bitboard
	rookMasks    [64]Bitboard
	bishopShifts [64]int
	rookShifts   [64]int
	bishopTable  [64][]Bitboard
	rookTable    [64][]Bitboard
)

var bishopMagicNumbers = [64]uint64{
	0x0002020202020200, 0x0002020202020000, 0x0004010202000000, 0x0004040080000000,
	0x0001104000000000, 0x0000821040000000, 0x0000410410400000, 0x0000104104104000,
	0x0000040404040400, 0x0000020202020200, 0x0000040102020000, 0x0000040400800000,
	0x0000011040000000, 0x0000008210400000, 0x0000004104104000, 0x0000002082082000,
	0x0004000808080800, 0x0002000404040400, 0x0001000202020200, 0x0000800802004000,
	0x0000800400A00000, 0x0000200100884000, 0x0000400082082000, 0x0000200041041000,
	0x0002080010101000, 0x0001040008080800, 0x0000208004010400, 0x0000404004010200,
	0x0000840000802000, 0x0000404002011000, 0x0000808001041000, 0x0000404000820800,
	0x0001041000202000, 0x0000820800101000, 0x0000104400080800, 0x0000020080080080,
	0x0000404040040100, 0x0000808100020100, 0x0001010100020800, 0x0000808080010400,
	0x0000820820004000, 0x0000410410002000, 0x0000082088001000, 0x0000002011000800,
	0x0000080100400400, 0x0001010101000200, 0x0002020202000400, 0x0001010101000200,
	0x0000410410400000, 0x0000208208200000, 0x0000002084100000, 0x0000000020880000,
	0x0000001002020000, 0x0000040408020000, 0x0004040404040000, 0x0002020202020000,
	0x0000104104104000, 0x0000002082082000, 0x0000000020841000, 0x0000000000208800,
	0x0000000010020200, 0x0000000404080200, 0x0000040404040400, 0x0002020202020200,
}

var rookMagicNumbers = [64]uint64{
	0x0080001020400080, 0x0040001000200040, 0x0080081000200080, 0x0080040800100080,
	0x0080020400080080, 0x0080010200040080, 0x0080008001000200, 0x0080002040800100,
	0x0000800020400080, 0x0000400020005000, 0x0000801000200080, 0x0000800800100080,
	0x0000800400080080, 0x0000800200040080, 0x0000800100020080, 0x0000800040800100,
	0x0000208000400080, 0x0000404000201000, 0x0000808010002000, 0x0000808008001000,
	0x0000808004000800, 0x0000808002000400, 0x0000010100020004, 0x0000020000408104,
	0x0000208080004000, 0x0000200040005000, 0x0000100080200080, 0x0000080080100080,
	0x0000040080080080, 0x0000020080040080, 0x0000010080800200, 0x0000800080004100,
	0x0000204000800080, 0x0000200040401000, 0x0000100080802000, 0x0000080080801000,
	0x0000040080800800, 0x0000020080800400, 0x0000020001010004, 0x0000800040800100,
	0x0000204000808000, 0x0000200040008080, 0x0000100020008080, 0x0000080010008080,
	0x0000040008008080, 0x0000020004008080, 0x0000010002008080, 0x0000004081020004,
	0x0000204000800080, 0x0000200040008080, 0x0000100020008080, 0x0000080010008080,
	0x0000040008008080, 0x0000020004008080, 0x0000800100020080, 0x0000800041000080,
	0x00FFFCDDFCED714A, 0x007FFCDDFCED714A, 0x003FFFCDFFD88096, 0x0000040810002101,
	0x0001000204080011, 0x0001000204000801, 0x0001000082000401, 0x0001FFFAABFAD1A2,
}

func init() {
	initMagicBishop()
	initMagicRook()
}

func initMagicBishop() {
	for sq := 0; sq < 64; sq++ {
		bishopMasks[sq] = computeBishopMask(Square(sq))
		bishopMagics[sq] = bishopMagicNumbers[sq]
		bits := bits.OnesCount64(uint64(bishopMasks[sq]))
		bishopShifts[sq] = 64 - bits
		tableSize := 1 << bits
		bishopTable[sq] = make([]Bitboard, tableSize)

		occ := Bitboard(0)
		for {
			idx := (uint64(occ) * bishopMagics[sq]) >> uint(bishopShifts[sq])
			bishopTable[sq][idx] = computeBishopAttacks(Square(sq), occ)
			occ = (occ - bishopMasks[sq]) & bishopMasks[sq]
			if occ == 0 {
				break
			}
		}
	}
}

func initMagicRook() {
	for sq := 0; sq < 64; sq++ {
		rookMasks[sq] = computeRookMask(Square(sq))
		rookMagics[sq] = rookMagicNumbers[sq]
		bits := bits.OnesCount64(uint64(rookMasks[sq]))
		rookShifts[sq] = 64 - bits
		tableSize := 1 << bits
		rookTable[sq] = make([]Bitboard, tableSize)

		occ := Bitboard(0)
		for {
			idx := (uint64(occ) * rookMagics[sq]) >> uint(rookShifts[sq])
			rookTable[sq][idx] = computeRookAttacks(Square(sq), occ)
			occ = (occ - rookMasks[sq]) & rookMasks[sq]
			if occ == 0 {
				break
			}
		}
	}
}

func computeBishopMask(sq Square) Bitboard {
	var mask Bitboard
	r, f := int(sq/8), int(sq%8)
	for _, d := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		dr, df := d[0], d[1]
		rr, ff := r+dr, f+df
		for rr > 0 && rr < 7 && ff > 0 && ff < 7 {
			mask |= 1 << uint(rr*8+ff)
			rr += dr
			ff += df
		}
	}
	return mask
}

func computeRookMask(sq Square) Bitboard {
	var mask Bitboard
	r, f := int(sq/8), int(sq%8)
	for ff := f + 1; ff < 7; ff++ {
		mask |= 1 << uint(r*8+ff)
	}
	for ff := f - 1; ff > 0; ff-- {
		mask |= 1 << uint(r*8+ff)
	}
	for rr := r + 1; rr < 7; rr++ {
		mask |= 1 << uint(rr*8+f)
	}
	for rr := r - 1; rr > 0; rr-- {
		mask |= 1 << uint(rr*8+f)
	}
	return mask
}

func computeBishopAttacks(sq Square, blockers Bitboard) Bitboard {
	var attacks Bitboard
	r, f := int(sq/8), int(sq%8)
	for _, d := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		rr, ff := r+d[0], f+d[1]
		for rr >= 0 && rr <= 7 && ff >= 0 && ff <= 7 {
			bit := Bitboard(1) << uint(rr*8+ff)
			attacks |= bit
			if blockers&bit != 0 {
				break
			}
			rr += d[0]
			ff += d[1]
		}
	}
	return attacks
}

func computeRookAttacks(sq Square, blockers Bitboard) Bitboard {
	var attacks Bitboard
	r, f := int(sq/8), int(sq%8)
	for ff := f + 1; ff <= 7; ff++ {
		bit := Bitboard(1) << uint(r*8+ff)
		attacks |= bit
		if blockers&bit != 0 {
			break
		}
	}
	for ff := f - 1; ff >= 0; ff-- {
		bit := Bitboard(1) << uint(r*8+ff)
		attacks |= bit
		if blockers&bit != 0 {
			break
		}
	}
	for rr := r + 1; rr <= 7; rr++ {
		bit := Bitboard(1) << uint(rr*8+f)
		attacks |= bit
		if blockers&bit != 0 {
			break
		}
	}
	for rr := r - 1; rr >= 0; rr-- {
		bit := Bitboard(1) << uint(rr*8+f)
		attacks |= bit
		if blockers&bit != 0 {
			break
		}
	}
	return attacks
}

// bishopAttacks returns bishop attacks from sq given board occupancy.
func bishopAttacks(sq Square, occ Bitboard) Bitboard {
	occ &= bishopMasks[sq]
	idx := (uint64(occ) * bishopMagics[sq]) >> uint(bishopShifts[sq])
	return bishopTable[sq][idx]
}

// rookAttacks returns rook attacks from sq given board occupancy.
func rookAttacks(sq Square, occ Bitboard) Bitboard {
	occ &= rookMasks[sq]
	idx := (uint64(occ) * rookMagics[sq]) >> uint(rookShifts[sq])
	return rookTable[sq][idx]
}

// queenAttacks returns queen attacks (bishop + rook combined).
func queenAttacks(sq Square, occ Bitboard) Bitboard {
	return bishopAttacks(sq, occ) | rookAttacks(sq, occ)
}
