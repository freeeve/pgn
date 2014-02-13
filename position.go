package pgn

import "fmt"

type Position uint64

const NoPosition Position = 0

const (
	A1 Position = 1 << iota
	B1 Position = 1 << iota
	C1 Position = 1 << iota
	D1 Position = 1 << iota
	E1 Position = 1 << iota
	F1 Position = 1 << iota
	G1 Position = 1 << iota
	H1 Position = 1 << iota
	A2 Position = 1 << iota
	B2 Position = 1 << iota
	C2 Position = 1 << iota
	D2 Position = 1 << iota
	E2 Position = 1 << iota
	F2 Position = 1 << iota
	G2 Position = 1 << iota
	H2 Position = 1 << iota
	A3 Position = 1 << iota
	B3 Position = 1 << iota
	C3 Position = 1 << iota
	D3 Position = 1 << iota
	E3 Position = 1 << iota
	F3 Position = 1 << iota
	G3 Position = 1 << iota
	H3 Position = 1 << iota
	A4 Position = 1 << iota
	B4 Position = 1 << iota
	C4 Position = 1 << iota
	D4 Position = 1 << iota
	E4 Position = 1 << iota
	F4 Position = 1 << iota
	G4 Position = 1 << iota
	H4 Position = 1 << iota
	A5 Position = 1 << iota
	B5 Position = 1 << iota
	C5 Position = 1 << iota
	D5 Position = 1 << iota
	E5 Position = 1 << iota
	F5 Position = 1 << iota
	G5 Position = 1 << iota
	H5 Position = 1 << iota
	A6 Position = 1 << iota
	B6 Position = 1 << iota
	C6 Position = 1 << iota
	D6 Position = 1 << iota
	E6 Position = 1 << iota
	F6 Position = 1 << iota
	G6 Position = 1 << iota
	H6 Position = 1 << iota
	A7 Position = 1 << iota
	B7 Position = 1 << iota
	C7 Position = 1 << iota
	D7 Position = 1 << iota
	E7 Position = 1 << iota
	F7 Position = 1 << iota
	G7 Position = 1 << iota
	H7 Position = 1 << iota
	A8 Position = 1 << iota
	B8 Position = 1 << iota
	C8 Position = 1 << iota
	D8 Position = 1 << iota
	E8 Position = 1 << iota
	F8 Position = 1 << iota
	G8 Position = 1 << iota
	H8 Position = 1 << iota
)

func ParsePosition(pstr string) (Position, error) {
	switch pstr[0] {
	case 'A', 'a':
		switch pstr[1] {
		case '1':
			return A1, nil
		case '2':
			return A2, nil
		case '3':
			return A3, nil
		case '4':
			return A4, nil
		case '5':
			return A5, nil
		case '6':
			return A6, nil
		case '7':
			return A7, nil
		case '8':
			return A8, nil
		}
	case 'B', 'b':
		switch pstr[1] {
		case '1':
			return B1, nil
		case '2':
			return B2, nil
		case '3':
			return B3, nil
		case '4':
			return B4, nil
		case '5':
			return B5, nil
		case '6':
			return B6, nil
		case '7':
			return B7, nil
		case '8':
			return B8, nil
		}
	case 'C', 'c':
		switch pstr[1] {
		case '1':
			return C1, nil
		case '2':
			return C2, nil
		case '3':
			return C3, nil
		case '4':
			return C4, nil
		case '5':
			return C5, nil
		case '6':
			return C6, nil
		case '7':
			return C7, nil
		case '8':
			return C8, nil
		}
	case 'D', 'd':
		switch pstr[1] {
		case '1':
			return D1, nil
		case '2':
			return D2, nil
		case '3':
			return D3, nil
		case '4':
			return D4, nil
		case '5':
			return D5, nil
		case '6':
			return D6, nil
		case '7':
			return D7, nil
		case '8':
			return D8, nil
		}
	case 'E', 'e':
		switch pstr[1] {
		case '1':
			return E1, nil
		case '2':
			return E2, nil
		case '3':
			return E3, nil
		case '4':
			return E4, nil
		case '5':
			return E5, nil
		case '6':
			return E6, nil
		case '7':
			return E7, nil
		case '8':
			return E8, nil
		}
	case 'F', 'f':
		switch pstr[1] {
		case '1':
			return F1, nil
		case '2':
			return F2, nil
		case '3':
			return F3, nil
		case '4':
			return F4, nil
		case '5':
			return F5, nil
		case '6':
			return F6, nil
		case '7':
			return F7, nil
		case '8':
			return F8, nil
		}
	case 'G', 'g':
		switch pstr[1] {
		case '1':
			return G1, nil
		case '2':
			return G2, nil
		case '3':
			return G3, nil
		case '4':
			return G4, nil
		case '5':
			return G5, nil
		case '6':
			return G6, nil
		case '7':
			return G7, nil
		case '8':
			return G8, nil
		}
	case 'H', 'h':
		switch pstr[1] {
		case '1':
			return H1, nil
		case '2':
			return H2, nil
		case '3':
			return H3, nil
		case '4':
			return H4, nil
		case '5':
			return H5, nil
		case '6':
			return H6, nil
		case '7':
			return H7, nil
		case '8':
			return H8, nil
		}
	}
	return NoPosition, fmt.Errorf("pgn: invalid position string: %s", pstr)
}

func (p Position) String() string {
	switch p {
	case A1:
		return "a1"
	case A2:
		return "a2"
	case A3:
		return "a3"
	case A4:
		return "a4"
	case A5:
		return "a5"
	case A6:
		return "a6"
	case A7:
		return "a7"
	case A8:
		return "a8"
	case B1:
		return "b1"
	case B2:
		return "b2"
	case B3:
		return "b3"
	case B4:
		return "b4"
	case B5:
		return "b5"
	case B6:
		return "b6"
	case B7:
		return "b7"
	case B8:
		return "a8"
	case C1:
		return "c1"
	case C2:
		return "c2"
	case C3:
		return "c3"
	case C4:
		return "c4"
	case C5:
		return "c5"
	case C6:
		return "c6"
	case C7:
		return "c7"
	case C8:
		return "c8"
	case D1:
		return "d1"
	case D2:
		return "d2"
	case D3:
		return "d3"
	case D4:
		return "d4"
	case D5:
		return "d5"
	case D6:
		return "d6"
	case D7:
		return "d7"
	case D8:
		return "d8"
	case E1:
		return "e1"
	case E2:
		return "e2"
	case E3:
		return "e3"
	case E4:
		return "e4"
	case E5:
		return "e5"
	case E6:
		return "e6"
	case E7:
		return "e7"
	case E8:
		return "e8"
	case F1:
		return "f1"
	case F2:
		return "f2"
	case F3:
		return "f3"
	case F4:
		return "f4"
	case F5:
		return "f5"
	case F6:
		return "f6"
	case F7:
		return "f7"
	case F8:
		return "f8"
	case G1:
		return "g1"
	case G2:
		return "g2"
	case G3:
		return "g3"
	case G4:
		return "g4"
	case G5:
		return "g5"
	case G6:
		return "g6"
	case G7:
		return "g7"
	case G8:
		return "g8"
	case H1:
		return "h1"
	case H2:
		return "h2"
	case H3:
		return "h3"
	case H4:
		return "h4"
	case H5:
		return "h5"
	case H6:
		return "h6"
	case H7:
		return "h7"
	case H8:
		return "h8"
	default:
		return "-"
	}
}
