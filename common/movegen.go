package common

const (
	f1g1Mask = (uint64(1) << SquareF1) | (uint64(1) << SquareG1)
	b1d1Mask = (uint64(1) << SquareB1) | (uint64(1) << SquareC1) | (uint64(1) << SquareD1)
	f8g8Mask = (uint64(1) << SquareF8) | (uint64(1) << SquareG8)
	b8d8Mask = (uint64(1) << SquareB8) | (uint64(1) << SquareC8) | (uint64(1) << SquareD8)
)

var (
	whiteKingSideCastle  = makeMove(SquareE1, SquareG1, King, Empty)
	whiteQueenSideCastle = makeMove(SquareE1, SquareC1, King, Empty)
	blackKingSideCastle  = makeMove(SquareE8, SquareG8, King, Empty)
	blackQueenSideCastle = makeMove(SquareE8, SquareC8, King, Empty)
)

func (p *Position) GenerateLegalMoves() []Move {
	var result []Move
	var buffer [MaxMoves]OrderedMove
	var child Position
	var ml = p.GenerateMoves(buffer[:])
	for i := range ml {
		var m = ml[i].Move
		if p.MakeMove(m, &child) {
			result = append(result, m)
		}
	}
	return result
}

func (p *Position) GenerateMoves(ml []OrderedMove) []OrderedMove {
	var count = 0
	var fromBB, toBB, ownPieces, oppPieces uint64
	var from, to int

	if p.WhiteMove {
		ownPieces = p.White
		oppPieces = p.Black
	} else {
		ownPieces = p.Black
		oppPieces = p.White
	}

	var target = ^ownPieces
	if p.Checkers != 0 {
		var kingSq = FirstOne(p.Kings & ownPieces)
		target = p.Checkers | betweenMask[FirstOne(p.Checkers)][kingSq]
	}

	var allPieces = p.White | p.Black
	var ownPawns = p.Pawns & ownPieces

	if p.EpSquare != SquareNone {
		for fromBB = PawnAttacks(p.EpSquare, !p.WhiteMove) & ownPawns; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			ml[count].Move = makeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		for fromBB = p.Pawns & ownPieces & ^Rank7Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (SquareMask[from+8] & allPieces) == 0 {
				ml[count].Move = makeMove(from, from+8, Pawn, Empty)
				count++
				if Rank(from) == Rank2 && (SquareMask[from+16]&allPieces) == 0 {
					ml[count].Move = makeMove(from, from+16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (SquareMask[from+7]&oppPieces) != 0 {
				ml[count].Move = makeMove(from, from+7, Pawn, p.WhatPiece(from+7))
				count++
			}
			if File(from) < FileH && (SquareMask[from+9]&oppPieces) != 0 {
				ml[count].Move = makeMove(from, from+9, Pawn, p.WhatPiece(from+9))
				count++
			}
		}
		for fromBB = p.Pawns & ownPieces & Rank7Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (SquareMask[from+8] & allPieces) == 0 {
				count += addPromotions(ml[count:], makeMove(from, from+8, Pawn, Empty))
			}
			if File(from) > FileA && (SquareMask[from+7]&oppPieces) != 0 {
				count += addPromotions(ml[count:], makeMove(from, from+7, Pawn, p.WhatPiece(from+7)))
			}
			if File(from) < FileH && (SquareMask[from+9]&oppPieces) != 0 {
				count += addPromotions(ml[count:], makeMove(from, from+9, Pawn, p.WhatPiece(from+9)))
			}
		}
	} else {
		for fromBB = p.Pawns & ownPieces & ^Rank2Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (SquareMask[from-8] & allPieces) == 0 {
				ml[count].Move = makeMove(from, from-8, Pawn, Empty)
				count++
				if Rank(from) == Rank7 && (SquareMask[from-16]&allPieces) == 0 {
					ml[count].Move = makeMove(from, from-16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (SquareMask[from-9]&oppPieces) != 0 {
				ml[count].Move = makeMove(from, from-9, Pawn, p.WhatPiece(from-9))
				count++
			}
			if File(from) < FileH && (SquareMask[from-7]&oppPieces) != 0 {
				ml[count].Move = makeMove(from, from-7, Pawn, p.WhatPiece(from-7))
				count++
			}
		}
		for fromBB = p.Pawns & ownPieces & Rank2Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (SquareMask[from-8] & allPieces) == 0 {
				count += addPromotions(ml[count:], makeMove(from, from-8, Pawn, Empty))
			}
			if File(from) > FileA && (SquareMask[from-9]&oppPieces) != 0 {
				count += addPromotions(ml[count:], makeMove(from, from-9, Pawn, p.WhatPiece(from-9)))
			}
			if File(from) < FileH && (SquareMask[from-7]&oppPieces) != 0 {
				count += addPromotions(ml[count:], makeMove(from, from-7, Pawn, p.WhatPiece(from-7)))
			}
		}
	}

	for fromBB = p.Knights & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = KnightAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = KingAttacks[from] &^ ownPieces; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, King, p.WhatPiece(to))
			count++
		}

		if p.WhiteMove {
			if (p.CastleRights&WhiteKingSide) != 0 &&
				(allPieces&f1g1Mask) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareF1, false) {
				ml[count].Move = whiteKingSideCastle
				count++
			}
			if (p.CastleRights&WhiteQueenSide) != 0 &&
				(allPieces&b1d1Mask) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareD1, false) {
				ml[count].Move = whiteQueenSideCastle
				count++
			}
		} else {
			if (p.CastleRights&BlackKingSide) != 0 &&
				(allPieces&f8g8Mask) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareF8, true) {
				ml[count].Move = blackKingSideCastle
				count++
			}
			if (p.CastleRights&BlackQueenSide) != 0 &&
				(allPieces&b8d8Mask) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareD8, true) {
				ml[count].Move = blackQueenSideCastle
				count++
			}
		}
	}

	return ml[:count]
}

func (p *Position) GenerateCaptures(ml []OrderedMove) []OrderedMove {
	var count = 0
	var fromBB, toBB, ownPieces, oppPieces uint64
	var from, to, promotion int

	if p.WhiteMove {
		ownPieces = p.White
		oppPieces = p.Black
	} else {
		ownPieces = p.Black
		oppPieces = p.White
	}

	var target = oppPieces
	var allPieces = p.White | p.Black
	var ownPawns = p.Pawns & ownPieces

	if p.EpSquare != SquareNone {
		for fromBB = PawnAttacks(p.EpSquare, !p.WhiteMove) & ownPawns; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			ml[count].Move = makeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		fromBB = (AllBlackPawnAttacks(oppPieces) | Rank7Mask) & p.Pawns & p.White
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank7, Queen, Empty)
			if Rank(from) == Rank7 && (SquareMask[from+8]&allPieces) == 0 {
				ml[count].Move = makePawnMove(from, from+8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (SquareMask[from+7]&oppPieces) != 0 {
				ml[count].Move = makePawnMove(from, from+7, p.WhatPiece(from+7), promotion)
				count++
			}
			if File(from) < FileH && (SquareMask[from+9]&oppPieces) != 0 {
				ml[count].Move = makePawnMove(from, from+9, p.WhatPiece(from+9), promotion)
				count++
			}
		}
	} else {
		fromBB = (AllWhitePawnAttacks(oppPieces) | Rank2Mask) & p.Pawns & p.Black
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank2, Queen, Empty)
			if Rank(from) == Rank2 && (SquareMask[from-8]&allPieces) == 0 {
				ml[count].Move = makePawnMove(from, from-8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (SquareMask[from-9]&oppPieces) != 0 {
				ml[count].Move = makePawnMove(from, from-9, p.WhatPiece(from-9), promotion)
				count++
			}
			if File(from) < FileH && (SquareMask[from-7]&oppPieces) != 0 {
				ml[count].Move = makePawnMove(from, from-7, p.WhatPiece(from-7), promotion)
				count++
			}
		}
	}

	for fromBB = p.Knights & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = KnightAttacks[from] & (target); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & (target); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & (target); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & (target); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = KingAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count].Move = makeMove(from, to, King, p.WhatPiece(to))
			count++
		}
	}

	return ml[:count]
}

func addPromotions(ml []OrderedMove, move Move) (count int) {
	ml[0].Move = move ^ Move(Queen<<18)
	ml[1].Move = move ^ Move(Rook<<18)
	ml[2].Move = move ^ Move(Bishop<<18)
	ml[3].Move = move ^ Move(Knight<<18)
	return 4
}
