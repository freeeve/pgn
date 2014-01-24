package pgn

import (
	. "launchpad.net/gocheck"
)

type BoardSuite struct{}

var _ = Suite(&BoardSuite{})

func (s *BoardSuite) TestBoardString(c *C) {
	b := Board{}
	c.Assert(b.String(), Equals, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
}

func (s *BoardSuite) TestBoardNewFEN(c *C) {
	c.Skip("skipping until fen is farther along")
	b := NewBoardFEN("rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")
	c.Assert(b.String(), Equals, "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")
}
