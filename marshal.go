package pgn

import (
	"encoding/gob"
	"os"
)

type exportedBoard struct {
	WPawns        uint64
	BPawns        uint64
	WRooks        uint64
	BRooks        uint64
	WKnights      uint64
	BKnights      uint64
	WBishops      uint64
	BBishops      uint64
	WQueens       uint64
	BQueens       uint64
	WKings        uint64
	BKings        uint64
	LastMove      Move
	WCastle       CastleStatus
	BCastle       CastleStatus
	ToMove        Color
	Fullmove      int
	HalfmoveClock int
}

func (b *Board) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var eb exportedBoard
	eb.WPawns = b.wPawns
	eb.BPawns = b.bPawns
	eb.WRooks = b.wRooks
	eb.BRooks = b.bRooks
	eb.WKnights = b.wKnights
	eb.BKnights = b.bKnights
	eb.WBishops = b.wBishops
	eb.BBishops = b.bBishops
	eb.WQueens = b.wQueens
	eb.BQueens = b.bQueens
	eb.WKings = b.wKings
	eb.BKings = b.bKings
	eb.LastMove = b.lastMove
	eb.WCastle = b.wCastle
	eb.BCastle = b.bCastle
	eb.ToMove = b.toMove
	eb.Fullmove = b.fullmove
	eb.HalfmoveClock = b.halfmoveClock
	return gob.NewEncoder(file).Encode(eb)
}
