package pgn

// import "gopkg.in/wfreeman/pgn.v2"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type PGNScanner struct {
	ch        chan token
	tokenizer *PGNTokenizer
}

type PGNTokenizer struct {
	ch       chan token
	r        io.Reader
	done     bool
	onlyTags bool
}

type Game struct {
	Moves []Move
	Tags  map[string]string
}

type Move struct {
	From    Position
	To      Position
	Promote Piece
}

func (m Move) String() string {
	if m.Promote == NoPiece {
		return string([]byte{byte(m.From), byte(m.To)})
	}
	return string([]byte{byte(m.From), byte(m.To), byte(m.Promote)})
}

var (
	NilMove Move = Move{From: NoPosition, To: NoPosition}
)

type tokenType uint8

const (
	tagToken tokenType = iota
	moveNumberToken
	moveToken
	commentToken
	resultToken
)

type token struct {
	tokenType
	token string
}

func (tokenizer *PGNTokenizer) tokenize() error {
	br := bufio.NewReader(tokenizer.r)
	var buf bytes.Buffer
	inComment := false
	inTag := false
	for {
		next, _, err := br.ReadRune()
		if err != nil {
			tokenizer.done = true
			close(tokenizer.ch)
			return err
		}
		if inComment {
			if next == '}' {
				tokenizer.ch <- token{commentToken, buf.String()}
				inComment = false
				buf.Reset()
			} else if next == '\n' || next == '\r' {
				// ignore
			} else {
				buf.WriteRune(next)
			}
		} else if inTag {
			switch next {
			case ']':
				tokenizer.ch <- token{tagToken, buf.String()}
				buf.Reset()
				inTag = false
			default:
				buf.WriteRune(next)
			}
		} else {
			switch next {
			case '.':
				buf.WriteRune(next)
				if !tokenizer.onlyTags {
					tokenizer.ch <- token{moveNumberToken, buf.String()}
				}
				buf.Reset()
			case '[':
				inTag = true
			case ' ', '\t', '\n', '\r':
				if strings.TrimSpace(buf.String()) != "" {
					if !tokenizer.onlyTags || (tokenizer.onlyTags && inTag) {
						switch strings.TrimSpace(buf.String()) {
						case "1-0", "0-1", "*", "1/2-1/2":
							tokenizer.ch <- token{resultToken, buf.String()}
						default:
							tokenizer.ch <- token{moveToken, buf.String()}
						}
					}
				}
				buf.Reset()
			case '{':
				inComment = true
				buf.Reset()
			default:
				buf.WriteRune(next)
			}
		}
	}
	close(tokenizer.ch)
	return nil
}

func NewPGNTagScanner(r io.Reader) *PGNScanner {
	return newPGNScanner(r, true)
}

func NewPGNScanner(r io.Reader) *PGNScanner {
	return newPGNScanner(r, false)
}

func newPGNScanner(r io.Reader, onlyTags bool) *PGNScanner {
	tokens := make(chan token, 128)
	tokenizer := &PGNTokenizer{r: r, onlyTags: onlyTags, ch: tokens}
	go tokenizer.tokenize()
	return &PGNScanner{tokenizer: tokenizer, ch: tokens}
}

func (ps *PGNScanner) Next() bool {
	if ps.tokenizer.done {
		return false
	}
	return true
}

func (ps *PGNScanner) ParseGame() (*Game, error) {
	g := Game{Tags: map[string]string{}, Moves: []Move{}}
	var board *Board
	var err error
	var next = White
	for v := range ps.ch {
		switch v.tokenType {
		case tagToken:
			g.addTag(v.token)
		case moveNumberToken:
			if len(g.Moves) == 0 {
				if len(g.Tags["FEN"]) > 0 {
					board, err = NewBoardFEN(g.Tags["FEN"])
					if err != nil {
						return nil, err
					}
				} else {
					board = NewBoard()
				}
			}
		case moveToken:
			move, err := board.MoveFromAlgebraic(v.token, next)
			if err != nil {
				fmt.Println(board, "[", v.token, "]")
				return nil, err
			}
			g.Moves = append(g.Moves, move)
			board.MakeMove(move)
			if next == White {
				next = Black
			} else {
				next = White
			}
		case commentToken:
			// fmt.Println("this is a comment:", v)
		case resultToken:
			return &g, nil
		}
	}
	return &g, nil
}

func (g *Game) addTag(tag string) error {
	firstSpace := strings.Index(tag, " ")
	if firstSpace == -1 {
		return ErrUnparseableTag
	}
	g.Tags[tag[:firstSpace]] = strings.Trim(tag[firstSpace:], " \"")
	return nil
}

func (ps *PGNScanner) Scan() (*Game, error) {
	game, err := ps.ParseGame()
	return game, err
}
