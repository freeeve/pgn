package pgn

import (
	. "gopkg.in/check.v1"
)

type PositionSuite struct{}

var _ = Suite(&PositionSuite{})

func (s *PositionSuite) TestRank1(c *C) {
	c.Assert(B1.Rank(), Equals, Rank1)
}

func (s *PositionSuite) TestRank2(c *C) {
	c.Assert(A2.Rank(), Equals, Rank2)
}

func (s *PositionSuite) TestRank3(c *C) {
	c.Assert(D3.Rank(), Equals, Rank3)
}

func (s *PositionSuite) TestRank4(c *C) {
	c.Assert(C4.Rank(), Equals, Rank4)
}

func (s *PositionSuite) TestRank5(c *C) {
	c.Assert(F5.Rank(), Equals, Rank5)
}

func (s *PositionSuite) TestRank6(c *C) {
	c.Assert(E6.Rank(), Equals, Rank6)
}

func (s *PositionSuite) TestRank7(c *C) {
	c.Assert(H7.Rank(), Equals, Rank7)
}

func (s *PositionSuite) TestRank8(c *C) {
	c.Assert(G8.Rank(), Equals, Rank8)
}

func (s *PositionSuite) TestNoRank(c *C) {
	c.Assert(NoPosition.Rank(), Equals, NoPosition)
}

func (s *PositionSuite) TestFileA(c *C) {
	c.Assert(A1.File(), Equals, FileA)
}

func (s *PositionSuite) TestFileB(c *C) {
	c.Assert(B6.File(), Equals, FileB)
}

func (s *PositionSuite) TestFileC(c *C) {
	c.Assert(C7.File(), Equals, FileC)
}

func (s *PositionSuite) TestFileD(c *C) {
	c.Assert(D5.File(), Equals, FileD)
}

func (s *PositionSuite) TestFileE(c *C) {
	c.Assert(E8.File(), Equals, FileE)
}

func (s *PositionSuite) TestFileF(c *C) {
	c.Assert(F4.File(), Equals, FileF)
}

func (s *PositionSuite) TestFileG(c *C) {
	c.Assert(G2.File(), Equals, FileG)
}

func (s *PositionSuite) TestFileH(c *C) {
	c.Assert(H3.File(), Equals, FileH)
}

func (s *PositionSuite) TestNoFile(c *C) {
	c.Assert(NoPosition.File(), Equals, NoPosition)
}

func (s *PositionSuite) TestPositionString(c *C) {
	c.Assert(A1.String(), Equals, "a1")
}

func (s *PositionSuite) TestParsePosition(c *C) {
	p, err := ParsePosition("c3")
	c.Assert(err, IsNil)
	c.Assert(p, Equals, C3)
}

func (s *PositionSuite) TestParsePosition_uppercase(c *C) {
	p, err := ParsePosition("H8")
	c.Assert(err, IsNil)
	c.Assert(p, Equals, H8)
}

func (s *PositionSuite) TestParsePositionFile(c *C) {
	p, err := ParsePosition("d")
	c.Assert(err, IsNil)
	c.Assert(p, Equals, FileD)
}
