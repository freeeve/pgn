package pgn

import "fmt"

type Rank byte

const (
	NoRank Rank = '0' + iota
	Rank1
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

type File byte

const (
	FileA File = 'a' + iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
	NoFile File = ' '
)

type Position uint64

const (
	A1 Position = 1 << iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8

	NoPosition Position = 0
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
	if NoPosition == p {
		return "-"
	}
	return fmt.Sprintf("%c%c", p.GetFile(), p.GetRank())
}

func (p Position) GetRank() Rank {
	if uint64(p) <= uint64(NoPosition) {
		return NoRank
	} else if uint64(p) <= uint64(H1) {
		return Rank1
	} else if uint64(p) <= uint64(H2) {
		return Rank2
	} else if uint64(p) <= uint64(H3) {
		return Rank3
	} else if uint64(p) <= uint64(H4) {
		return Rank4
	} else if uint64(p) <= uint64(H5) {
		return Rank5
	} else if uint64(p) <= uint64(H6) {
		return Rank6
	} else if uint64(p) <= uint64(H7) {
		return Rank7
	} else {
		return Rank8
	}
}

func (p Position) GetFile() File {
	switch uint64(p) % 255 {
	case 1 << 7:
		return FileH
	case 1 << 6:
		return FileG
	case 1 << 5:
		return FileF
	case 1 << 4:
		return FileE
	case 1 << 3:
		return FileD
	case 1 << 2:
		return FileC
	case 1 << 1:
		return FileB
	case 1 << 0:
		return FileA
	}
	return NoFile
}

func PositionFromFileRank(f File, r Rank) (p Position) {
	pos, err := ParsePosition(fmt.Sprintf("%c%c", f, r))
	if err != nil {
		//fmt.Println(err)
		return NoPosition
	}
	return pos
}
