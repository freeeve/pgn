// packed_position.go provides a compact 26-byte position encoding for URLs and databases.
//
// PackedPosition encodes a chess position into 26 bytes:
//   - 8 bytes: Occupancy bitmap (bit i set means square i has a piece)
//   - 16 bytes: Piece codes (4 bits each, up to 32 pieces, in occupancy bit order)
//   - 1 byte: Flags (side to move, castling rights)
//   - 1 byte: En-passant file (0-7, or 0xFF for none)
//
// This is 8 bytes smaller than the naive 34-byte encoding (4 bits per square).
// The base64 encoding produces a 35-character URL-safe string.

package pgn

import (
	"encoding/base64"
	"fmt"
	"math/bits"
)

// PackedPosition is a compact 26-byte representation of a chess position.
// Layout:
//   - Bytes 0-7: Occupancy bitmap (bit i set means square i has a piece)
//   - Bytes 8-23: Piece codes (4 bits each, up to 32 pieces, in occupancy bit order)
//   - Byte 24: Flags (side to move, castling rights)
//   - Byte 25: EP file (0-7, or 0xFF for none)
type PackedPosition [26]byte

// PackedFEN is a compact 29-byte representation containing all FEN fields.
// Layout:
//   - Bytes 0-25: PackedPosition (board + metadata)
//   - Byte 26: Halfmove clock (0-255)
//   - Bytes 27-28: Fullmove number (little-endian, 1-65535)
type PackedFEN [29]byte

// Piece codes for packed representation (0-11)
const (
	pcWP = iota // White Pawn
	pcWN        // White Knight
	pcWB        // White Bishop
	pcWR        // White Rook
	pcWQ        // White Queen
	pcWK        // White King
	pcBP        // Black Pawn
	pcBN        // Black Knight
	pcBB        // Black Bishop
	pcBR        // Black Rook
	pcBQ        // Black Queen
	pcBK        // Black King
)

// Flag bits for byte 24
const (
	flagBlackToMove = 1 << 0
	flagWKCastle    = 1 << 1
	flagWQCastle    = 1 << 2
	flagBKCastle    = 1 << 3
	flagBQCastle    = 1 << 4
	noEPFile        = 0xFF
)

// pieceToCode converts a piece character to its packed code.
func pieceToCode(p byte) byte {
	switch p {
	case 'P':
		return pcWP
	case 'N':
		return pcWN
	case 'B':
		return pcWB
	case 'R':
		return pcWR
	case 'Q':
		return pcWQ
	case 'K':
		return pcWK
	case 'p':
		return pcBP
	case 'n':
		return pcBN
	case 'b':
		return pcBB
	case 'r':
		return pcBR
	case 'q':
		return pcBQ
	case 'k':
		return pcBK
	default:
		return 0
	}
}

// codeToPiece converts a packed code to piece character.
func codeToPiece(code byte) byte {
	pieces := [12]byte{'P', 'N', 'B', 'R', 'Q', 'K', 'p', 'n', 'b', 'r', 'q', 'k'}
	if code < 12 {
		return pieces[code]
	}
	return 0
}

// codeToV2Index converts a packed code to v2 piece index.
func codeToV2Index(code byte) int {
	indices := [12]int{v2WPawn, v2WKnight, v2WBishop, v2WRook, v2WQueen, v2WKing,
		v2BPawn, v2BKnight, v2BBishop, v2BRook, v2BQueen, v2BKing}
	if code < 12 {
		return indices[code]
	}
	return -1
}

// -----------------------------------------------------------------------------
// PackedPosition methods
// -----------------------------------------------------------------------------

// Pack encodes a GameState into a PackedPosition (26 bytes with metadata).
func (gs *GameState) Pack() PackedPosition {
	var pp PackedPosition

	// Build occupancy bitmap and collect pieces in square order
	var occupancy uint64
	pieceIdx := 0

	for sq := 0; sq < 64; sq++ {
		p := gs.PieceAt(Square(sq))
		if p != 0 {
			occupancy |= 1 << sq

			// Store piece code (4 bits each)
			code := pieceToCode(p)
			byteIdx := 8 + pieceIdx/2
			if pieceIdx%2 == 0 {
				pp[byteIdx] = code & 0x0F
			} else {
				pp[byteIdx] |= (code & 0x0F) << 4
			}
			pieceIdx++
		}
	}

	// Store occupancy (little-endian)
	for i := 0; i < 8; i++ {
		pp[i] = byte(occupancy >> (i * 8))
	}

	// Store flags (byte 24)
	var flags byte
	if gs.SideToMove == Black {
		flags |= flagBlackToMove
	}
	if gs.Castle&(1<<0) != 0 {
		flags |= flagWKCastle
	}
	if gs.Castle&(1<<1) != 0 {
		flags |= flagWQCastle
	}
	if gs.Castle&(1<<2) != 0 {
		flags |= flagBKCastle
	}
	if gs.Castle&(1<<3) != 0 {
		flags |= flagBQCastle
	}
	pp[24] = flags

	// Store EP file (byte 25)
	if gs.EP >= 0 && gs.EP < 64 {
		pp[25] = byte(gs.EP % 8)
	} else {
		pp[25] = noEPFile
	}

	return pp
}

// Unpack decodes a PackedPosition into a GameState.
func (pp PackedPosition) Unpack() *GameState {
	gs := &GameState{EP: -1, Fullmove: 1}

	// Read occupancy (little-endian)
	var occupancy uint64
	for i := 0; i < 8; i++ {
		occupancy |= uint64(pp[i]) << (i * 8)
	}

	// Read pieces in occupancy order
	pieceIdx := 0
	for sq := 0; sq < 64; sq++ {
		if occupancy&(1<<sq) == 0 {
			continue
		}

		// Get piece code
		byteIdx := 8 + pieceIdx/2
		var code byte
		if pieceIdx%2 == 0 {
			code = pp[byteIdx] & 0x0F
		} else {
			code = (pp[byteIdx] >> 4) & 0x0F
		}
		pieceIdx++

		// Place piece
		idx := codeToV2Index(code)
		if idx >= 0 {
			gs.setPiece(idx, Square(sq))
		}
	}

	// Read flags (byte 24)
	flags := pp[24]
	if flags&flagBlackToMove != 0 {
		gs.SideToMove = Black
	} else {
		gs.SideToMove = White
	}
	if flags&flagWKCastle != 0 {
		gs.Castle |= 1 << 0
	}
	if flags&flagWQCastle != 0 {
		gs.Castle |= 1 << 1
	}
	if flags&flagBKCastle != 0 {
		gs.Castle |= 1 << 2
	}
	if flags&flagBQCastle != 0 {
		gs.Castle |= 1 << 3
	}

	// Read EP file (byte 25)
	epFile := pp[25]
	if epFile != noEPFile && epFile < 8 {
		if gs.SideToMove == White {
			gs.EP = Square(40 + int(epFile)) // rank 6
		} else {
			gs.EP = Square(16 + int(epFile)) // rank 3
		}
	}

	return gs
}

// String returns the base64 URL-safe encoding of the packed position.
func (pp PackedPosition) String() string {
	return base64.RawURLEncoding.EncodeToString(pp[:])
}

// ToFEN converts the packed position to a FEN string (with default metadata).
func (pp PackedPosition) ToFEN() string {
	return pp.Unpack().ToFEN()
}

// ParsePackedPosition decodes a base64 URL-safe encoded packed position.
func ParsePackedPosition(s string) (PackedPosition, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return PackedPosition{}, fmt.Errorf("invalid base64: %w", err)
	}
	if len(data) != 26 {
		return PackedPosition{}, fmt.Errorf("packed position must be 26 bytes, got %d", len(data))
	}
	var pp PackedPosition
	copy(pp[:], data)
	return pp, nil
}

// Occupancy returns the 64-bit bitmap of occupied squares.
func (pp PackedPosition) Occupancy() uint64 {
	var occ uint64
	for i := 0; i < 8; i++ {
		occ |= uint64(pp[i]) << (i * 8)
	}
	return occ
}

// PieceCount returns the number of pieces on the board.
func (pp PackedPosition) PieceCount() int {
	return bits.OnesCount64(pp.Occupancy())
}

// PieceAt returns the piece at the given square, or 0 if empty.
// Returns piece character: P, N, B, R, Q, K (white) or p, n, b, r, q, k (black).
func (pp PackedPosition) PieceAt(sq Square) byte {
	occ := pp.Occupancy()
	bit := uint64(1) << sq

	if occ&bit == 0 {
		return 0 // empty square
	}

	// Count pieces before this square to find index
	mask := bit - 1 // all bits below sq
	idx := bits.OnesCount64(occ & mask)

	// Read piece code
	byteIdx := 8 + idx/2
	var code byte
	if idx%2 == 0 {
		code = pp[byteIdx] & 0x0F
	} else {
		code = (pp[byteIdx] >> 4) & 0x0F
	}

	return codeToPiece(code)
}

// Pieces iterates over all pieces on the board.
// Calls fn with (square, piece character) for each piece.
func (pp PackedPosition) Pieces(fn func(sq Square, piece byte)) {
	occ := pp.Occupancy()
	idx := 0
	for sq := 0; sq < 64; sq++ {
		if occ&(1<<sq) == 0 {
			continue
		}

		byteIdx := 8 + idx/2
		var code byte
		if idx%2 == 0 {
			code = pp[byteIdx] & 0x0F
		} else {
			code = (pp[byteIdx] >> 4) & 0x0F
		}
		idx++

		fn(Square(sq), codeToPiece(code))
	}
}

// SideToMove returns the side to move (White or Black).
func (pp PackedPosition) SideToMove() Color {
	if pp[24]&flagBlackToMove != 0 {
		return Black
	}
	return White
}

// CastlingRights returns the castling rights byte.
// Bit 0: White kingside, Bit 1: White queenside,
// Bit 2: Black kingside, Bit 3: Black queenside.
func (pp PackedPosition) CastlingRights() byte {
	flags := pp[24]
	var castle byte
	if flags&flagWKCastle != 0 {
		castle |= 1 << 0
	}
	if flags&flagWQCastle != 0 {
		castle |= 1 << 1
	}
	if flags&flagBKCastle != 0 {
		castle |= 1 << 2
	}
	if flags&flagBQCastle != 0 {
		castle |= 1 << 3
	}
	return castle
}

// EPFile returns the en passant file (0-7), or -1 if none.
func (pp PackedPosition) EPFile() int {
	if pp[25] == noEPFile {
		return -1
	}
	return int(pp[25])
}

// EPSquare returns the en passant target square, or -1 if none.
func (pp PackedPosition) EPSquare() Square {
	file := pp.EPFile()
	if file < 0 {
		return -1
	}
	if pp.SideToMove() == White {
		return Square(40 + file) // rank 6
	}
	return Square(16 + file) // rank 3
}

// SetSideToMove sets the side to move.
func (pp *PackedPosition) SetSideToMove(c Color) {
	if c == Black {
		pp[24] |= flagBlackToMove
	} else {
		pp[24] &^= flagBlackToMove
	}
}

// SetCastlingRights sets the castling rights.
func (pp *PackedPosition) SetCastlingRights(castle byte) {
	// Clear existing castling flags
	pp[24] &^= flagWKCastle | flagWQCastle | flagBKCastle | flagBQCastle
	// Set new flags
	if castle&(1<<0) != 0 {
		pp[24] |= flagWKCastle
	}
	if castle&(1<<1) != 0 {
		pp[24] |= flagWQCastle
	}
	if castle&(1<<2) != 0 {
		pp[24] |= flagBKCastle
	}
	if castle&(1<<3) != 0 {
		pp[24] |= flagBQCastle
	}
}

// SetEPFile sets the en passant file (0-7), or -1 for none.
func (pp *PackedPosition) SetEPFile(file int) {
	if file < 0 || file > 7 {
		pp[25] = noEPFile
	} else {
		pp[25] = byte(file)
	}
}

// -----------------------------------------------------------------------------
// PackedFEN methods
// -----------------------------------------------------------------------------

// PackFEN encodes a GameState into a PackedFEN (29 bytes with all metadata).
func (gs *GameState) PackFEN() PackedFEN {
	var pf PackedFEN

	// Copy PackedPosition (26 bytes including flags and EP)
	pp := gs.Pack()
	copy(pf[:26], pp[:])

	// Pack halfmove clock (byte 26, capped at 255)
	if gs.Halfmove > 255 {
		pf[26] = 255
	} else {
		pf[26] = byte(gs.Halfmove)
	}

	// Pack fullmove number (bytes 27-28, little-endian, capped at 65535)
	fullmove := gs.Fullmove
	if fullmove > 65535 {
		fullmove = 65535
	}
	if fullmove < 1 {
		fullmove = 1
	}
	pf[27] = byte(fullmove & 0xFF)
	pf[28] = byte((fullmove >> 8) & 0xFF)

	return pf
}

// PackFEN encodes a FEN string into a PackedFEN.
func PackFEN(fen string) (PackedFEN, error) {
	gs, err := NewGame(fen)
	if err != nil {
		return PackedFEN{}, err
	}
	return gs.PackFEN(), nil
}

// Unpack decodes a PackedFEN into a GameState.
func (pf PackedFEN) Unpack() *GameState {
	// Unpack PackedPosition (includes board, flags, EP)
	var pp PackedPosition
	copy(pp[:], pf[:26])
	gs := pp.Unpack()

	// Unpack halfmove clock
	gs.Halfmove = int(pf[26])

	// Unpack fullmove number (little-endian)
	gs.Fullmove = int(pf[27]) | (int(pf[28]) << 8)

	return gs
}

// String returns the base64 URL-safe encoding of the packed FEN.
func (pf PackedFEN) String() string {
	return base64.RawURLEncoding.EncodeToString(pf[:])
}

// ToFEN converts the packed FEN to a FEN string.
func (pf PackedFEN) ToFEN() string {
	return pf.Unpack().ToFEN()
}

// ParsePackedFEN decodes a base64 URL-safe encoded packed FEN.
func ParsePackedFEN(s string) (PackedFEN, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return PackedFEN{}, fmt.Errorf("invalid base64: %w", err)
	}
	if len(data) != 29 {
		return PackedFEN{}, fmt.Errorf("packed FEN must be 29 bytes, got %d", len(data))
	}
	var pf PackedFEN
	copy(pf[:], data)
	return pf, nil
}

// ToPackedPosition extracts the position (without move counters).
func (pf PackedFEN) ToPackedPosition() PackedPosition {
	var pp PackedPosition
	copy(pp[:], pf[:26])
	return pp
}

// -----------------------------------------------------------------------------
// PackedFEN getters
// -----------------------------------------------------------------------------

// Occupancy returns the 64-bit bitmap of occupied squares.
func (pf PackedFEN) Occupancy() uint64 {
	return pf.ToPackedPosition().Occupancy()
}

// PieceCount returns the number of pieces on the board.
func (pf PackedFEN) PieceCount() int {
	return pf.ToPackedPosition().PieceCount()
}

// PieceAt returns the piece at the given square, or 0 if empty.
func (pf PackedFEN) PieceAt(sq Square) byte {
	return pf.ToPackedPosition().PieceAt(sq)
}

// SideToMove returns the side to move (White or Black).
func (pf PackedFEN) SideToMove() Color {
	if pf[24]&flagBlackToMove != 0 {
		return Black
	}
	return White
}

// CastlingRights returns the castling rights byte.
// Bit 0: White kingside, Bit 1: White queenside,
// Bit 2: Black kingside, Bit 3: Black queenside.
func (pf PackedFEN) CastlingRights() byte {
	flags := pf[24]
	var castle byte
	if flags&flagWKCastle != 0 {
		castle |= 1 << 0
	}
	if flags&flagWQCastle != 0 {
		castle |= 1 << 1
	}
	if flags&flagBKCastle != 0 {
		castle |= 1 << 2
	}
	if flags&flagBQCastle != 0 {
		castle |= 1 << 3
	}
	return castle
}

// EPFile returns the en passant file (0-7), or -1 if none.
func (pf PackedFEN) EPFile() int {
	if pf[25] == noEPFile {
		return -1
	}
	return int(pf[25])
}

// EPSquare returns the en passant target square, or -1 if none.
func (pf PackedFEN) EPSquare() Square {
	file := pf.EPFile()
	if file < 0 {
		return -1
	}
	if pf.SideToMove() == White {
		return Square(40 + file) // rank 6
	}
	return Square(16 + file) // rank 3
}

// Halfmove returns the halfmove clock (0-255).
func (pf PackedFEN) Halfmove() int {
	return int(pf[26])
}

// Fullmove returns the fullmove number (1-65535).
func (pf PackedFEN) Fullmove() int {
	return int(pf[27]) | (int(pf[28]) << 8)
}

// -----------------------------------------------------------------------------
// PackedFEN setters
// -----------------------------------------------------------------------------

// SetSideToMove sets the side to move.
func (pf *PackedFEN) SetSideToMove(c Color) {
	if c == Black {
		pf[24] |= flagBlackToMove
	} else {
		pf[24] &^= flagBlackToMove
	}
}

// SetCastlingRights sets the castling rights.
func (pf *PackedFEN) SetCastlingRights(castle byte) {
	// Clear existing castling flags
	pf[24] &^= flagWKCastle | flagWQCastle | flagBKCastle | flagBQCastle
	// Set new flags
	if castle&(1<<0) != 0 {
		pf[24] |= flagWKCastle
	}
	if castle&(1<<1) != 0 {
		pf[24] |= flagWQCastle
	}
	if castle&(1<<2) != 0 {
		pf[24] |= flagBKCastle
	}
	if castle&(1<<3) != 0 {
		pf[24] |= flagBQCastle
	}
}

// SetEPFile sets the en passant file (0-7), or -1 for none.
func (pf *PackedFEN) SetEPFile(file int) {
	if file < 0 || file > 7 {
		pf[25] = noEPFile
	} else {
		pf[25] = byte(file)
	}
}

// SetHalfmove sets the halfmove clock (capped at 255).
func (pf *PackedFEN) SetHalfmove(hm int) {
	if hm > 255 {
		pf[26] = 255
	} else if hm < 0 {
		pf[26] = 0
	} else {
		pf[26] = byte(hm)
	}
}

// SetFullmove sets the fullmove number (capped at 65535, minimum 1).
func (pf *PackedFEN) SetFullmove(fm int) {
	if fm > 65535 {
		fm = 65535
	}
	if fm < 1 {
		fm = 1
	}
	pf[27] = byte(fm & 0xFF)
	pf[28] = byte((fm >> 8) & 0xFF)
}

// -----------------------------------------------------------------------------
// Convenience functions
// -----------------------------------------------------------------------------

// PackedPositionFromFEN is a convenience function to get a base64 string from FEN.
// Returns just the board encoding (no metadata).
func PackedPositionFromFEN(fen string) (string, error) {
	gs, err := NewGame(fen)
	if err != nil {
		return "", err
	}
	return gs.Pack().String(), nil
}

// PackedFENFromFEN is a convenience function to get a base64 string from FEN.
// Returns the full FEN encoding (with metadata).
func PackedFENFromFEN(fen string) (string, error) {
	gs, err := NewGame(fen)
	if err != nil {
		return "", err
	}
	return gs.PackFEN().String(), nil
}
