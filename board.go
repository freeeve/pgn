package pgn

type Board struct {
	wPawns   uint64
	bPawns   uint64
	wRooks   uint64
	bRooks   uint64
	wKnights uint64
	bKnights uint64
	wBishops uint64
	bBishops uint64
	wQueens  uint64
	bQueens  uint64
	wKings   uint64
	bKings   uint64
	lastMove Move
	wCastle  CastleStatus
	bCastle  CastleStatus
}

type Color int8

const (
	Black Color = iota
	White
)

type CastleStatus int8

const (
	Both CastleStatus = iota
	King
	Queen
)

func (b *Board) MoveFromAlgebraic(str string, color Color) *Move {
	return &Move{"d1", "d4"}
}

func NewBoardFEN(fen string) *Board {
	return &Board{}
}

func (b *Board) String() string {
	return "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
}
