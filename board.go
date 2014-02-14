package pgn

import (
	"errors"
	"fmt"
	"strconv"
)

type Board struct {
	wPawns        uint64
	bPawns        uint64
	wRooks        uint64
	bRooks        uint64
	wKnights      uint64
	bKnights      uint64
	wBishops      uint64
	bBishops      uint64
	wQueens       uint64
	bQueens       uint64
	wKings        uint64
	bKings        uint64
	lastMove      Move
	wCastle       CastleStatus
	bCastle       CastleStatus
	toMove        Color
	fullmove      int
	halfmoveClock int
}

type Piece byte

const (
	Empty       Piece = ' '
	BlackPawn   Piece = 'p'
	BlackKnight Piece = 'n'
	BlackBishop Piece = 'b'
	BlackRook   Piece = 'r'
	BlackQueen  Piece = 'q'
	BlackKing   Piece = 'k'
	WhitePawn   Piece = 'P'
	WhiteKnight Piece = 'N'
	WhiteBishop Piece = 'B'
	WhiteRook   Piece = 'R'
	WhiteQueen  Piece = 'Q'
	WhiteKing   Piece = 'K'
)

type Color int8

const (
	Black Color = iota
	White
)

func (c Color) String() string {
	if c == White {
		return "w"
	}
	return "b"
}

type CastleStatus int8

const (
	Both CastleStatus = iota
	None
	King
	Queen
)

func (cs CastleStatus) String(c Color) string {
	ret := ""
	switch cs {
	case Both:
		switch c {
		case Black:
			ret = "kq"
		case White:
			ret = "KQ"
		}
	case None:
		return "-"
	case King:
		switch c {
		case Black:
			ret = "k"
		case White:
			ret = "K"
		}
	case Queen:
		switch c {
		case Black:
			ret = "q"
		case White:
			ret = "Q"
		}
	}
	return ret
}

func (b *Board) MoveFromAlgebraic(str string, color Color) (*Move, error) {
	pos, err := ParsePosition(str)
	// if it's a raw position, it's a pawn move
	if err == nil {
		r := pos.GetRank()
		f := pos.GetFile()
		for {
			if color == White {
				r--
				if b.GetPiece(PositionFromFileRank(f, r)) == WhitePawn {
					return &Move{PositionFromFileRank(f, r), pos}, nil
				}
				if r == Rank('0') {
					return nil, fmt.Errorf("Rank out of bounds: %c", r)
				}
			} else {
				r++
				if b.GetPiece(PositionFromFileRank(f, r)) == BlackPawn {
					return &Move{PositionFromFileRank(f, r), pos}, nil
				}
				if r == Rank('9') {
					return nil, fmt.Errorf("Rank out of bounds: %c", r)
				}
			}
		}
	} else {
		// otherwise it's a non-pawn move
		return nil, errors.New("unsupported move")
	}
	return &Move{D2, D4}, nil
}

func NewBoard() *Board {
	b, _ := NewBoardFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return b
}

func NewBoardFEN(fen string) (*Board, error) {
	f, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	b := Board{}
	b.toMove = f.ToMove
	b.wCastle = f.WhiteCastleStatus
	b.bCastle = f.BlackCastleStatus
	b.fullmove = f.Fullmove
	b.halfmoveClock = f.HalfmoveClock

	x := byte('a')
	y := byte('8')
	for i := 0; i < len(f.FOR); i++ {
		// if we're at the end of the row
		if f.FOR[i] == '/' {
			x = 'a'
			y--
			continue
		} else if f.FOR[i] >= '1' && f.FOR[i] <= '8' {
			// if we have blank squares
			j, err := strconv.Atoi(string(f.FOR[i]))
			if err != nil {
				fmt.Println(err)
			}
			x += byte(j)
			continue
		} else {
			// if we have a piece
			pos, err := ParsePosition(fmt.Sprintf("%c%c", x, y))
			if err != nil {
				fmt.Println(err)
			}
			b.SetPiece(pos, Piece(f.FOR[i]))
			x++
		}
	}
	return &b, nil
}

func (b *Board) String() string {
	return FENFromBoard(b).String()
}

func (b *Board) SetPiece(pos Position, p Piece) {
	switch p {
	case WhitePawn:
		b.wPawns |= uint64(pos)
	case BlackPawn:
		b.bPawns |= uint64(pos)
	case WhiteKnight:
		b.wKnights |= uint64(pos)
	case BlackKnight:
		b.bKnights |= uint64(pos)
	case WhiteBishop:
		b.wBishops |= uint64(pos)
	case BlackBishop:
		b.bBishops |= uint64(pos)
	case WhiteRook:
		b.wRooks |= uint64(pos)
	case BlackRook:
		b.bRooks |= uint64(pos)
	case WhiteQueen:
		b.wQueens |= uint64(pos)
	case BlackQueen:
		b.bQueens |= uint64(pos)
	case WhiteKing:
		b.wKings |= uint64(pos)
	case BlackKing:
		b.bKings |= uint64(pos)
	}

}

func (b Board) GetPiece(p Position) Piece {
	if b.bPawns&uint64(p) != 0 {
		return BlackPawn
	}
	if b.wPawns&uint64(p) != 0 {
		return WhitePawn
	}
	if b.bKnights&uint64(p) != 0 {
		return BlackKnight
	}
	if b.wKnights&uint64(p) != 0 {
		return WhiteKnight
	}
	if b.bBishops&uint64(p) != 0 {
		return BlackBishop
	}
	if b.wBishops&uint64(p) != 0 {
		return WhiteBishop
	}
	if b.bRooks&uint64(p) != 0 {
		return BlackRook
	}
	if b.wRooks&uint64(p) != 0 {
		return WhiteRook
	}
	if b.bQueens&uint64(p) != 0 {
		return BlackQueen
	}
	if b.wQueens&uint64(p) != 0 {
		return WhiteQueen
	}
	if b.bKings&uint64(p) != 0 {
		return BlackKing
	}
	if b.wKings&uint64(p) != 0 {
		return WhiteKing
	}
	return Empty
}
