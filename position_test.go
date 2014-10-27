package pgn

import (
	. "gopkg.in/check.v1"
)

type PositionSuite struct{}

var _ = Suite(&PositionSuite{})

func (s *PositionSuite) TestGetRank1(c *C) {
	c.Assert(B1.GetRank(), Equals, Rank1)
}

func (s *PositionSuite) TestGetRank2(c *C) {
	c.Assert(A2.GetRank(), Equals, Rank2)
}

func (s *PositionSuite) TestGetRank3(c *C) {
	c.Assert(D3.GetRank(), Equals, Rank3)
}

func (s *PositionSuite) TestGetRank4(c *C) {
	c.Assert(C4.GetRank(), Equals, Rank4)
}

func (s *PositionSuite) TestGetRank5(c *C) {
	c.Assert(F5.GetRank(), Equals, Rank5)
}

func (s *PositionSuite) TestGetRank6(c *C) {
	c.Assert(E6.GetRank(), Equals, Rank6)
}

func (s *PositionSuite) TestGetRank7(c *C) {
	c.Assert(H7.GetRank(), Equals, Rank7)
}

func (s *PositionSuite) TestGetRank8(c *C) {
	c.Assert(G8.GetRank(), Equals, Rank8)
}

func (s *PositionSuite) TestGetNoRank(c *C) {
	c.Assert(NoPosition.GetRank(), Equals, NoRank)
}

func (s *PositionSuite) TestGetFileA(c *C) {
	c.Assert(A1.GetFile(), Equals, FileA)
}

func (s *PositionSuite) TestGetFileB(c *C) {
	c.Assert(B6.GetFile(), Equals, FileB)
}

func (s *PositionSuite) TestGetFileC(c *C) {
	c.Assert(C7.GetFile(), Equals, FileC)
}

func (s *PositionSuite) TestGetFileD(c *C) {
	c.Assert(D5.GetFile(), Equals, FileD)
}

func (s *PositionSuite) TestGetFileE(c *C) {
	c.Assert(E8.GetFile(), Equals, FileE)
}

func (s *PositionSuite) TestGetFileF(c *C) {
	c.Assert(F4.GetFile(), Equals, FileF)
}

func (s *PositionSuite) TestGetFileG(c *C) {
	c.Assert(G2.GetFile(), Equals, FileG)
}

func (s *PositionSuite) TestGetFileH(c *C) {
	c.Assert(H3.GetFile(), Equals, FileH)
}

func (s *PositionSuite) TestGetNoFile(c *C) {
	c.Assert(NoPosition.GetFile(), Equals, NoFile)
}

func (s *PositionSuite) TestPositionString(c *C) {
	c.Assert(A1.String(), Equals, "a1")
}
