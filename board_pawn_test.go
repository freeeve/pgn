package pgn

import (
	. "gopkg.in/check.v1"
)

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

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhitePawnTakesPromoteQueen(c *C) {
	b, err := NewBoardFEN("n1b2rk1/1P1p1pbp/5np1/r2p4/2p2N2/4P3/P1B2PPP/RN1RB1K1 w KQkq - 1 5")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("bxa8=Q", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, B7)
	c.Assert(move.To, Equals, A8)
	c.Assert(move.Promote, Equals, BlackQueen)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackPawnTakesEnPassant(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/ppp1pppp/8/4P3/2Pp4/8/PP1P1PPP/RNBQKBNR b KQkq c3 0 3")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("dxc3", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D4)
	c.Assert(move.To, Equals, C3)
	c.Assert(move.San, Equals, "dxc3")
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhitePawnTakesEnPassant(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/pppp1pp1/8/4p1Pp/8/8/PPPPPP1P/RNBQKBNR w KQkq h6 0 3")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("gxh6", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G5)
	c.Assert(move.To, Equals, H6)
	c.Assert(move.San, Equals, "gxh6")
}
