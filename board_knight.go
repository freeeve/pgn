package pgn

type knight struct{}

func (knight) mask(sq Position) (mask Position) {
	// sq is a single board square (one set bit). mask will be a set of
	// bits representing the possible locations from which a knight could
	// reach sq (sq can also represent the starting point for a knight, in
	// which case the mask is all the reachable squares).

	if sq.IsSquare() {
		return NoPosition
	}

	// mask of possible knight moves, not yet adjusted for input square.
	mask = 0x0a1100110a

	// offset to the "center" of the mask.
	// XXX: might be 18 actually.
	const offset = 19

	const board = ^Position(0) // all bits set

	f := sq.FileOrd()
	// zero out up to 2 out of 4 files from the mask based on the input sq.
	// this will avoid wrap-around problems when we shift later
	switch f {
	case 0:
		mask &= board ^ FileA ^ FileB
	case 1:
		mask &= board ^ FileA
	case 6:
		mask &= board ^ FileD
	case 7:
		mask &= board ^ FileD ^ FileE
	}

	// ord is the square ordinal (0-63)
	ord := uint(sq.RankOrd()*8 + f)

	// use bit shifting to move the mask into place. we don't have to worry
	// about proximity to the top or the bottom of the board, since those
	// bits won't wrap-around (they'll just "fall off" the end).

	if ord < offset {
		return mask >> (offset - ord)
	}
	return mask << (ord - offset)
}

func (b Board) findAttackingKnight(pos Position, color Color, check bool) (Position, error) {
	count := 0
	r := pos.RankOrd()
	f := pos.FileOrd()
	retPos := NoPosition

	//fmt.Println("pos", pos)
	testPos := PositionFromOrd(f+1, r+2)
	//fmt.Println("testPos-1", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f+1, r-2)
	//fmt.Println("testPos0", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f+2, r+1)
	//fmt.Println("testPos1", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f+2, r-1)
	//fmt.Println("testPos2", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f-2, r-1)
	//fmt.Println("testPos3", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f-2, r+1)
	//fmt.Println("testPos4", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f-1, r-2)
	//fmt.Println("testPos5", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	testPos = PositionFromOrd(f-1, r+2)
	//fmt.Println("testPos6", testPos)
	if testPos != NoPosition &&
		b.checkKnightColor(testPos, color) &&
		(!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
		count++
		retPos = testPos
	}

	if count > 1 {
		return NoPosition, ErrAmbiguousMove
	}
	if count == 0 {
		return NoPosition, ErrAttackerNotFound
	}
	return retPos, nil
}

func (b Board) findAttackingKnightFromFile(pos Position, color Color, file Position) (Position, error) {
	//fmt.Println("finding attacking knight from file:", pos, color, file)
	count := 0
	r := pos.RankOrd()
	f := pos.FileOrd()
	retPos := NoPosition

	if f+1 == file.FileOrd() {
		testPos := PositionFromOrd(f+1, r+2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f+1, r-2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if f+2 == file.FileOrd() {
		testPos := PositionFromOrd(f+2, r+1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f+2, r-1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if f-2 == file.FileOrd() {
		testPos := PositionFromOrd(f-2, r-1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f-2, r+1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if f-1 == file.FileOrd() {
		testPos := PositionFromOrd(f-1, r-2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f-1, r+2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if count > 1 {
		return NoPosition, ErrAmbiguousMove
	}
	if count == 0 {
		return NoPosition, ErrAttackerNotFound
	}
	return retPos, nil
}

func (b Board) findAttackingKnightFromRank(pos Position, color Color, rank Position) (Position, error) {
	//fmt.Println("finding attacking knight from rank:", rank)
	count := 0
	r := pos.RankOrd()
	f := pos.FileOrd()
	retPos := NoPosition

	if r+2 == rank.RankOrd() {
		testPos := PositionFromOrd(f+1, r+2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f-1, r+2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if r+1 == rank.RankOrd() {
		testPos := PositionFromOrd(f+2, r+1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f-2, r+1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if r-1 == rank.RankOrd() {
		testPos := PositionFromOrd(f-2, r-1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f+2, r-1)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if r-2 == rank.RankOrd() {
		testPos := PositionFromOrd(f-1, r-2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}

		testPos = PositionFromOrd(f+1, r-2)
		if testPos != NoPosition && b.checkKnightColor(testPos, color) {
			count++
			retPos = testPos
		}
	}

	if count > 1 {
		return NoPosition, ErrAmbiguousMove
	}
	if count == 0 {
		return NoPosition, ErrAttackerNotFound
	}
	return retPos, nil
}

func (b Board) checkKnightColor(pos Position, color Color) bool {
	return (b.GetPiece(pos) == WhiteKnight && color == White) ||
		(b.GetPiece(pos) == BlackKnight && color == Black)
}
