package pgn

import (
	"testing"
	. "launchpad.net/gocheck"
)

func Test(t *testing.T) { TestingT(t) }

type PGNSuite struct{}

var _ = Suite(&PGNSuite{})

var simple = `[Event "State Ch."]
[Site "New York, USA"]
[Date "1910.??.??"]
[Round "?"]
[White "Capablanca"]
[Black "Jaffe"]
[Result "1-0"]
[ECO "D46"]
[Opening "Queen's Gambit Dec."]
[Annotator "Reinfeld, Fred"]
[WhiteTitle "GM"]
[WhiteCountry "Cuba"]
[BlackCountry "United States"]

1. d4 d5 2. Nf3 Nf6 3. e3 c6 4. c4 e6 5. Nc3 Nbd7 6. Bd3 Bd6
7. O-O O-O 8. e4 dxe4 9. Nxe4 Nxe4 10. Bxe4 Nf6 11. Bc2 h6
12. b3 b6 13. Bb2 Bb7 14. Qd3 g6 15. Rae1 Nh5 16. Bc1 Kg7
17. Rxe6 Nf6 18. Ne5 c5 19. Bxh6+ Kxh6 20. Nxf7+ 1-0
`

func (s *PGNSuite) TestParse(c *C) {
	c.Skip("temporarly skipping until board is further along")
	game, err := Parse(simple)
	if err != nil {
		c.Fatal(err)
	}
	if game.Tags["Site"] != "New York, USA" {
		c.Fatal("Site tag wrong: ", game.Tags["Site"])
	}
	if len(game.Moves) == 0 || game.Moves[0].From != "d1" || game.Moves[0].To != "d4" {
		c.Fatal("first move is wrong", game.Moves[0])
	}
	if len(game.Moves) != 41 || game.Moves[0].From != "e5" || game.Moves[0].To != "f7" {
		c.Fatal("last move is wrong", game.Moves[0])
	}
}
