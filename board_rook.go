package pgn

func (b Board) findAttackingRook(pos Position, color Color, check bool) (Position, error) {
	//fmt.Println("find attacking rook:", pos, color, check)
	count := 0
	retPos := NoPosition

	r := pos.RankOrd()
	f := pos.FileOrd()
	for {
		f--
		testPos := PositionFromOrd(f, r)
		if b.checkRookColor(testPos, color) && (!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.RankOrd()
	f = pos.FileOrd()
	for {
		f++
		testPos := PositionFromOrd(f, r)
		//fmt.Println("find attacking rook f++:", pos, color, check, testPos)
		if b.checkRookColor(testPos, color) && (!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.RankOrd()
	f = pos.FileOrd()
	for {
		r++
		testPos := PositionFromOrd(f, r)
		//fmt.Println("find attacking rook r++:", pos, color, check, testPos)
		if b.checkRookColor(testPos, color) && (!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.RankOrd()
	f = pos.FileOrd()
	for {
		r--
		testPos := PositionFromOrd(f, r)
		//fmt.Println("find attacking rook r--:", pos, color, check, testPos)
		//fmt.Println("find attacking rook r--; checkRookColor:", b.checkRookColor(testPos, color))
		//fmt.Println("find attacking rook r--; moveintocheck:", !check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color))
		if b.checkRookColor(testPos, color) && (!check || !b.moveIntoCheck(Move{testPos, pos, NoPiece}, color)) {
			//	fmt.Println("wut r-- if")
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			//fmt.Println("wut r-- else if")
			break
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

func (b Board) findAttackingRookFromFile(pos Position, color Color, file Position) (Position, error) {
	count := 0
	retPos := NoPosition

	r := pos.RankOrd()
	f := pos.FileOrd()
	for {
		f--
		testPos := PositionFromOrd(f, r)
		if file.FileOrd() == f && b.checkRookColor(testPos, color) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.RankOrd()
	f = pos.FileOrd()
	for {
		f++
		testPos := PositionFromOrd(f, r)
		if file.FileOrd() == f && b.checkRookColor(testPos, color) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	if file.FileOrd() == pos.FileOrd() {
		r = pos.RankOrd()
		f = pos.FileOrd()
		for {
			r++
			testPos := PositionFromOrd(f, r)
			if b.checkRookColor(testPos, color) {
				retPos = testPos
				count++
				break
			} else if testPos == NoPosition || b.containsPieceAt(testPos) {
				break
			}
		}

		r = pos.RankOrd()
		f = pos.FileOrd()
		for {
			r--
			testPos := PositionFromOrd(f, r)
			if b.checkRookColor(testPos, color) {
				retPos = testPos
				count++
				break
			} else if testPos == NoPosition || b.containsPieceAt(testPos) {
				break
			}
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

func (b Board) findAttackingRookFromRank(pos Position, color Color, rank Position) (Position, error) {
	count := 0
	retPos := NoPosition

	r := pos.RankOrd()
	f := pos.FileOrd()
	for {
		r--
		testPos := PositionFromOrd(f, r)
		if rank.RankOrd() == r && b.checkRookColor(testPos, color) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	r = pos.RankOrd()
	f = pos.FileOrd()
	for {
		r++
		testPos := PositionFromOrd(f, r)
		if rank.RankOrd() == r && b.checkRookColor(testPos, color) {
			retPos = testPos
			count++
			break
		} else if testPos == NoPosition || b.containsPieceAt(testPos) {
			break
		}
	}

	if rank.RankOrd() == pos.RankOrd() {
		r := pos.RankOrd()
		f := pos.FileOrd()
		for {
			f--
			testPos := PositionFromOrd(f, r)
			if b.checkRookColor(testPos, color) {
				retPos = testPos
				count++
				break
			} else if testPos == NoPosition || b.containsPieceAt(testPos) {
				break
			}
		}

		r = pos.RankOrd()
		f = pos.FileOrd()
		for {
			f++
			testPos := PositionFromOrd(f, r)
			if b.checkRookColor(testPos, color) {
				retPos = testPos
				count++
				break
			} else if testPos == NoPosition || b.containsPieceAt(testPos) {
				break
			}
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

func (b Board) checkRookColor(pos Position, color Color) bool {
	return (b.GetPiece(pos) == WhiteRook && color == White) ||
		(b.GetPiece(pos) == BlackRook && color == Black)
}
