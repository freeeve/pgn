package pgn

import (
	. "gopkg.in/check.v1"
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
	c.Assert(move.To, Equals, C1)
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
	c.Assert(move.To, Equals, C8)
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

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKingG8G7(c *C) {
	b, err := NewBoardFEN("rnbq1rk1/pppp1p1p/5npb/4p3/4P3/5NPB/PPPP1PKP/RNBQ1R2 b - - 5 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Kg7", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G8)
	c.Assert(move.To, Equals, G7)
}

func (s *BoardSuite) TestBoardMakeAlgebraicMoveWhiteKingsideCastle(c *C) {
	b, err := NewBoardFEN("rn1qr1k1/p4pbp/bp1p1np1/2pP4/4PB2/2N5/PP1NBPPP/R2QK2R w KQkq - 0 1")
	c.Assert(err, IsNil)
	err = b.MakeAlgebraicMove("O-O", White)
	c.Assert(err, IsNil)
	c.Assert(b.GetPiece(G1), Equals, WhiteKing)
	c.Assert(b.GetPiece(F1), Equals, WhiteRook)
	c.Assert(b.GetPiece(E1), Equals, NoPiece)
	c.Assert(b.GetPiece(H1), Equals, NoPiece)
}

func (s *BoardSuite) TestBoardMakeAlgebraicMoveBlackKingsideCastle(c *C) {
	b, err := NewBoardFEN("rnbqk2r/pppp1ppp/5n2/2b1p3/4P3/3P1N2/PPP1BPPP/RNBQK2R b KQkq - 2 4")
	c.Assert(err, IsNil)
	err = b.MakeAlgebraicMove("O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(b.GetPiece(G8), Equals, BlackKing)
	c.Assert(b.GetPiece(F8), Equals, BlackRook)
	c.Assert(b.GetPiece(E8), Equals, NoPiece)
	c.Assert(b.GetPiece(H8), Equals, NoPiece)
}

func (s *BoardSuite) TestBoardMakeAlgebraicMoveWhiteQueensideCastle(c *C) {
	b, err := NewBoardFEN("r2qr1k1/p4pbp/bpnp1np1/2pP4/4PB2/2N5/PPQNBPPP/R3K2R w KQkq - 2 2")
	c.Assert(err, IsNil)
	err = b.MakeAlgebraicMove("O-O-O", White)
	c.Assert(err, IsNil)
	c.Assert(b.GetPiece(C1), Equals, WhiteKing)
	c.Assert(b.GetPiece(D1), Equals, WhiteRook)
	c.Assert(b.GetPiece(A1), Equals, NoPiece)
	c.Assert(b.GetPiece(E1), Equals, NoPiece)
}

func (s *BoardSuite) TestBoardMakeAlgebraicMoveBlackQueensideCastle(c *C) {
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n1b3/3pp3/3PP2q/2NQBN2/PPP2PPP/R3KB1R b KQkq - 6 6")
	c.Assert(err, IsNil)
	err = b.MakeAlgebraicMove("O-O-O", Black)
	c.Assert(err, IsNil)
	c.Assert(b.GetPiece(C8), Equals, BlackKing)
	c.Assert(b.GetPiece(D8), Equals, BlackRook)
	c.Assert(b.GetPiece(E8), Equals, NoPiece)
	c.Assert(b.GetPiece(A8), Equals, NoPiece)
}
