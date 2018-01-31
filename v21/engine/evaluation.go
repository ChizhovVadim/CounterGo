package engine

import (
	"fmt"
	"math"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	PawnValue          = 100
	KnightValue        = 400
	BishopValue        = 400
	RookValue          = 600
	QueenValue         = 1200
	PawnEndgameBonus   = 5
	PawnDoubled        = -10
	PawnIsolated       = -15
	PawnCenter         = 10
	PawnPassed         = 130
	PawnPassedKingDist = 10
	PawnPassedSquare   = 200
	PawnPassedBlocker  = 15
	KnightCenter       = 35
	KnightKingTropism  = 15
	BishopPairEndgame  = 60
	BishopMobility     = 50
	BishopKingTropism  = 10
	Rook7th            = 30
	RookSemiopen       = 20
	RookOpen           = 25
	RookMobility       = 25
	RookKingTropism    = 15
	QueenKingTropism   = 80
	QueenCenterEndgame = 20
	Queen7th           = 20
	KingCenterMidgame  = -35
	KingCenterEndgame  = 30
	KingPawnShield     = 120
	WeakField          = -10
	MinorOnStrongField = 10
)

const DarkSquares uint64 = 0xAA55AA55AA55AA55

var (
	center = [64]int{
		-3, -2, -1, 0, 0, -1, -2, -3,
		-2, -1, 0, 1, 1, 0, -1, -2,
		-1, 0, 1, 2, 2, 1, 0, -1,
		0, 1, 2, 3, 3, 2, 1, 0,
		0, 1, 2, 3, 3, 2, 1, 0,
		-1, 0, 1, 2, 2, 1, 0, -1,
		-2, -1, 0, 1, 1, 0, -1, -2,
		-3, -2, -1, 0, 0, -1, -2, -3,
	}

	center_k = [64]int{
		1, 0, 0, 1, 0, 1, 0, 1,
		2, 2, 2, 2, 2, 2, 2, 2,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4,
	}

	knightPst, queenEndgamePst, kingOpeningPst, kingEndgamePst [64]int
	BB_WPAWN_SQUARE, BB_BPAWN_SQUARE                           [64]uint64
	BISHOP_MOBILITY                                            [13 + 1]int
	ROOK_MOBILITY                                              [14 + 1]int
	KNIGHT_KING_TROPISM, BISHOP_KING_TROPISM,
	ROOK_KING_TROPISM, QUEEN_KING_TROPISM [10]int
	KING_PAWN_SHIELD [10]int
	PAWN_PASSED      [8]int
	dist             [][]int
)

func Evaluate(p *Position) int {
	var (
		x, b                           uint64
		sq, keySq                      int
		wn, bn, wb, bb, wr, br, wq, bq int
		score, opening, endgame        int

		allPieces = p.White | p.Black
		wkingSq   = FirstOne(p.Kings & p.White)
		bkingSq   = FirstOne(p.Kings & p.Black)
		wp        = Popcount_1s_Max15(p.Pawns & p.White)
		bp        = Popcount_1s_Max15(p.Pawns & p.Black)
	)

	score += PawnIsolated * (Popcount_1s_Max15(GetIsolatedPawns(p.Pawns&p.White)) -
		Popcount_1s_Max15(GetIsolatedPawns(p.Pawns&p.Black)))

	score += PawnDoubled * (Popcount_1s_Max15(GetDoubledPawns(p.Pawns&p.White)) -
		Popcount_1s_Max15(GetDoubledPawns(p.Pawns&p.Black)))

	b = p.Pawns & p.White & (Rank4Mask | Rank5Mask | Rank6Mask)
	if (b & FileDMask) != 0 {
		score += PawnCenter
	}
	if (b & FileEMask) != 0 {
		score += PawnCenter
	}

	b = p.Pawns & p.Black & (Rank5Mask | Rank4Mask | Rank3Mask)
	if (b & FileDMask) != 0 {
		score -= PawnCenter
	}
	if (b & FileEMask) != 0 {
		score -= PawnCenter
	}

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		score += knightPst[sq]
		score += KNIGHT_KING_TROPISM[dist[sq][bkingSq]]
		b = KnightAttacks[sq]
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		score -= knightPst[sq]
		score -= KNIGHT_KING_TROPISM[dist[sq][wkingSq]]
		b = KnightAttacks[sq]
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score += BISHOP_MOBILITY[Popcount_1s_Max15(b)]
		score += BISHOP_KING_TROPISM[dist[sq][bkingSq]]
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score -= BISHOP_MOBILITY[Popcount_1s_Max15(b)]
		score -= BISHOP_KING_TROPISM[dist[sq][wkingSq]]
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 {
			score += Rook7th
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		score += ROOK_MOBILITY[Popcount_1s_Max15(b)]
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				score += RookOpen
			} else {
				score += RookSemiopen
			}
		}
		score += ROOK_KING_TROPISM[dist[sq][bkingSq]]
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 {
			score -= Rook7th
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		score -= ROOK_MOBILITY[Popcount_1s_Max15(b)]
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				score -= RookOpen
			} else {
				score -= RookSemiopen
			}
		}
		score -= ROOK_KING_TROPISM[dist[sq][wkingSq]]
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		wq++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 {
			score += Queen7th
		}
		endgame += queenEndgamePst[sq]
		score += QUEEN_KING_TROPISM[dist[sq][bkingSq]]
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 {
			score -= Queen7th
		}
		endgame -= queenEndgamePst[sq]
		score -= QUEEN_KING_TROPISM[dist[sq][wkingSq]]
	}

	var matIndexWhite = Min(32, (wn+wb)*3+wr*5+wq*10)
	var matIndexBlack = Min(32, (bn+bb)*3+br*5+bq*10)

	for x = GetWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		score += PAWN_PASSED[Rank(sq)]

		keySq = sq + 8
		if (SquareMask[keySq] & p.Black) != 0 {
			score -= PawnPassedBlocker
		}

		if matIndexBlack == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (BB_WPAWN_SQUARE[f1] & p.Kings & p.Black) == 0 {
				score += PawnPassedSquare * Rank(f1) / 6
			}
		} else if matIndexBlack < 10 {
			score += PawnPassedKingDist * dist[keySq][bkingSq]
		}
	}

	for x = GetBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		score -= PAWN_PASSED[Rank(FlipSquare(sq))]

		keySq = sq - 8
		if (SquareMask[keySq] & p.White) != 0 {
			score += PawnPassedBlocker
		}

		if matIndexWhite == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (BB_BPAWN_SQUARE[f1] & p.Kings & p.White) == 0 {
				score -= PawnPassedSquare * Rank(FlipSquare(f1)) / 6
			}
		} else if matIndexWhite < 10 {
			score -= PawnPassedKingDist * dist[keySq][wkingSq]
		}
	}

	opening += KING_PAWN_SHIELD[ShelterWKingSquare(p, wkingSq)]
	opening += kingOpeningPst[wkingSq]
	endgame += kingEndgamePst[wkingSq]

	opening -= KING_PAWN_SHIELD[ShelterBKingSquare(p, bkingSq)]
	opening -= kingOpeningPst[FlipSquare(bkingSq)]
	endgame -= kingEndgamePst[bkingSq]

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	score += 10 * (Popcount_1s_Max15(wStrongFields) - Popcount_1s_Max15(bStrongFields))

	var wMinorOnStrongFields = wStrongFields & (p.Knights | p.Bishops) & p.White
	var bMinorOnStrongFields = bStrongFields & (p.Knights | p.Bishops) & p.Black

	score += MinorOnStrongField * (Popcount_1s_Max15(wMinorOnStrongFields) -
		Popcount_1s_Max15(bMinorOnStrongFields))

	score += PawnValue*(wp-bp) + KnightValue*(wn-bn) +
		BishopValue*(wb-bb) + RookValue*(wr-br) + QueenValue*(wq-bq)

	endgame += PawnEndgameBonus * (wp - bp)
	if wb >= 2 {
		endgame += BishopPairEndgame
	}
	if bb >= 2 {
		endgame -= BishopPairEndgame
	}

	var phase = matIndexWhite + matIndexBlack
	score += (opening*phase + endgame*(64-phase)) / 64

	if wp == 0 && score > 0 {
		if wn <= 2 && wb+wr+wq == 0 {
			score /= 2
		}
		if wn+wb <= 1 && wr+wq == 0 {
			score /= 2
		}
	}

	if bp == 0 && score < 0 {
		if bn <= 2 && bb+br+bq == 0 {
			score /= 2
		}
		if bn+bb <= 1 && br+bq == 0 {
			score /= 2
		}
	}

	if (p.Knights|p.Rooks|p.Queens) == 0 &&
		wb == 1 && bb == 1 && AbsDelta(wp, bp) <= 2 &&
		(p.Bishops&DarkSquares) != 0 &&
		(p.Bishops & ^DarkSquares) != 0 {
		score /= 2
	}

	if !p.WhiteMove {
		score = -score
	}
	return score
}

func GetDoubledPawns(pawns uint64) uint64 {
	return DownFill(Down(pawns)) & pawns
}

func GetIsolatedPawns(pawns uint64) uint64 {
	return ^FileFill(Left(pawns)|Right(pawns)) & pawns
}

func GetWhitePassedPawns(p *Position) uint64 {
	var allFrontSpans = DownFill(Down(p.Black & p.Pawns))
	allFrontSpans |= Right(allFrontSpans) | Left(allFrontSpans)
	return p.White & p.Pawns &^ allFrontSpans
}

func GetBlackPassedPawns(p *Position) uint64 {
	var allFrontSpans = UpFill(Up(p.White & p.Pawns))
	allFrontSpans |= Right(allFrontSpans) | Left(allFrontSpans)
	return p.Black & p.Pawns &^ allFrontSpans
}

func ShelterWKingSquare(p *Position, square int) int {
	var file = File(square)
	if file == FileA {
		file++
	} else if file == FileH {
		file--
	}
	var penalty = 0
	for i := 0; i < 3; i++ {
		var mask = FileMask[file+i-1] & p.White & p.Pawns
		if (mask & Rank2Mask) != 0 {

		} else if (mask & Rank3Mask) != 0 {
			penalty += 1
		} else if (mask & Rank4Mask) != 0 {
			penalty += 2
		} else {
			penalty += 3
		}
	}
	return penalty
}

func ShelterBKingSquare(p *Position, square int) int {
	var file = File(square)
	if file == FileA {
		file++
	} else if file == FileH {
		file--
	}
	var penalty = 0
	for i := 0; i < 3; i++ {
		var mask = FileMask[file+i-1] & p.Black & p.Pawns
		if (mask & Rank7Mask) != 0 {
		} else if (mask & Rank6Mask) != 0 {
			penalty += 1
		} else if (mask & Rank5Mask) != 0 {
			penalty += 2
		} else {
			penalty += 3
		}
	}
	return penalty
}

func InterpolateSquare(arg, argFrom, argTo, valFrom, valTo float64) float64 {
	// A*x*x + B*x + C
	var x = arg - argFrom
	var xMax = argTo - argFrom
	var A = (valTo - valFrom) / (xMax * xMax)
	var B = 0.0
	var C = valFrom
	return A*x*x + B*x + C
}

func InterpolateLinear(arg, argFrom, argTo, valFrom, valTo float64) float64 {
	// A*x + B
	return valFrom + (valTo-valFrom)*(arg-argFrom)/(argTo-argFrom)
}

func PrintPst(name string, source [64]int) {
	fmt.Println(name)
	for i := 0; i < 64; i++ {
		var sq = FlipSquare(i)
		fmt.Printf("%3v", source[sq])
		if File(sq) == FileH {
			fmt.Println()
		} else {
			fmt.Print(" ")
		}
	}
}

func PrintVector(name string, source []int) {
	fmt.Printf("%v %v\n", name, source)
}

func TraceEvalSettings() {
	PrintPst("KnightPst", knightPst)
	PrintPst("QueenEndgamePst", queenEndgamePst)
	PrintPst("KingOpeningPst", kingOpeningPst)
	PrintPst("KingEndgamePst", kingEndgamePst)

	PrintVector("BISHOP_MOBILITY", BISHOP_MOBILITY[:])
	PrintVector("ROOK_MOBILITY", ROOK_MOBILITY[:])

	PrintVector("KNIGHT_KING_TROPISM", KNIGHT_KING_TROPISM[:])
	PrintVector("BISHOP_KING_TROPISM", BISHOP_KING_TROPISM[:])
	PrintVector("ROOK_KING_TROPISM", ROOK_KING_TROPISM[:])
	PrintVector("QUEEN_KING_TROPISM", QUEEN_KING_TROPISM[:])
	PrintVector("KING_PAWN_SHIELD", KING_PAWN_SHIELD[:])
	PrintVector("PAWN_PASSED", PAWN_PASSED[:])
}

func init() {
	for sq := 0; sq < 64; sq++ {
		knightPst[sq] = center[sq] * KnightCenter / 3
		queenEndgamePst[sq] = center[sq] * QueenCenterEndgame / 3
		kingOpeningPst[sq] = center_k[sq] * KingCenterMidgame / 4
		kingEndgamePst[sq] = center[sq] * KingCenterEndgame / 3
	}

	for i := 0; i < len(BISHOP_MOBILITY); i++ {
		BISHOP_MOBILITY[i] = int(InterpolateSquare(float64(i), 13, 1, BishopMobility, -BishopMobility))
	}

	for i := 0; i < len(ROOK_MOBILITY); i++ {
		ROOK_MOBILITY[i] = int(InterpolateSquare(float64(i), 14, 2, RookMobility, -RookMobility))
	}

	for i := 0; i < 10; i++ {
		KNIGHT_KING_TROPISM[i] = int(InterpolateLinear(float64(i), 9, 1, 0, KnightKingTropism))
		BISHOP_KING_TROPISM[i] = int(InterpolateLinear(float64(i), 9, 1, 0, BishopKingTropism))
		ROOK_KING_TROPISM[i] = int(InterpolateLinear(float64(i), 9, 1, 0, RookKingTropism))
		QUEEN_KING_TROPISM[i] = int(InterpolateLinear(float64(i), 9, 1, 0, QueenKingTropism))
	}

	for i := 0; i < len(KING_PAWN_SHIELD); i++ {
		KING_PAWN_SHIELD[i] = -int(InterpolateSquare(float64(i), 0, 9, 0, KingPawnShield))
	}

	for i := 0; i < len(PAWN_PASSED); i++ {
		PAWN_PASSED[i] = int(InterpolateSquare(float64(i), 0, 7, 0, PawnPassed))
	}

	dist = make([][]int, 64)
	for i := 0; i < 64; i++ {
		dist[i] = make([]int, 64)
		for j := 0; j < 64; j++ {
			var rd = RankDistance(i, j)
			var fd = FileDistance(i, j)
			dist[i][j] = int(math.Sqrt(float64(rd*rd + fd*fd)))
		}
	}

	for sq := 0; sq < 64; sq++ {
		var x = UpFill(SquareMask[sq])
		for j := 0; j < Rank(FlipSquare(sq)); j++ {
			x |= Left(x) | Right(x)
		}
		BB_WPAWN_SQUARE[sq] = x
	}

	for sq := 0; sq < 64; sq++ {
		var x = DownFill(SquareMask[sq])
		for j := 0; j < Rank(sq); j++ {
			x |= Left(x) | Right(x)
		}
		BB_BPAWN_SQUARE[sq] = x
	}
}
