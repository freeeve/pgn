// packed_position.go provides a compact 34-byte position encoding for URLs and databases.
//
// PackedPosition encodes a chess position into 34 bytes:
//   - 32 bytes: 64 squares × 4 bits each (2 squares per byte, A1-H8 order)
//   - 1 byte: flags (side to move, castling rights)
//   - 1 byte: en-passant file (0-7, or 0xFF for none)
//
// The base64 encoding produces a 46-character URL-safe string.

package pgn

import (
	"encoding/base64"
	"fmt"
)

// PackedPosition is a compact 34-byte representation of a chess position.
type PackedPosition [34]byte

const (
	ppEmpty = 0
	ppWP    = 1
	ppWN    = 2
	ppWB    = 3
	ppWR    = 4
	ppWQ    = 5
	ppWK    = 6
	ppBP    = 7
	ppBN    = 8
	ppBB    = 9
	ppBR    = 10
	ppBQ    = 11
	ppBK    = 12
)

const (
	ppFlagSideToMove = 1 << 0
	ppFlagWKCastle   = 1 << 1
	ppFlagWQCastle   = 1 << 2
	ppFlagBKCastle   = 1 << 3
	ppFlagBQCastle   = 1 << 4
	ppNoEP           = 0xFF
)

// Pack encodes a GameState into a PackedPosition.
func (gs *GameState) Pack() PackedPosition {
	var pp PackedPosition

	for i := 0; i < 64; i++ {
		sq := Square(i)
		p := gs.PieceAt(sq)
		var code byte
		switch p {
		case 'P':
			code = ppWP
		case 'N':
			code = ppWN
		case 'B':
			code = ppWB
		case 'R':
			code = ppWR
		case 'Q':
			code = ppWQ
		case 'K':
			code = ppWK
		case 'p':
			code = ppBP
		case 'n':
			code = ppBN
		case 'b':
			code = ppBB
		case 'r':
			code = ppBR
		case 'q':
			code = ppBQ
		case 'k':
			code = ppBK
		default:
			code = ppEmpty
		}

		byteIdx := i / 2
		if i%2 == 0 {
			pp[byteIdx] = code & 0x0F
		} else {
			pp[byteIdx] |= (code & 0x0F) << 4
		}
	}

	var flags byte
	if gs.SideToMove == Black {
		flags |= ppFlagSideToMove
	}
	if gs.Castle&(1<<0) != 0 {
		flags |= ppFlagWKCastle
	}
	if gs.Castle&(1<<1) != 0 {
		flags |= ppFlagWQCastle
	}
	if gs.Castle&(1<<2) != 0 {
		flags |= ppFlagBKCastle
	}
	if gs.Castle&(1<<3) != 0 {
		flags |= ppFlagBQCastle
	}
	pp[32] = flags

	if gs.EP >= 0 && gs.EP < 64 {
		pp[33] = byte(gs.EP % 8)
	} else {
		pp[33] = ppNoEP
	}

	return pp
}

// PackFEN encodes a FEN string into a PackedPosition.
func PackFEN(fen string) (PackedPosition, error) {
	gs, err := NewGame(fen)
	if err != nil {
		return PackedPosition{}, err
	}
	return gs.Pack(), nil
}

// Unpack decodes a PackedPosition into a GameState.
func (pp PackedPosition) Unpack() *GameState {
	gs := &GameState{EP: -1, Fullmove: 1}

	for i := 0; i < 64; i++ {
		byteIdx := i / 2
		var code byte
		if i%2 == 0 {
			code = pp[byteIdx] & 0x0F
		} else {
			code = (pp[byteIdx] >> 4) & 0x0F
		}

		if code == ppEmpty {
			continue
		}

		sq := Square(i)
		var idx int
		switch code {
		case ppWP:
			idx = v2WPawn
		case ppWN:
			idx = v2WKnight
		case ppWB:
			idx = v2WBishop
		case ppWR:
			idx = v2WRook
		case ppWQ:
			idx = v2WQueen
		case ppWK:
			idx = v2WKing
		case ppBP:
			idx = v2BPawn
		case ppBN:
			idx = v2BKnight
		case ppBB:
			idx = v2BBishop
		case ppBR:
			idx = v2BRook
		case ppBQ:
			idx = v2BQueen
		case ppBK:
			idx = v2BKing
		}
		gs.setPiece(idx, sq)
	}

	flags := pp[32]
	if flags&ppFlagSideToMove != 0 {
		gs.SideToMove = Black
	} else {
		gs.SideToMove = White
	}

	if flags&ppFlagWKCastle != 0 {
		gs.Castle |= 1 << 0
	}
	if flags&ppFlagWQCastle != 0 {
		gs.Castle |= 1 << 1
	}
	if flags&ppFlagBKCastle != 0 {
		gs.Castle |= 1 << 2
	}
	if flags&ppFlagBQCastle != 0 {
		gs.Castle |= 1 << 3
	}

	epFile := pp[33]
	if epFile != ppNoEP && epFile < 8 {
		if gs.SideToMove == White {
			gs.EP = Square(40 + int(epFile))
		} else {
			gs.EP = Square(16 + int(epFile))
		}
	}

	return gs
}

// String returns the base64 URL-safe encoding of the packed position.
func (pp PackedPosition) String() string {
	return base64.RawURLEncoding.EncodeToString(pp[:])
}

// ToFEN converts the packed position to a FEN string.
func (pp PackedPosition) ToFEN() string {
	return pp.Unpack().ToFEN()
}

// ParsePackedPosition decodes a base64 URL-safe encoded packed position.
func ParsePackedPosition(s string) (PackedPosition, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return PackedPosition{}, fmt.Errorf("invalid base64: %w", err)
	}
	if len(data) != 34 {
		return PackedPosition{}, fmt.Errorf("packed position must be 34 bytes, got %d", len(data))
	}
	var pp PackedPosition
	copy(pp[:], data)
	return pp, nil
}

// PackedPositionFromFEN is a convenience function to get a base64 string from FEN.
func PackedPositionFromFEN(fen string) (string, error) {
	pp, err := PackFEN(fen)
	if err != nil {
		return "", err
	}
	return pp.String(), nil
}

