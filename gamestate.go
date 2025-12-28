// gamestate.go defines the bitboard-based game state and FEN parsing.
//
// GameState holds the complete state of a chess position using bitboards
// for efficient piece lookup and move generation.

package pgn

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	v2WPawn = iota
	v2WKnight
	v2WBishop
	v2WRook
	v2WQueen
	v2WKing
	v2BPawn
	v2BKnight
	v2BBishop
	v2BRook
	v2BQueen
	v2BKing
	v2PieceCount
)

// GameState represents a chess position using bitboards.
type GameState struct {
	pieces     [v2PieceCount]Bitboard
	occ        [2]Bitboard
	occAll     Bitboard
	SideToMove Color
	Castle     uint8
	EP         Square
	Halfmove   int
	Fullmove   int
	FEN        string
}

// startingFEN is the standard chess starting position.
const startingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// NewStartingPosition creates a new game at the standard starting position.
func NewStartingPosition() *GameState {
	gs, _ := NewGame(startingFEN)
	return gs
}

// NewGame parses a FEN string into a GameState.
func NewGame(fen string) (*GameState, error) {
	if fen == "" {
		return nil, errors.New("empty FEN")
	}
	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid FEN (need at least 4 fields): %q", fen)
	}
	pos := &GameState{EP: -1, Fullmove: 1}

	rank := 7
	file := 0
	for _, ch := range parts[0] {
		if ch == '/' {
			rank--
			file = 0
			continue
		}
		if ch >= '1' && ch <= '8' {
			file += int(ch - '0')
			continue
		}
		if file > 7 || rank < 0 {
			return nil, fmt.Errorf("invalid board coords in FEN: %q", fen)
		}
		sq := Square(rank*8 + file)
		switch ch {
		case 'P':
			pos.setPiece(v2WPawn, sq)
		case 'N':
			pos.setPiece(v2WKnight, sq)
		case 'B':
			pos.setPiece(v2WBishop, sq)
		case 'R':
			pos.setPiece(v2WRook, sq)
		case 'Q':
			pos.setPiece(v2WQueen, sq)
		case 'K':
			pos.setPiece(v2WKing, sq)
		case 'p':
			pos.setPiece(v2BPawn, sq)
		case 'n':
			pos.setPiece(v2BKnight, sq)
		case 'b':
			pos.setPiece(v2BBishop, sq)
		case 'r':
			pos.setPiece(v2BRook, sq)
		case 'q':
			pos.setPiece(v2BQueen, sq)
		case 'k':
			pos.setPiece(v2BKing, sq)
		default:
			return nil, fmt.Errorf("invalid piece char %c in FEN: %q", ch, fen)
		}
		file++
	}

	switch parts[1] {
	case "w":
		pos.SideToMove = White
	case "b":
		pos.SideToMove = Black
	default:
		return nil, fmt.Errorf("invalid side to move: %q", parts[1])
	}

	if parts[2] != "-" {
		for _, ch := range parts[2] {
			switch ch {
			case 'K':
				pos.Castle |= 1 << 0
			case 'Q':
				pos.Castle |= 1 << 1
			case 'k':
				pos.Castle |= 1 << 2
			case 'q':
				pos.Castle |= 1 << 3
			default:
				return nil, fmt.Errorf("invalid castle char %c", ch)
			}
		}
	}

	if parts[3] != "-" {
		sq, err := parseSquare(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid ep square: %v", err)
		}
		pos.EP = sq
	}

	if len(parts) >= 5 {
		hm, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, fmt.Errorf("invalid halfmove clock: %v", err)
		}
		pos.Halfmove = hm
	}
	if len(parts) >= 6 {
		fm, err := strconv.Atoi(parts[5])
		if err != nil {
			return nil, fmt.Errorf("invalid fullmove number: %v", err)
		}
		pos.Fullmove = fm
		if pos.Fullmove <= 0 {
			pos.Fullmove = 1
		}
	} else {
		pos.Fullmove = 1
	}

	pos.FEN = fen
	return pos, nil
}

// ToFEN returns the FEN string for the position.
func (p *GameState) ToFEN() string {
	if p.FEN != "" {
		return p.FEN
	}
	board := make([]byte, 0, 64*2)
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			sq := Square(rank*8 + file)
			ch := pieceAtV2(p, sq)
			if ch == 0 {
				empty++
			} else {
				if empty > 0 {
					board = append(board, byte('0'+empty))
					empty = 0
				}
				board = append(board, ch)
			}
		}
		if empty > 0 {
			board = append(board, byte('0'+empty))
		}
		if rank != 0 {
			board = append(board, '/')
		}
	}
	side := "w"
	if p.SideToMove == Black {
		side = "b"
	}
	castle := "-"
	cmask := []struct {
		bit uint8
		ch  byte
	}{
		{1 << 0, 'K'},
		{1 << 1, 'Q'},
		{1 << 2, 'k'},
		{1 << 3, 'q'},
	}
	var cbuf []byte
	for _, c := range cmask {
		if p.Castle&c.bit != 0 {
			cbuf = append(cbuf, c.ch)
		}
	}
	if len(cbuf) > 0 {
		castle = string(cbuf)
	}
	ep := "-"
	if p.EP >= 0 && p.EP < 64 {
		file := byte('a' + (p.EP % 8))
		rank := byte('1' + (p.EP / 8))
		ep = string([]byte{file, rank})
	}
	return fmt.Sprintf("%s %s %s %s %d %d", string(board), side, castle, ep, p.Halfmove, p.Fullmove)
}

func (p *GameState) setPiece(idx int, sq Square) {
	p.pieces[idx] |= 1 << uint(sq)
	color := White
	if idx >= v2BPawn {
		color = Black
	}
	p.occ[color] |= 1 << uint(sq)
	p.occAll |= 1 << uint(sq)
}

func parseSquare(str string) (Square, error) {
	if len(str) != 2 {
		return -1, fmt.Errorf("bad square %q", str)
	}
	file := int(str[0] - 'a')
	rank := int(str[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return -1, fmt.Errorf("bad square %q", str)
	}
	return Square(rank*8 + file), nil
}

// PieceAt returns the piece character at the given square.
// Returns 0 for empty, uppercase for white, lowercase for black.
func (p *GameState) PieceAt(sq Square) byte {
	return pieceAtV2(p, sq)
}

func pieceAtV2(p *GameState, sq Square) byte {
	mask := Bitboard(1) << uint(sq)
	switch {
	case p.pieces[v2WPawn]&mask != 0:
		return 'P'
	case p.pieces[v2WKnight]&mask != 0:
		return 'N'
	case p.pieces[v2WBishop]&mask != 0:
		return 'B'
	case p.pieces[v2WRook]&mask != 0:
		return 'R'
	case p.pieces[v2WQueen]&mask != 0:
		return 'Q'
	case p.pieces[v2WKing]&mask != 0:
		return 'K'
	case p.pieces[v2BPawn]&mask != 0:
		return 'p'
	case p.pieces[v2BKnight]&mask != 0:
		return 'n'
	case p.pieces[v2BBishop]&mask != 0:
		return 'b'
	case p.pieces[v2BRook]&mask != 0:
		return 'r'
	case p.pieces[v2BQueen]&mask != 0:
		return 'q'
	case p.pieces[v2BKing]&mask != 0:
		return 'k'
	}
	return 0
}

// String returns a human-readable board diagram.
func (p *GameState) String() string {
	var sb strings.Builder
	sb.WriteString("  a b c d e f g h\n")
	for rank := 7; rank >= 0; rank-- {
		sb.WriteByte(byte('1' + rank))
		sb.WriteByte(' ')
		for file := 0; file < 8; file++ {
			sq := Square(rank*8 + file)
			ch := pieceAtV2(p, sq)
			if ch == 0 {
				ch = '.'
			}
			sb.WriteByte(ch)
			sb.WriteByte(' ')
		}
		sb.WriteByte(byte('1' + rank))
		sb.WriteByte('\n')
	}
	sb.WriteString("  a b c d e f g h\n")
	return sb.String()
}

// findKing returns the square of the king for the given color.
func (p *GameState) findKing(color Color) Square {
	var bb Bitboard
	if color == White {
		bb = p.pieces[v2WKing]
	} else {
		bb = p.pieces[v2BKing]
	}
	if bb == 0 {
		return -1
	}
	return Square(bitscanForward(bb))
}

// IsInCheck returns true if the side to move is in check.
func (p *GameState) IsInCheck() bool {
	kingSq := p.findKing(p.SideToMove)
	if kingSq < 0 {
		return false
	}
	return squareAttacked(p, kingSq, p.SideToMove^1)
}

// Copy returns a deep copy of the game state.
func (p *GameState) Copy() *GameState {
	cp := *p
	return &cp
}
