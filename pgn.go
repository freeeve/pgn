package pgn

import (
	"strings"
	"text/scanner"
)

type Game struct {
	Moves []Move
	Tags  map[string]string
}

type Move struct {
	From Position
	To   Position
}

func Parse(str string) (*Game, error) {
	g := Game{Tags: map[string]string{}, Moves: []Move{}}
	r := strings.NewReader(str)
	s := scanner.Scanner{}
	s.Init(r)
	err := ParseTags(&s, &g)
	if err != nil {
		return nil, err
	}
	err = ParseMoves(&s, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func ParseTags(s *scanner.Scanner, g *Game) error {
	run := s.Peek()
	for run != scanner.EOF {
		switch run {
		case '[':
			run = s.Next()
		case ']':
			run = s.Next()
		case '\n':
			run = s.Next()
		case '1':
			return nil
		default:
			s.Scan()
			tag := s.TokenText()
			s.Scan()
			val := s.TokenText()
			g.Tags[tag] = strings.Trim(val, "\"")
		}
		run = s.Peek()
	}
	return nil
}

func isEnd(str string) bool {
	if str == "1/2-1/2" {
		return true
	}
	if str == "0-1" {
		return true
	}
	if str == "1-0" {
		return true
	}
	return false
}

func ParseMoves(s *scanner.Scanner, g *Game) error {
	run := s.Peek()
	board := &Board{}
	var err error
	if len(g.Tags["FEN"]) > 0 {
		board, err = NewBoardFEN(g.Tags["FEN"])
		if err != nil {
			return err
		}
	}
	num := ""
	white := ""
	black := ""
	for run != scanner.EOF {
		switch run {
		case '(':
			for run != ')' && run != scanner.EOF {
				run = s.Next()
			}
		case '{':
			for run != '}' && run != scanner.EOF {
				run = s.Next()
			}
		default:
			s.Scan()
			if num == "" {
				num = s.TokenText()
				if isEnd(num) {
					return nil
				}
			} else if white == "" {
				white = s.TokenText()
				if isEnd(white) {
					return nil
				}
				move, err := board.MoveFromAlgebraic(white, White)
				if err != nil {
					return err
				}
				g.Moves = append(g.Moves, *move)
			} else if black == "" {
				black = s.TokenText()
				if isEnd(black) {
					return nil
				}
				move, err := board.MoveFromAlgebraic(black, Black)
				if err != nil {
					return err
				}
				g.Moves = append(g.Moves, *move)
				num = ""
				white = ""
				black = ""
			}
			run = s.Peek()
		}
	}
	return nil
}
