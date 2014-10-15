package pgn

import (
	. "gopkg.in/check.v1"
)

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackBishop(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/ppp1pppp/8/3p4/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 2")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Bg4", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, C8)
	c.Assert(move.To, Equals, G4)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackBishopBad(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/ppp1pppp/8/3p4/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 2")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Bg5", Black)
	c.Assert(err, Equals, ErrAttackerNotFound)
	c.Assert(move, Equals, NilMove)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackBishopAmbiguous(c *C) {
	b, err := NewBoardFEN("r5nr/p2k2pp/5p2/3b4/P7/b1B5/5PPP/2b2K1R b - - 6 26")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Bb2", Black)
	c.Assert(err, Equals, ErrAmbiguousMove)
	c.Assert(move, Equals, NilMove)
}
