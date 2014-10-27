package pgn

import (
	. "gopkg.in/check.v1"
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
	c.Assert(fen.Fullmove, Equals, 1)
}

func (s *FENSuite) TestParseNoWhiteCastle(c *C) {
	fen, err := ParseFEN("rn3bnr/pppPkppp/8/4pb2/3P3q/8/PPP1KPPP/RNBQ1BNR b - - 2 6")
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(fen.FOR, Equals, "rn3bnr/pppPkppp/8/4pb2/3P3q/8/PPP1KPPP/RNBQ1BNR")
	c.Assert(fen.ToMove, Equals, Black)
	c.Assert(fen.WhiteCastleStatus, Equals, None)
	c.Assert(fen.BlackCastleStatus, Equals, None)
	c.Assert(fen.EnPassantVulnerable, Equals, NoPosition)
	c.Assert(fen.HalfmoveClock, Equals, 2)
	c.Assert(fen.Fullmove, Equals, 6)
}

func (s *FENSuite) TestParseNoBlackCastle(c *C) {
	fen, err := ParseFEN("rn3bnr/pppPkppp/8/4pb2/3P3q/8/PPP2PPP/RNBQKBNR w KQ - 1 6")
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(fen.FOR, Equals, "rn3bnr/pppPkppp/8/4pb2/3P3q/8/PPP2PPP/RNBQKBNR")
	c.Assert(fen.ToMove, Equals, White)
	c.Assert(fen.WhiteCastleStatus, Equals, Both)
	c.Assert(fen.BlackCastleStatus, Equals, None)
	c.Assert(fen.EnPassantVulnerable, Equals, NoPosition)
	c.Assert(fen.HalfmoveClock, Equals, 1)
	c.Assert(fen.Fullmove, Equals, 6)
}
