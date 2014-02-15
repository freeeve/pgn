package pgn

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrAmbiguousMove       error = errors.New("pgn: ambiguous algebraic move")
	ErrUnknownMove         error = errors.New("pgn: unknown move")
	ErrAttackerNotFound    error = errors.New("pgn: attacker not found")
	ErrMoveFromEmptySquare error = errors.New("pgn: move from empty square")
	ErrMoveWrongColor      error = errors.New("pgn: move from wrong color")
	ErrMoveThroughPiece    error = errors.New("pgn: move through piece")
	ErrMoveThroughCheck    error = errors.New("pgn: move through check")
	ErrMoveIntoCheck       error = errors.New("pgn: move into check")
	ErrMoveInvalidCastle   error = errors.New("pgn: move invalid castle")
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

func (p Piece) Color() Color {
	if byte(p) >= byte('a') && byte(p) <= byte('z') {
		return Black
	}
	if byte(p) >= byte('A') && byte(p) <= byte('Z') {
		return White
	}
	return NoColor
}

type Color int8

const (
	NoColor Color = iota
	Black
	White
)

func (c Color) String() string {
	if c == White {
		return "w"
	} else if c == Black {

		return "b"
	}
	return " "
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

func (b *Board) MoveFromAlgebraic(str string, color Color) (Move, error) {
	if b.toMove != color {
		return NilMove, ErrMoveWrongColor
	}
	pos, err := ParsePosition(str)
	testPos := pos
	// if it's a raw position, it's a pawn move
	if err == nil {
		for {
			if color == White {
				testPos >>= 8
				if b.GetPiece(testPos) == WhitePawn {
					return Move{testPos, pos}, nil
				}
				if pos == NoPosition {
					return NilMove, fmt.Errorf("Position out of bounds")
				}
			} else {
				testPos <<= 8
				if b.GetPiece(testPos) == BlackPawn {
					return Move{testPos, pos}, nil
				}
				if pos == NoPosition {
					return NilMove, fmt.Errorf("Position out of bounds")
				}
			}
		}
	} else {
		// otherwise it's a non-pawn move
		switch str[0] {
		case 'O':
			if str == "O-O" {
				return b.getKingsideCastle(color)
			} else if str == "O-O-O" {
				return b.getQueensideCastle(color)
			} else {
				return NilMove, ErrUnknownMove
			}
		case 'N':
			pos, err := ParsePosition(str[len(str)-2 : len(str)])
			if err != nil {
				return NilMove, err
			}
			fromPos, err := b.findAttackingKnight(pos, color)
			if err != nil {
				return NilMove, err
			}
			return Move{fromPos, pos}, nil
		case 'B':
			pos, err := ParsePosition(str[len(str)-2 : len(str)])
			if err != nil {
				return NilMove, err
			}
			fromPos, err := b.findAttackingBishop(pos, color)
			if err != nil {
				return NilMove, err
			}
			return Move{fromPos, pos}, nil
		}
		return NilMove, ErrUnknownMove
	}
	return Move{D2, D4}, nil
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

func (b *Board) MakeMove(m Move) error {
	p := b.GetPiece(m.From)
	if p == Empty {
		return ErrMoveFromEmptySquare
	}
	if p.Color() != b.toMove {
		return ErrMoveWrongColor
	}
	take := b.GetPiece(m.To)
	if take != Empty {
		b.RemovePiece(m.To, take)
	}
	b.SetPiece(m.To, p)
	b.RemovePiece(m.From, p)
	switch b.toMove {
	case White:
		b.toMove = Black
	case Black:
		b.toMove = White
	}
	return nil
}

func (b *Board) RemovePiece(pos Position, p Piece) {
	switch p {
	case WhitePawn:
		b.wPawns &= ^uint64(pos)
	case BlackPawn:
		b.bPawns &= ^uint64(pos)
	case WhiteKnight:
		b.wKnights &= ^uint64(pos)
	case BlackKnight:
		b.bKnights &= ^uint64(pos)
	case WhiteBishop:
		b.wBishops &= ^uint64(pos)
	case BlackBishop:
		b.bBishops &= ^uint64(pos)
	case WhiteRook:
		b.wRooks &= ^uint64(pos)
	case BlackRook:
		b.bRooks &= ^uint64(pos)
	case WhiteQueen:
		b.wQueens &= ^uint64(pos)
	case BlackQueen:
		b.bQueens &= ^uint64(pos)
	case WhiteKing:
		b.wKings &= ^uint64(pos)
	case BlackKing:
		b.bKings &= ^uint64(pos)
	}
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

func (b Board) findAttackingKnight(pos Position, color Color) (Position, error) {
	count := 0
	r := pos.GetRank()
	f := pos.GetFile()
	retPos := NoPosition

	testPos := PositionFromFileRank(f+1, r+2)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f+1, r-2)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f+2, r+1)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f+2, r-1)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f-2, r-1)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f-2, r+1)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f-1, r-2)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	testPos = PositionFromFileRank(f-1, r+2)
	if testPos != NoPosition && b.checkKnightColor(testPos, color) {
		count++
		retPos = testPos
	}

	if count > 1 {
		return NoPosition, ErrAmbiguousMove
	}
	if count == 0 {
		return NoPosition, ErrAttackerNotFound
	}
	return retPos, nil
}

func (b Board) checkKnightColor(pos Position, color Color) bool {
	return (b.GetPiece(pos) == WhiteKnight && color == White) ||
		(b.GetPiece(pos) == BlackKnight && color == Black)
}

func (b Board) findAttackingBishop(pos Position, color Color) (Position, error) {
	count := 0
	retPos := NoPosition

	r := pos.GetRank()
	f := pos.GetFile()
	for {
		f--
		r--
		testPos := PositionFromFileRank(f, r)
		if b.checkBishopColor(testPos, color) {
			retPos = testPos
			count++
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.GetRank()
	f = pos.GetFile()
	for {
		f--
		r++
		testPos := PositionFromFileRank(f, r)
		if b.checkBishopColor(testPos, color) {
			retPos = testPos
			count++
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.GetRank()
	f = pos.GetFile()
	for {
		f++
		r++
		testPos := PositionFromFileRank(f, r)
		if b.checkBishopColor(testPos, color) {
			retPos = testPos
			count++
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.GetRank()
	f = pos.GetFile()
	for {
		f++
		r--
		testPos := PositionFromFileRank(f, r)
		if b.checkBishopColor(testPos, color) {
			retPos = testPos
			count++
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	if count > 1 {
		return NoPosition, ErrAmbiguousMove
	}
	if count == 0 {
		return NoPosition, ErrAttackerNotFound
	}
	return retPos, nil
}

func (b Board) checkBishopColor(pos Position, color Color) bool {
	return (b.GetPiece(pos) == WhiteBishop && color == White) ||
		(b.GetPiece(pos) == BlackBishop && color == Black)
}

func (b Board) containsPieceAt(pos Position) bool {
	return (uint64(b.wPawns)|uint64(b.bPawns)|uint64(b.wKnights)|uint64(b.bKnights)|
		uint64(b.wBishops)|uint64(b.bBishops)|uint64(b.wRooks)|uint64(b.bRooks)|
		uint64(b.wQueens)|uint64(b.bQueens)|uint64(b.wKings)|uint64(b.bKings))&uint64(pos) > 0
}

func (b Board) getKingsideCastle(color Color) (Move, error) {
	if color == White {
		if b.wCastle != Both && b.wCastle != King {
			return NilMove, ErrMoveInvalidCastle
		}
		if b.containsPieceAt(F1) || b.containsPieceAt(G1) {
			return NilMove, ErrMoveThroughPiece
		}
		if b.positionAttackedBy(F1, Black) || b.positionAttackedBy(G1, Black) {
			return NilMove, ErrMoveThroughCheck
		}
		return Move{E1, G1}, nil
	} else {
		if b.bCastle != Both && b.bCastle != King {
			return NilMove, ErrMoveInvalidCastle
		}
		if b.containsPieceAt(F8) || b.containsPieceAt(G8) {
			return NilMove, ErrMoveThroughPiece
		}
		if b.positionAttackedBy(F8, White) || b.positionAttackedBy(G8, White) {
			return NilMove, ErrMoveThroughCheck
		}
		return Move{E8, G8}, nil
	}
}

func (b Board) getQueensideCastle(color Color) (Move, error) {
	if color == White {
		if b.wCastle != Both && b.wCastle != Queen {
			return NilMove, ErrMoveInvalidCastle
		}
		if b.containsPieceAt(B1) || b.containsPieceAt(C1) || b.containsPieceAt(D1) {
			return NilMove, ErrMoveThroughPiece
		}
		if b.positionAttackedBy(B1, Black) || b.positionAttackedBy(C1, Black) || b.positionAttackedBy(D1, Black) {
			return NilMove, ErrMoveThroughCheck
		}
		return Move{E1, B1}, nil
	} else {
		if b.bCastle != Both && b.bCastle != Queen {
			return NilMove, ErrMoveInvalidCastle
		}
		if b.containsPieceAt(B8) || b.containsPieceAt(C8) || b.containsPieceAt(D8) {
			return NilMove, ErrMoveThroughPiece
		}
		if b.positionAttackedBy(B8, White) || b.positionAttackedBy(C8, White) || b.positionAttackedBy(D8, White) {
			return NilMove, ErrMoveThroughCheck
		}
		return Move{E8, B8}, nil
	}
}

func (b Board) positionAttackedBy(pos Position, color Color) bool {
	// TODO implement this
	if color == White {
		return false
	} else {
		return false
	}
}
