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

func addPromotions(ml []Move, move Move) (count int) {
	ml[0] = move ^ Move(Queen<<18)
	ml[1] = move ^ Move(Rook<<18)
	ml[2] = move ^ Move(Bishop<<18)
	ml[3] = move ^ Move(Knight<<18)
	return 4
}

func GenerateMoves(ml []Move, p *Position) []Move {
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
			ml[count] = makeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		for fromBB = p.Pawns & ownPieces & ^Rank7Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (SquareMask[from+8] & allPieces) == 0 {
				ml[count] = makeMove(from, from+8, Pawn, Empty)
				count++
				if Rank(from) == Rank2 && (SquareMask[from+16]&allPieces) == 0 {
					ml[count] = makeMove(from, from+16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (SquareMask[from+7]&oppPieces) != 0 {
				ml[count] = makeMove(from, from+7, Pawn, p.WhatPiece(from+7))
				count++
			}
			if File(from) < FileH && (SquareMask[from+9]&oppPieces) != 0 {
				ml[count] = makeMove(from, from+9, Pawn, p.WhatPiece(from+9))
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
				ml[count] = makeMove(from, from-8, Pawn, Empty)
				count++
				if Rank(from) == Rank7 && (SquareMask[from-16]&allPieces) == 0 {
					ml[count] = makeMove(from, from-16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (SquareMask[from-9]&oppPieces) != 0 {
				ml[count] = makeMove(from, from-9, Pawn, p.WhatPiece(from-9))
				count++
			}
			if File(from) < FileH && (SquareMask[from-7]&oppPieces) != 0 {
				ml[count] = makeMove(from, from-7, Pawn, p.WhatPiece(from-7))
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
			ml[count] = makeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = KingAttacks[from] &^ ownPieces; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, King, p.WhatPiece(to))
			count++
		}

		if p.WhiteMove {
			if (p.CastleRights&WhiteKingSide) != 0 &&
				(allPieces&f1g1Mask) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareF1, false) {
				ml[count] = whiteKingSideCastle
				count++
			}
			if (p.CastleRights&WhiteQueenSide) != 0 &&
				(allPieces&b1d1Mask) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareD1, false) {
				ml[count] = whiteQueenSideCastle
				count++
			}
		} else {
			if (p.CastleRights&BlackKingSide) != 0 &&
				(allPieces&f8g8Mask) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareF8, true) {
				ml[count] = blackKingSideCastle
				count++
			}
			if (p.CastleRights&BlackQueenSide) != 0 &&
				(allPieces&b8d8Mask) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareD8, true) {
				ml[count] = blackQueenSideCastle
				count++
			}
		}
	}

	return ml[:count]
}

func GenerateCaptures(ml []Move, p *Position, genChecks bool) []Move {
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
			ml[count] = makeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		fromBB = (AllBlackPawnAttacks(oppPieces) | Rank7Mask) & p.Pawns & p.White
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank7, Queen, Empty)
			if Rank(from) == Rank7 && (SquareMask[from+8]&allPieces) == 0 {
				ml[count] = makePawnMove(from, from+8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (SquareMask[from+7]&oppPieces) != 0 {
				ml[count] = makePawnMove(from, from+7, p.WhatPiece(from+7), promotion)
				count++
			}
			if File(from) < FileH && (SquareMask[from+9]&oppPieces) != 0 {
				ml[count] = makePawnMove(from, from+9, p.WhatPiece(from+9), promotion)
				count++
			}
		}
		if genChecks {
			var oppKing = FirstOne(p.Kings & oppPieces)

			if (((p.Pawns&p.White & ^FileHMask)<<17)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing-9]&allPieces) == 0 {
				ml[count] = makeMove(oppKing-17, oppKing-9, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.White&Rank2Mask & ^FileHMask)<<25)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing-9]&allPieces) == 0 &&
				(SquareMask[oppKing-17]&allPieces) == 0 {
				ml[count] = makeMove(oppKing-25, oppKing-9, Pawn, Empty)
				count++
			}

			if (((p.Pawns&p.White & ^FileAMask)<<15)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing-7]&allPieces) == 0 {
				ml[count] = makeMove(oppKing-15, oppKing-7, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.White&Rank2Mask & ^FileAMask)<<23)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing-7]&allPieces) == 0 &&
				(SquareMask[oppKing-15]&allPieces) == 0 {
				ml[count] = makeMove(oppKing-23, oppKing-7, Pawn, Empty)
				count++
			}
		}
	} else {
		fromBB = (AllWhitePawnAttacks(oppPieces) | Rank2Mask) & p.Pawns & p.Black
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank2, Queen, Empty)
			if Rank(from) == Rank2 && (SquareMask[from-8]&allPieces) == 0 {
				ml[count] = makePawnMove(from, from-8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (SquareMask[from-9]&oppPieces) != 0 {
				ml[count] = makePawnMove(from, from-9, p.WhatPiece(from-9), promotion)
				count++
			}
			if File(from) < FileH && (SquareMask[from-7]&oppPieces) != 0 {
				ml[count] = makePawnMove(from, from-7, p.WhatPiece(from-7), promotion)
				count++
			}
		}
		if genChecks {
			var oppKing = FirstOne(p.Kings & oppPieces)

			if (((p.Pawns&p.Black & ^FileHMask)>>15)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing+7]&allPieces) == 0 {
				ml[count] = makeMove(oppKing+15, oppKing+7, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.Black&Rank7Mask & ^FileHMask)>>23)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing+7]&allPieces) == 0 &&
				(SquareMask[oppKing+15]&allPieces) == 0 {
				ml[count] = makeMove(oppKing+23, oppKing+7, Pawn, Empty)
				count++
			}

			if (((p.Pawns&p.Black & ^FileAMask)>>17)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing+9]&allPieces) == 0 {
				ml[count] = makeMove(oppKing+17, oppKing+9, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.Black&Rank7Mask & ^FileAMask)>>25)&p.Kings&oppPieces) != 0 &&
				(SquareMask[oppKing+9]&allPieces) == 0 &&
				(SquareMask[oppKing+17]&allPieces) == 0 {
				ml[count] = makeMove(oppKing+25, oppKing+9, Pawn, Empty)
				count++
			}
		}
	}

	var checksN, checksB, checksR, checksQ uint64
	if genChecks {
		var oppKing = FirstOne(p.Kings & oppPieces)
		checksN = KnightAttacks[oppKing] &^ allPieces
		checksB = BishopAttacks(oppKing, allPieces) &^ allPieces
		checksR = RookAttacks(oppKing, allPieces) &^ allPieces
		checksQ = checksB | checksR

		//discovered checks
		//TODO pawn, king discovered checks
		for fromBB = (p.Rooks | p.Queens) & ownPieces & rookMoves[oppKing]; fromBB != 0; fromBB &= fromBB - 1 {
			var blockers = betweenMask[FirstOne(fromBB)][oppKing] & allPieces
			if blockers&(blockers-1) == 0 {
				from = FirstOne(blockers)
				if (SquareMask[from] & ownPieces) != 0 {
					var piece = p.WhatPiece(from)
					if piece == Knight {
						for toBB = KnightAttacks[from] & ^allPieces & ^checksN; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml[count] = makeMove(from, to, Knight, p.WhatPiece(to))
							count++
						}
					} else if piece == Bishop {
						for toBB = BishopAttacks(from, allPieces) & ^allPieces & ^checksB; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml[count] = makeMove(from, to, Bishop, p.WhatPiece(to))
							count++
						}
					}
				}
			}
		}

		for fromBB = (p.Bishops | p.Queens) & ownPieces & bishopMoves[oppKing]; fromBB != 0; fromBB &= fromBB - 1 {
			var blockers = betweenMask[FirstOne(fromBB)][oppKing] & allPieces
			if blockers&(blockers-1) == 0 {
				from = FirstOne(blockers)
				if (SquareMask[from] & ownPieces) != 0 {
					var piece = p.WhatPiece(from)
					if piece == Knight {
						for toBB = KnightAttacks[from] & ^allPieces & ^checksN; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml[count] = makeMove(from, to, Knight, p.WhatPiece(to))
							count++
						}
					} else if piece == Rook {
						for toBB = RookAttacks(from, allPieces) & ^allPieces & ^checksR; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml[count] = makeMove(from, to, Rook, p.WhatPiece(to))
							count++
						}
					} else if piece == Pawn {
						if p.WhiteMove {
							if (allPieces&SquareMask[from+8]) == 0 &&
								Rank(from) != Rank7 &&
								(SquareMask[from+8]&PawnAttacks(oppKing, false)) == 0 {
								ml[count] = makeMove(from, from+8, Pawn, Empty)
								count++
							}
						} else {
							if (allPieces&SquareMask[from-8]) == 0 &&
								Rank(from) != Rank2 &&
								(SquareMask[from-8]&PawnAttacks(oppKing, true)) == 0 {
								ml[count] = makeMove(from, from-8, Pawn, Empty)
								count++
							}
						}
					}
				}
			}
		}
	}

	for fromBB = p.Knights & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = KnightAttacks[from] & (target | checksN); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & (target | checksB); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & (target | checksR); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & (target | checksQ); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = KingAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml[count] = makeMove(from, to, King, p.WhatPiece(to))
			count++
		}
	}

	return ml[:count]
}

func GenerateLegalMoves(pos *Position) (ml []Move) {
	var buffer [MaxMoves]Move
	var child Position
	for _, m := range GenerateMoves(buffer[:], pos) {
		if pos.MakeMove(m, &child) {
			ml = append(ml, m)
		}
	}
	return ml
}
