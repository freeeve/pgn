// scanner.go provides an in-memory PGN parser using the bitboard engine.
//
// PGNScanner reads games one at a time from a buffer, using GameState and
// the optimized ParseSANBytes for move parsing.

package pgn

import (
	"fmt"
	"io"
	"sync"
)

// Game represents a parsed PGN game with tags and moves.
type Game struct {
	Tags  map[string]string
	Moves []Mv
}

var tagMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16)
	},
}

// internTagName returns an interned string for common tag names to avoid allocation.
func internTagName(buf []byte) string {
	switch len(buf) {
	case 3:
		if buf[0] == 'E' && buf[1] == 'C' && buf[2] == 'O' {
			return "ECO"
		}
		if buf[0] == 'F' && buf[1] == 'E' && buf[2] == 'N' {
			return "FEN"
		}
	case 4:
		if buf[0] == 'S' && buf[1] == 'i' && buf[2] == 't' && buf[3] == 'e' {
			return "Site"
		}
		if buf[0] == 'D' && buf[1] == 'a' && buf[2] == 't' && buf[3] == 'e' {
			return "Date"
		}
	case 5:
		if buf[0] == 'E' && buf[1] == 'v' && buf[2] == 'e' && buf[3] == 'n' && buf[4] == 't' {
			return "Event"
		}
		if buf[0] == 'R' && buf[1] == 'o' && buf[2] == 'u' && buf[3] == 'n' && buf[4] == 'd' {
			return "Round"
		}
		if buf[0] == 'W' && buf[1] == 'h' && buf[2] == 'i' && buf[3] == 't' && buf[4] == 'e' {
			return "White"
		}
		if buf[0] == 'B' && buf[1] == 'l' && buf[2] == 'a' && buf[3] == 'c' && buf[4] == 'k' {
			return "Black"
		}
		if buf[0] == 'S' && buf[1] == 'e' && buf[2] == 't' && buf[3] == 'U' && buf[4] == 'p' {
			return "SetUp"
		}
	case 6:
		if buf[0] == 'R' && buf[1] == 'e' && buf[2] == 's' && buf[3] == 'u' && buf[4] == 'l' && buf[5] == 't' {
			return "Result"
		}
	case 7:
		if buf[0] == 'U' && buf[1] == 'T' && buf[2] == 'C' {
			if buf[3] == 'D' && buf[4] == 'a' && buf[5] == 't' && buf[6] == 'e' {
				return "UTCDate"
			}
			if buf[3] == 'T' && buf[4] == 'i' && buf[5] == 'm' && buf[6] == 'e' {
				return "UTCTime"
			}
		}
		if buf[0] == 'O' && buf[1] == 'p' && buf[2] == 'e' && buf[3] == 'n' && buf[4] == 'i' && buf[5] == 'n' && buf[6] == 'g' {
			return "Opening"
		}
		if buf[0] == 'V' && buf[1] == 'a' && buf[2] == 'r' && buf[3] == 'i' && buf[4] == 'a' && buf[5] == 'n' && buf[6] == 't' {
			return "Variant"
		}
	case 8:
		if buf[0] == 'W' && buf[1] == 'h' && buf[2] == 'i' && buf[3] == 't' && buf[4] == 'e' && buf[5] == 'E' && buf[6] == 'l' && buf[7] == 'o' {
			return "WhiteElo"
		}
		if buf[0] == 'B' && buf[1] == 'l' && buf[2] == 'a' && buf[3] == 'c' && buf[4] == 'k' && buf[5] == 'E' && buf[6] == 'l' && buf[7] == 'o' {
			return "BlackElo"
		}
		if buf[0] == 'P' && buf[1] == 'l' && buf[2] == 'y' && buf[3] == 'C' && buf[4] == 'o' && buf[5] == 'u' && buf[6] == 'n' && buf[7] == 't' {
			return "PlyCount"
		}
	case 9:
		if buf[0] == 'A' && buf[1] == 'n' && buf[2] == 'n' && buf[3] == 'o' && buf[4] == 't' && buf[5] == 'a' && buf[6] == 't' && buf[7] == 'o' && buf[8] == 'r' {
			return "Annotator"
		}
	case 11:
		if string(buf) == "TimeControl" {
			return "TimeControl"
		}
		if string(buf) == "Termination" {
			return "Termination"
		}
	}
	return string(buf)
}

// PGNScanner scans PGN games from a reader one at a time.
type PGNScanner struct {
	reader   io.Reader
	buf      []byte
	pos      int
	done     bool
	game     Game
	gs       *GameState
	startPos *GameState
	tagMap   map[string]string
}

// NewPGNScanner creates a new PGN scanner that reads from r.
func NewPGNScanner(r io.Reader) *PGNScanner {
	buf := make([]byte, 0, 1024*1024)
	chunk := make([]byte, 64*1024)
	for {
		n, err := r.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			break
		}
	}

	startPos, _ := NewGame("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	return &PGNScanner{
		buf:      buf,
		startPos: startPos,
		game:     Game{Moves: make([]Mv, 0, 100)},
		tagMap:   make(map[string]string, 16),
	}
}

// Next returns true if there are more games to scan.
func (ps *PGNScanner) Next() bool {
	return ps.pos < len(ps.buf)
}

// Scan parses and returns the next game, or an error if parsing fails.
func (ps *PGNScanner) Scan() (*Game, error) {
	for k := range ps.tagMap {
		delete(ps.tagMap, k)
	}
	ps.game.Tags = ps.tagMap
	ps.game.Moves = ps.game.Moves[:0]

	if err := ps.parseTags(); err != nil {
		return nil, err
	}

	if fen, ok := ps.game.Tags["FEN"]; ok {
		var err error
		ps.gs, err = NewGame(fen)
		if err != nil {
			return nil, err
		}
	} else {
		gs := *ps.startPos
		ps.gs = &gs
	}

	if err := ps.parseMoves(); err != nil {
		return nil, err
	}

	return &ps.game, nil
}

func (ps *PGNScanner) parseTags() error {
	for ps.pos < len(ps.buf) {
		ps.skipWhitespace()
		if ps.pos >= len(ps.buf) {
			break
		}

		c := ps.buf[ps.pos]
		if c == '[' {
			ps.pos++
			ps.skipWhitespace()

			start := ps.pos
			for ps.pos < len(ps.buf) && ps.buf[ps.pos] != ' ' && ps.buf[ps.pos] != '"' {
				ps.pos++
			}
			tagName := internTagName(ps.buf[start:ps.pos])

			ps.skipWhitespace()
			if ps.pos < len(ps.buf) && ps.buf[ps.pos] == '"' {
				ps.pos++
				start = ps.pos
				for ps.pos < len(ps.buf) && ps.buf[ps.pos] != '"' {
					ps.pos++
				}
				tagValue := string(ps.buf[start:ps.pos])
				ps.game.Tags[tagName] = tagValue
				ps.pos++
			}

			for ps.pos < len(ps.buf) && ps.buf[ps.pos] != '\n' && ps.buf[ps.pos] != '\r' {
				ps.pos++
			}
		} else if c >= '1' && c <= '9' {
			return nil
		} else if c == '\n' || c == '\r' {
			ps.pos++
		} else {
			return nil
		}
	}
	return nil
}

func (ps *PGNScanner) parseMoves() error {
	for ps.pos < len(ps.buf) {
		ps.skipWhitespace()
		if ps.pos >= len(ps.buf) {
			break
		}

		c := ps.buf[ps.pos]

		if c == '{' {
			for ps.pos < len(ps.buf) && ps.buf[ps.pos] != '}' {
				ps.pos++
			}
			if ps.pos < len(ps.buf) {
				ps.pos++
			}
			continue
		}

		if c == '(' {
			depth := 1
			ps.pos++
			for ps.pos < len(ps.buf) && depth > 0 {
				if ps.buf[ps.pos] == '(' {
					depth++
				} else if ps.buf[ps.pos] == ')' {
					depth--
				}
				ps.pos++
			}
			continue
		}

		if c == '[' {
			return nil
		}

		if c == '1' || c == '0' || c == '*' {
			token := ps.readToken()
			if isResult(token) {
				return nil
			}
			continue
		}

		if c == '$' || c == '!' || c == '?' {
			ps.readToken()
			continue
		}

		if (c >= 'a' && c <= 'h') || (c >= 'A' && c <= 'Z') || c == 'O' {
			start, end := ps.readMoveTokenBytes()
			if start == end {
				continue
			}

			mv, err := ParseSANBytes(ps.gs, ps.buf[start:end])
			if err != nil {
				return fmt.Errorf("move %d: %w", len(ps.game.Moves)+1, err)
			}
			ps.game.Moves = append(ps.game.Moves, mv)
			MakeMove(ps.gs, mv)
		} else {
			ps.pos++
		}
	}
	return nil
}

func (ps *PGNScanner) skipWhitespace() {
	for ps.pos < len(ps.buf) {
		c := ps.buf[ps.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '.' {
			ps.pos++
		} else {
			break
		}
	}
}

func (ps *PGNScanner) readToken() string {
	start := ps.pos
	for ps.pos < len(ps.buf) {
		c := ps.buf[ps.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '.' {
			break
		}
		ps.pos++
	}
	return string(ps.buf[start:ps.pos])
}

func (ps *PGNScanner) readMoveTokenBytes() (int, int) {
	start := ps.pos
	for ps.pos < len(ps.buf) {
		c := ps.buf[ps.pos]
		if (c >= 'a' && c <= 'h') || (c >= '1' && c <= '8') ||
			(c >= 'A' && c <= 'Z') || c == '-' || c == '=' ||
			c == '+' || c == '#' || c == 'x' {
			ps.pos++
		} else {
			break
		}
	}
	return start, ps.pos
}

func isResult(s string) bool {
	return s == "1-0" || s == "0-1" || s == "1/2-1/2" || s == "*"
}
