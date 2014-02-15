package pgn

import (
	. "launchpad.net/gocheck"
)

type BoardSuite struct{}

var _ = Suite(&BoardSuite{})

func (s *BoardSuite) TestBoardString(c *C) {
	b := NewBoard()
	c.Assert(b.String(), Equals, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func (s *BoardSuite) TestBoardNewFEN(c *C) {
	b, _ := NewBoardFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")
	c.Assert(b.String(), Equals, "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")
}

func (s *BoardSuite) TestBoardColorWhitePawn(c *C) {
	c.Assert(WhitePawn.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteKnight(c *C) {
	c.Assert(WhiteKnight.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteBishop(c *C) {
	c.Assert(WhiteBishop.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteRook(c *C) {
	c.Assert(WhiteRook.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteQueen(c *C) {
	c.Assert(WhiteQueen.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteKing(c *C) {
	c.Assert(WhiteKing.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorBlackPawn(c *C) {
	c.Assert(BlackPawn.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackKnight(c *C) {
	c.Assert(BlackKnight.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackBishop(c *C) {
	c.Assert(BlackBishop.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackRook(c *C) {
	c.Assert(BlackRook.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackQueen(c *C) {
	c.Assert(BlackQueen.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackKing(c *C) {
	c.Assert(BlackKing.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhitePawn(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("d4", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D2)
	c.Assert(move.To, Equals, D4)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackPawn(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("d4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("d5", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D7)
	c.Assert(move.To, Equals, D5)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKingsideCastle(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("O-O", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E1)
	c.Assert(move.To, Equals, G1)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueensideCastle(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("O-O-O", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E1)
	c.Assert(move.To, Equals, B1)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKingsideCastle(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	// TODO, make more moves so it's actually possible
	move, err = b.MoveFromAlgebraic("O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E8)
	c.Assert(move.To, Equals, G8)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueensideCastle(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("O-O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E8)
	c.Assert(move.To, Equals, B8)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKnight(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("Nf3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G1)
	c.Assert(move.To, Equals, F3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKnight(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("Nf6", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G8)
	c.Assert(move.To, Equals, F6)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackBishop(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("d5", Black)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("d4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("Bg4", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, C8)
	c.Assert(move.To, Equals, G4)
}

func (s *BoardSuite) TestBoardContainsPieceAtA1(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A1), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA2(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A2), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA3(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A3), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA4(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A4), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA5(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A5), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA6(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A6), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA7(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A7), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA8(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A8), Equals, true)
}
