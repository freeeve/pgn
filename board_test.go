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

func (s *BoardSuite) TestBoardColorWhitePawn(c *C) {
	c.Assert(WhitePawn.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteKnight(c *C) {
	c.Assert(WhiteKnight.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteBishop(c *C) {
	c.Assert(WhiteBishop.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteRook(c *C) {
	c.Assert(WhiteRook.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteQueen(c *C) {
	c.Assert(WhiteQueen.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorWhiteKing(c *C) {
	c.Assert(WhiteKing.Color(), Equals, White)
}

func (s *BoardSuite) TestBoardColorBlackPawn(c *C) {
	c.Assert(BlackPawn.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackKnight(c *C) {
	c.Assert(BlackKnight.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackBishop(c *C) {
	c.Assert(BlackBishop.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackRook(c *C) {
	c.Assert(BlackRook.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackQueen(c *C) {
	c.Assert(BlackQueen.Color(), Equals, Black)
}

func (s *BoardSuite) TestBoardColorBlackKing(c *C) {
	c.Assert(BlackKing.Color(), Equals, Black)
}

func (s *BoardSuite) TestNoColor(c *C) {
	c.Assert(Empty.Color(), Equals, NoColor)
}

func (s *BoardSuite) TestNoColorString(c *C) {
	c.Assert(Empty.Color().String(), Equals, " ")
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
	move, err := b.MoveFromAlgebraic("d4", White)
	c.Assert(err, IsNil)
	b.MakeMove(move)
	move, err = b.MoveFromAlgebraic("d5", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, D7)
	c.Assert(move.To, Equals, D5)
}

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
	c.Skip("not ready")
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
	c.Skip("not ready")
	b, err := NewBoardFEN("r3kbnr/ppp2ppp/2n4B/3pp1q1/3PP1Q1/2N4b/PPP2PPP/R2K1BNR b kq - 5 6")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("O-O-O", Black)
	c.Assert(err, Equals, ErrMoveThroughCheck)
	c.Assert(move, Equals, NilMove)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicWhiteKnight(c *C) {
	b := NewBoard()
	move, err := b.MoveFromAlgebraic("Nf3", White)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G1)
	c.Assert(move.To, Equals, F3)
}

func (s *BoardSuite) TestBoardMoveFromAlgebraicBlackKnight(c *C) {
	b, err := NewBoardFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1")
	c.Assert(err, IsNil)
	move, err := b.MoveFromAlgebraic("Nf6", Black)
	c.Assert(err, IsNil)
	c.Assert(move.From, Equals, G8)
	c.Assert(move.To, Equals, F6)
}

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

func (s *BoardSuite) TestBoardContainsPieceAtA1(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A1), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA2(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A2), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA3(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A3), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA4(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A4), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA5(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A5), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA6(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A6), Equals, false)
}

func (s *BoardSuite) TestBoardContainsPieceAtA7(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A7), Equals, true)
}

func (s *BoardSuite) TestBoardContainsPieceAtA8(c *C) {
	b := NewBoard()
	c.Assert(b.containsPieceAt(A8), Equals, true)
}
