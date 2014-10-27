package pgn

import (
	. "gopkg.in/check.v1"
)

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKnight(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("Nf3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G1)
	c.Assert(move.To, Equals, F3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKnight(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Nf6", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G8)
	c.Assert(move.To, Equals, F6)
}
