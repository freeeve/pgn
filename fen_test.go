package pgn

import (
	. "launchpad.net/gocheck"
)

type FENSuite struct{}

var _ = Suite(&FENSuite{})

func (s *FENSuite) TestParse(c *C) {
	fen, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(fen.FOR, Equals, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR")
	c.Assert(fen.ToMove, Equals, White)
	c.Assert(fen.WhiteCastleStatus, Equals, Both)
	c.Assert(fen.BlackCastleStatus, Equals, Both)
	c.Assert(fen.EnPassantVulnerable, Equals, NoPosition)
	c.Assert(fen.HalfmoveClock, Equals, 0)
	c.Assert(fen.Fullmove, Equals, 0)
}
