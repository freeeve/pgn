package pgn

import (
	. "launchpad.net/gocheck"
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
