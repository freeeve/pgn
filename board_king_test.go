package pgn

import (
	. "launchpad.net/gocheck"
)

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKingsideCastle(c *C) {
	b, err := NewBoardFEN("rnbq1rk1/pppp1ppp/5n2/2b1p3/4P3/3P1N2/PPP1BPPP/RNBQK2R w KQ - 3 5")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E1)
	c.Assert(move.To, Equals, G1)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueensideCastle(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2nqb3/3pp3/3PP3/2NQB3/PPP2PPP/R3KBNR w KQkq - 4 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O-O", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E1)
	c.Assert(move.To, Equals, B1)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKingsideCastleBad(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("O-O", White)
	c.Assert(err, Equals, ErrMoveThroughPiece)
	c.Assert(move, Equals, NilMove)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueensideCastleBad(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("O-O-O", White)
	c.Assert(err, Equals, ErrMoveThroughPiece)
	c.Assert(move, Equals, NilMove)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteQueensideCastleQueenCheck(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/R3KBNR w KQkq - 4 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O-O", White)
	c.Assert(err, Equals, ErrMoveThroughCheck)
	c.Assert(move, Equals, NilMove)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKingsideCastle(c *C) {
	b, err := NewBoardFEN("rnbqk2r/pppp1ppp/5n2/2b1p3/4P3/3P1N2/PPP1BPPP/RNBQK2R b KQkq - 2 4")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E8)
	c.Assert(move.To, Equals, G8)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueensideCastle(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2nqb3/3pp3/3PP3/2NQB3/PPP2PPP/2KR1BNR b kq - 5 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, E8)
	c.Assert(move.To, Equals, B8)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKingsideCastleBad(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("O-O", Black)
	c.Assert(err, Equals, ErrMoveThroughPiece)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueensideCastleBad(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("e4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("O-O-O", Black)
	c.Assert(err, Equals, ErrMoveThroughPiece)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackQueensideCastleQueenCheck(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/R2K1BNR b kq - 5 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O-O", Black)
	c.Assert(err, Equals, ErrMoveThroughCheck)
	c.Assert(move, Equals, NilMove)
}
