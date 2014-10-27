package pgn

import (
	. "gopkg.in/check.v1"
)

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteRookRank(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/R3KBNR w KQkq - 4 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Rd1", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, A1)
	c.Assert(move.To, Equals, D1)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteRookFile(c *C) {
	b, err := NewBoardFEN("3rkbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/3RKBNR w Kk - 6 7")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Rd3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D1)
	c.Assert(move.To, Equals, D3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackRookRank(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/3RKBNR b Kkq - 5 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Rd8", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, A8)
	c.Assert(move.To, Equals, D8)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackRookFile(c *C) {
	b, err := NewBoardFEN("3rkbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2NR3b/PPP2PPP/4KBNR b Kk - 7 7")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Rd6", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D8)
	c.Assert(move.To, Equals, D6)
}
