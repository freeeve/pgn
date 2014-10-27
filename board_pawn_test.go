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
