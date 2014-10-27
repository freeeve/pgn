package pgn

import (
	. "gopkg.in/check.v1"
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

func (s *BoardSuite) TestNoColor(c *C) {
	c.Assert(NoPiece.Color(), Equals, NoColor)
}

func (s *BoardSuite) TestNoColorString(c *C) {
	c.Assert(NoPiece.Color().String(), Equals, " ")
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

func (s *BoardSuite) TestBoardMoveCoord(c *C) {
	b := NewBoard()
	err := b.MakeCoordMove("e2e4")
	c.Assert(err, IsNil)
	c.Assert(b.String(), Equals, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
}

func (s *BoardSuite) TestBoardMoveCoordPromote(c *C) {
	b, err := NewBoardFEN("rn3bnr/pppPkppp/8/4pb2/3P3q/8/PPP2PPP/RNBQKBNR w KQ - 1 6")
	c.Assert(err, IsNil)
	err = b.MakeCoordMove("d7d8n")
	c.Assert(err, IsNil)
	c.Assert(b.String(), Equals, "rn1N1bnr/ppp1kppp/8/4pb2/3P3q/8/PPP2PPP/RNBQKBNR b KQ - 0 6")
}

func (s *BoardSuite) TestCheckmateIsTrimmed(c *C) {
	b, err := NewBoardFEN("7k/3P1K1p/6pB/p4p2/8/P6P/4p1P1/8 w - - 0 43")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("d8=R#", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D7)
	c.Assert(move.To, Equals, D8)
	c.Assert(move.Promote, Equals, BlackRook)
}
