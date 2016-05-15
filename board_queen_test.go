package pgn

import (
	. "gopkg.in/check.v1"
)

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueenRank(c *C) {
	b, err := NewBoardFEN("4k3/ppp2q2/8/8/2Q3Q1/8/PPP5/4K3 w KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Qce4", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, C4)
	c.Assert(move.To, Equals, E4)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueenFile(c *C) {
	b, err := NewBoardFEN("4k3/ppp2q2/8/3Q4/8/8/PPP5/3QK3 w KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Q5d3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D5)
	c.Assert(move.To, Equals, D3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueenDiagonal(c *C) {
	b, err := NewBoardFEN("4k3/ppp2q2/8/7Q/8/8/PPP5/3QK3 w KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Qdf3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D1)
	c.Assert(move.To, Equals, F3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueenRank(c *C) {
	b, err := NewBoardFEN("4k3/ppp5/8/3Q4/1q4q1/8/PPP5/3QK3 b KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Qge4", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G4)
	c.Assert(move.To, Equals, E4)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueenFile(c *C) {
	b, err := NewBoardFEN("4k3/ppp2q2/8/3Q4/8/5q2/PPP5/3QK3 b KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Q7f5", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, F7)
	c.Assert(move.To, Equals, F5)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueenDiagonal(c *C) {
	b, err := NewBoardFEN("4k3/pppq4/8/3Q4/6q1/8/PPP5/3QK3 b KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Qde6", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D7)
	c.Assert(move.To, Equals, E6)
}
