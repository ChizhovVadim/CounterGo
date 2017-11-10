package engine

import (
	"sort"
)

const (
	F1G1 = (uint64(1) << SquareF1) | (uint64(1) << SquareG1)
	B1D1 = (uint64(1) << SquareB1) | (uint64(1) << SquareC1) | (uint64(1) << SquareD1)
	F8G8 = (uint64(1) << SquareF8) | (uint64(1) << SquareG8)
	B8D8 = (uint64(1) << SquareB8) | (uint64(1) << SquareC8) | (uint64(1) << SquareD8)
)

var (
	whiteKingSideCastle  = MakeMove(SquareE1, SquareG1, King, Empty)
	whiteQueenSideCastle = MakeMove(SquareE1, SquareC1, King, Empty)
	blackKingSideCastle  = MakeMove(SquareE8, SquareG8, King, Empty)
	blackQueenSideCastle = MakeMove(SquareE8, SquareC8, King, Empty)
)

func (ml *MoveList) GenerateMoves(p *Position) {
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
			ml.Items[count].Move = MakeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		for fromBB = p.Pawns & ownPieces & ^Rank7Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (squareMask[from+8] & allPieces) == 0 {
				ml.Items[count].Move = MakeMove(from, from+8, Pawn, Empty)
				count++
				if Rank(from) == Rank2 && (squareMask[from+16]&allPieces) == 0 {
					ml.Items[count].Move = MakeMove(from, from+16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (squareMask[from+7]&oppPieces) != 0 {
				ml.Items[count].Move = MakeMove(from, from+7, Pawn, p.WhatPiece(from+7))
				count++
			}
			if File(from) < FileH && (squareMask[from+9]&oppPieces) != 0 {
				ml.Items[count].Move = MakeMove(from, from+9, Pawn, p.WhatPiece(from+9))
				count++
			}
		}
		for fromBB = p.Pawns & ownPieces & Rank7Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (squareMask[from+8] & allPieces) == 0 {
				count = ml.AddPromotions(MakeMove(from, from+8, Pawn, Empty), count)
			}
			if File(from) > FileA && (squareMask[from+7]&oppPieces) != 0 {
				count = ml.AddPromotions(MakeMove(from, from+7, Pawn, p.WhatPiece(from+7)), count)
			}
			if File(from) < FileH && (squareMask[from+9]&oppPieces) != 0 {
				count = ml.AddPromotions(MakeMove(from, from+9, Pawn, p.WhatPiece(from+9)), count)
			}
		}
	} else {
		for fromBB = p.Pawns & ownPieces & ^Rank2Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (squareMask[from-8] & allPieces) == 0 {
				ml.Items[count].Move = MakeMove(from, from-8, Pawn, Empty)
				count++
				if Rank(from) == Rank7 && (squareMask[from-16]&allPieces) == 0 {
					ml.Items[count].Move = MakeMove(from, from-16, Pawn, Empty)
					count++
				}
			}
			if File(from) > FileA && (squareMask[from-9]&oppPieces) != 0 {
				ml.Items[count].Move = MakeMove(from, from-9, Pawn, p.WhatPiece(from-9))
				count++
			}
			if File(from) < FileH && (squareMask[from-7]&oppPieces) != 0 {
				ml.Items[count].Move = MakeMove(from, from-7, Pawn, p.WhatPiece(from-7))
				count++
			}
		}
		for fromBB = p.Pawns & ownPieces & Rank2Mask; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			if (squareMask[from-8] & allPieces) == 0 {
				count = ml.AddPromotions(MakeMove(from, from-8, Pawn, Empty), count)
			}
			if File(from) > FileA && (squareMask[from-9]&oppPieces) != 0 {
				count = ml.AddPromotions(MakeMove(from, from-9, Pawn, p.WhatPiece(from-9)), count)
			}
			if File(from) < FileH && (squareMask[from-7]&oppPieces) != 0 {
				count = ml.AddPromotions(MakeMove(from, from-7, Pawn, p.WhatPiece(from-7)), count)
			}
		}
	}

	for fromBB = p.Knights & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = knightAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = kingAttacks[from] &^ ownPieces; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, King, p.WhatPiece(to))
			count++
		}

		if p.WhiteMove {
			if (p.CastleRights&WhiteKingSide) != 0 &&
				(allPieces&F1G1) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareF1, false) {
				ml.Items[count].Move = whiteKingSideCastle
				count++
			}
			if (p.CastleRights&WhiteQueenSide) != 0 &&
				(allPieces&B1D1) == 0 &&
				!p.isAttackedBySide(SquareE1, false) &&
				!p.isAttackedBySide(SquareD1, false) {
				ml.Items[count].Move = whiteQueenSideCastle
				count++
			}
		} else {
			if (p.CastleRights&BlackKingSide) != 0 &&
				(allPieces&F8G8) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareF8, true) {
				ml.Items[count].Move = blackKingSideCastle
				count++
			}
			if (p.CastleRights&BlackQueenSide) != 0 &&
				(allPieces&B8D8) == 0 &&
				!p.isAttackedBySide(SquareE8, true) &&
				!p.isAttackedBySide(SquareD8, true) {
				ml.Items[count].Move = blackQueenSideCastle
				count++
			}
		}
	}

	ml.Count = count
}

func (ml *MoveList) GenerateCaptures(p *Position, genChecks bool) {
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
			ml.Items[count].Move = MakeMove(from, p.EpSquare, Pawn, Pawn)
			count++
		}
	}

	if p.WhiteMove {
		fromBB = (AllBlackPawnAttacks(oppPieces) | Rank7Mask) & p.Pawns & p.White
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank7, Queen, Empty)
			if Rank(from) == Rank7 && (squareMask[from+8]&allPieces) == 0 {
				ml.Items[count].Move = MakePawnMove(from, from+8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (squareMask[from+7]&oppPieces) != 0 {
				ml.Items[count].Move = MakePawnMove(from, from+7, p.WhatPiece(from+7), promotion)
				count++
			}
			if File(from) < FileH && (squareMask[from+9]&oppPieces) != 0 {
				ml.Items[count].Move = MakePawnMove(from, from+9, p.WhatPiece(from+9), promotion)
				count++
			}
		}
		if genChecks {
			var oppKing = FirstOne(p.Kings & oppPieces)

			if (((p.Pawns&p.White & ^FileHMask)<<17)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing-9]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing-17, oppKing-9, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.White&Rank2Mask & ^FileHMask)<<25)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing-9]&allPieces) == 0 &&
				(squareMask[oppKing-17]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing-25, oppKing-9, Pawn, Empty)
				count++
			}

			if (((p.Pawns&p.White & ^FileAMask)<<15)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing-7]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing-15, oppKing-7, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.White&Rank2Mask & ^FileAMask)<<23)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing-7]&allPieces) == 0 &&
				(squareMask[oppKing-15]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing-23, oppKing-7, Pawn, Empty)
				count++
			}
		}
	} else {
		fromBB = (AllWhitePawnAttacks(oppPieces) | Rank2Mask) & p.Pawns & p.Black
		for ; fromBB != 0; fromBB &= fromBB - 1 {
			from = FirstOne(fromBB)
			promotion = let(Rank(from) == Rank2, Queen, Empty)
			if Rank(from) == Rank2 && (squareMask[from-8]&allPieces) == 0 {
				ml.Items[count].Move = MakePawnMove(from, from-8, Empty, promotion)
				count++
			}
			if File(from) > FileA && (squareMask[from-9]&oppPieces) != 0 {
				ml.Items[count].Move = MakePawnMove(from, from-9, p.WhatPiece(from-9), promotion)
				count++
			}
			if File(from) < FileH && (squareMask[from-7]&oppPieces) != 0 {
				ml.Items[count].Move = MakePawnMove(from, from-7, p.WhatPiece(from-7), promotion)
				count++
			}
		}
		if genChecks {
			var oppKing = FirstOne(p.Kings & oppPieces)

			if (((p.Pawns&p.Black & ^FileHMask)>>15)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing+7]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing+15, oppKing+7, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.Black&Rank7Mask & ^FileHMask)>>23)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing+7]&allPieces) == 0 &&
				(squareMask[oppKing+15]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing+23, oppKing+7, Pawn, Empty)
				count++
			}

			if (((p.Pawns&p.Black & ^FileAMask)>>17)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing+9]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing+17, oppKing+9, Pawn, Empty)
				count++
			}
			if (((p.Pawns&p.Black&Rank7Mask & ^FileAMask)>>25)&p.Kings&oppPieces) != 0 &&
				(squareMask[oppKing+9]&allPieces) == 0 &&
				(squareMask[oppKing+17]&allPieces) == 0 {
				ml.Items[count].Move = MakeMove(oppKing+25, oppKing+9, Pawn, Empty)
				count++
			}
		}
	}

	var checksN, checksB, checksR, checksQ uint64
	if genChecks {
		var oppKing = FirstOne(p.Kings & oppPieces)
		checksN = knightAttacks[oppKing] &^ allPieces
		checksB = BishopAttacks(oppKing, allPieces) &^ allPieces
		checksR = RookAttacks(oppKing, allPieces) &^ allPieces
		checksQ = checksB | checksR

		//discovered checks
		//TODO pawn, king discovered checks
		for fromBB = (p.Rooks | p.Queens) & ownPieces & rookMoves[oppKing]; fromBB != 0; fromBB &= fromBB - 1 {
			var blockers = betweenMask[FirstOne(fromBB)][oppKing] & allPieces
			if blockers&(blockers-1) == 0 {
				from = FirstOne(blockers)
				if (squareMask[from] & ownPieces) != 0 {
					var piece = p.WhatPiece(from)
					if piece == Knight {
						for toBB = knightAttacks[from] & ^allPieces & ^checksN; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml.Items[count].Move = MakeMove(from, to, Knight, p.WhatPiece(to))
							count++
						}
					} else if piece == Bishop {
						for toBB = BishopAttacks(from, allPieces) & ^allPieces & ^checksB; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml.Items[count].Move = MakeMove(from, to, Bishop, p.WhatPiece(to))
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
				if (squareMask[from] & ownPieces) != 0 {
					var piece = p.WhatPiece(from)
					if piece == Knight {
						for toBB = knightAttacks[from] & ^allPieces & ^checksN; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml.Items[count].Move = MakeMove(from, to, Knight, p.WhatPiece(to))
							count++
						}
					} else if piece == Rook {
						for toBB = RookAttacks(from, allPieces) & ^allPieces & ^checksR; toBB != 0; toBB &= toBB - 1 {
							to = FirstOne(toBB)
							ml.Items[count].Move = MakeMove(from, to, Rook, p.WhatPiece(to))
							count++
						}
					}
				}
			}
		}
	}

	for fromBB = p.Knights & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = knightAttacks[from] & (target | checksN); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Knight, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Bishops & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = BishopAttacks(from, allPieces) & (target | checksB); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Bishop, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Rooks & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = RookAttacks(from, allPieces) & (target | checksR); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Rook, p.WhatPiece(to))
			count++
		}
	}

	for fromBB = p.Queens & ownPieces; fromBB != 0; fromBB &= fromBB - 1 {
		from = FirstOne(fromBB)
		for toBB = QueenAttacks(from, allPieces) & (target | checksQ); toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, Queen, p.WhatPiece(to))
			count++
		}
	}

	{
		from = FirstOne(p.Kings & ownPieces)
		for toBB = kingAttacks[from] & target; toBB != 0; toBB &= toBB - 1 {
			to = FirstOne(toBB)
			ml.Items[count].Move = MakeMove(from, to, King, p.WhatPiece(to))
			count++
		}
	}

	ml.Count = count
}

func (ml *MoveList) AddPromotions(move Move, startIndex int) int {
	ml.Items[startIndex+0].Move = move ^ Move(Queen<<18)
	ml.Items[startIndex+1].Move = move ^ Move(Rook<<18)
	ml.Items[startIndex+2].Move = move ^ Move(Bishop<<18)
	ml.Items[startIndex+3].Move = move ^ Move(Knight<<18)
	return startIndex + 4
}

func (ml *MoveList) FilterLegalMoves(p *Position) {
	var child = &Position{}
	var legalMoves = 0
	for i := 0; i < ml.Count; i++ {
		if p.MakeMove(ml.Items[i].Move, child) {
			ml.Items[legalMoves] = ml.Items[i]
			legalMoves++
		}
	}
	ml.Count = legalMoves
}

func (ml *MoveList) MoveToBegin(index int) {
	if index == 0 {
		return
	}
	var item = ml.Items[index]
	for i := index; i > 0; i-- {
		ml.Items[i] = ml.Items[i-1]
	}
	ml.Items[0] = item
}

func (ml *MoveList) ElementAt(index int) Move {
	var bestIndex = index
	for i := bestIndex + 1; i < ml.Count; i++ {
		if ml.Items[i].Score > ml.Items[bestIndex].Score {
			bestIndex = i
		}
	}
	if bestIndex != index {
		var temp = ml.Items[bestIndex]
		ml.Items[bestIndex] = ml.Items[index]
		ml.Items[index] = temp
	}
	return ml.Items[index].Move
}

func (ml *MoveList) SortMoves() {
	sort.Slice(ml.Items[:ml.Count], func(i, j int) bool {
		return ml.Items[i].Score > ml.Items[j].Score
	})
}
