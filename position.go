package pgn

import "fmt"

type Rank byte

const (
	NoRank Rank = '0'
	Rank1  Rank = '1'
	Rank2  Rank = '2'
	Rank3  Rank = '3'
	Rank4  Rank = '4'
	Rank5  Rank = '5'
	Rank6  Rank = '6'
	Rank7  Rank = '7'
	Rank8  Rank = '8'
)

type File byte

const (
	NoFile File = ' '
	FileA  File = 'a'
	FileB  File = 'b'
	FileC  File = 'c'
	FileD  File = 'd'
	FileE  File = 'e'
	FileF  File = 'f'
	FileG  File = 'g'
	FileH  File = 'h'
)

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
