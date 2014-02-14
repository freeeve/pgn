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

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhitePawn(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("d4", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D2)
	c.Assert(move.To, Equals, D4)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackPawn(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("d5", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D7)
	c.Assert(move.To, Equals, D5)
}
