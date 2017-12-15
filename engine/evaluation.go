package engine

import (
	"fmt"
	"math"
)

const (
	PawnValue          = 100
	PawnEndgameBonus   = 5
	PawnDoubled        = -10
	PawnIsolated       = -15
	PawnCenter         = 10
	PawnPassed         = 130
	PawnPassedKingDist = 10
	PawnPassedSquare   = 200
	PawnPassedBlocker  = 15
	BishopPairEndgame  = 60
	WeakField          = -10
)

const (
	knightTropism = 3
	bishopTropism = 2
	rookTropism   = 4
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

	BB_WPAWN_SQUARE, BB_BPAWN_SQUARE [64]uint64
	PAWN_PASSED                      [8]int
	dist                             [][]int
)

type evaluation struct {
	experimentSettings bool
	pieceValue         []int
	kingSafety         [][]int
	knightPst          []int
	queenPst           []int
	kingOpeningPst     []int
	kingEndgamePst     []int
	bishopMobility     []int
	rookMobility       []int
	rook7th            int
	rookSemiopen       int
	rookOpen           int
	queen7th           int
	minorOnStrongField int
}

func NewEvaluation(experimentSettings bool) *evaluation {
	var e = &evaluation{}
	e.experimentSettings = experimentSettings
	e.pieceValue = []int{0, 100, 400, 400, 600, 1200}

	e.kingSafety = makeSlice2D(15, 9, 200, func(x, y float64) float64 {
		return math.Pow(1+5*x+13*y, 1.25)
	})

	var mobilityKernel = func(x float64) float64 {
		return math.Pow(x, 0.7)
	}

	e.bishopMobility = makeSlice(13, -50, 50, mobilityKernel)
	e.rookMobility = makeSlice(14, -25, 25, mobilityKernel)

	e.knightPst = scaleSlice(center[:], -35, 35)
	e.queenPst = scaleSlice(center[:], -20, 20)
	e.kingOpeningPst = scaleSlice(center_k[:], 0, -35)
	e.kingEndgamePst = scaleSlice(center[:], -30, 30)

	e.rook7th = 30
	e.rookSemiopen = 20
	e.rookOpen = 25
	e.queen7th = 20
	e.minorOnStrongField = 10

	return e
}

func (e *evaluation) MoveValue(move Move) int {
	var result = e.pieceValue[move.CapturedPiece()]
	if move.Promotion() != Empty {
		result += e.pieceValue[move.Promotion()] - e.pieceValue[Pawn]
	}
	return result
}

func (e *evaluation) Evaluate(p *Position) int {
	var (
		x, b                           uint64
		sq, keySq                      int
		wn, bn, wb, bb, wr, br, wq, bq int
		score, opening, endgame        int

		allPieces = p.White | p.Black
		wkingSq   = FirstOne(p.Kings & p.White)
		bkingSq   = FirstOne(p.Kings & p.Black)
		wp        = popcount_1s_Max15(p.Pawns & p.White)
		bp        = popcount_1s_Max15(p.Pawns & p.Black)
	)

	score += PawnIsolated * (popcount_1s_Max15(GetIsolatedPawns(p.Pawns&p.White)) -
		popcount_1s_Max15(GetIsolatedPawns(p.Pawns&p.Black)))

	score += PawnDoubled * (popcount_1s_Max15(GetDoubledPawns(p.Pawns&p.White)) -
		popcount_1s_Max15(GetDoubledPawns(p.Pawns&p.Black)))

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

	var wkingMoves = kingAttacks[wkingSq]
	var bkingMoves = kingAttacks[bkingSq]

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	var wtropism, btropism int

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		score += e.knightPst[sq]
		if (knightAttacks[sq] & bkingMoves) != 0 {
			wtropism += knightTropism
		}
		if (squareMask[sq] & wStrongFields) != 0 {
			score += e.minorOnStrongField
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		score -= e.knightPst[sq]
		if (knightAttacks[sq] & wkingMoves) != 0 {
			btropism += knightTropism
		}
		if (squareMask[sq] & bStrongFields) != 0 {
			score -= e.minorOnStrongField
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score += e.bishopMobility[PopCount(b)]
		if (b & bkingMoves) != 0 {
			wtropism += bishopTropism
		}
		if (squareMask[sq] & wStrongFields) != 0 {
			score += e.minorOnStrongField
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score -= e.bishopMobility[PopCount(b)]
		if (b & wkingMoves) != 0 {
			btropism += bishopTropism
		}
		if (squareMask[sq] & bStrongFields) != 0 {
			score -= e.minorOnStrongField
		}
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 {
			score += e.rook7th
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		if (b & bkingMoves) != 0 {
			wtropism += rookTropism
		}
		score += e.rookMobility[PopCount(b)]
		b = fileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				score += e.rookOpen
			} else {
				score += e.rookSemiopen
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 {
			score -= e.rook7th
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		if (b & wkingMoves) != 0 {
			btropism += rookTropism
		}
		score -= e.rookMobility[PopCount(b)]
		b = fileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				score -= e.rookOpen
			} else {
				score -= e.rookSemiopen
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		wq++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 {
			score += e.queen7th
		}
		score += e.queenPst[sq]
		wtropism += 7 - SquareDistance(sq, bkingSq)
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 {
			score -= e.queen7th
		}
		score -= e.queenPst[sq]
		btropism += 7 - SquareDistance(sq, wkingSq)
	}

	var matIndexWhite = min(32, (wn+wb)*3+wr*5+wq*10)
	var matIndexBlack = min(32, (bn+bb)*3+br*5+bq*10)

	for x = GetWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		score += PAWN_PASSED[Rank(sq)]

		keySq = sq + 8
		if (squareMask[keySq] & p.Black) != 0 {
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
		if (squareMask[keySq] & p.White) != 0 {
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

	opening -= e.kingSafety[min(btropism, 15)][ShelterWKingSquare(p, wkingSq)]
	opening += e.kingSafety[min(wtropism, 15)][ShelterBKingSquare(p, bkingSq)]
	endgame -= btropism * 5
	endgame += wtropism * 5

	opening += e.kingOpeningPst[wkingSq]
	endgame += e.kingEndgamePst[wkingSq]

	opening -= e.kingOpeningPst[FlipSquare(bkingSq)]
	endgame -= e.kingEndgamePst[bkingSq]

	score += 10 * (popcount_1s_Max15(wStrongFields) - popcount_1s_Max15(bStrongFields))
	score += 100*(wp-bp) + 400*(wn-bn+wb-bb) + 600*(wr-br) + 1200*(wq-bq)

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

func (e *evaluation) Trace() {
	PrintVector("bishopMobility", e.bishopMobility)
	PrintVector("rookMobility", e.rookMobility)
	PrintPst("knightPst", e.knightPst)
	PrintPst("queenPst", e.queenPst)
	PrintPst("kingOpeningPst", e.kingOpeningPst)
	PrintPst("kingEndgamePst", e.kingEndgamePst)
	PrintSlice2D("kingSafety", e.kingSafety)
}

func scaleSlice(source []int, minValue, maxValue int) []int {
	var low = source[0]
	var high = source[0]
	for _, item := range source[1:] {
		low = min(low, item)
		high = max(high, item)
	}
	var result = make([]int, len(source))
	for i := range result {
		result[i] = interpolateLinearInt(source[i], low, high, minValue, maxValue)
	}
	return result
}

func makeSlice(xmax, ymin, ymax int, g func(x float64) float64) []int {
	var gmin = g(0)
	var gmax = g(float64(xmax))
	var k = float64(ymax-ymin) / (gmax - gmin)
	var result = make([]int, xmax+1)
	for i := range result {
		result[i] = ymin + int(k*(g(float64(i))-gmin))
	}
	return result
}

func makeSlice2D(xmax, ymax, zmax int,
	g func(x, y float64) float64) [][]int {
	var k = float64(zmax) / (g(float64(xmax), float64(ymax)) - g(0, 0))
	var result = make([][]int, xmax+1)
	for i := range result {
		result[i] = make([]int, ymax+1)
		for j := range result[i] {
			result[i][j] = int(k * g(float64(i), float64(j)))
		}
	}
	return result
}

func identity(x float64) float64 { return x }

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
		var mask = fileMask[file+i-1] & p.White & p.Pawns
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
		var mask = fileMask[file+i-1] & p.Black & p.Pawns
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

func PrintPst(name string, source []int) {
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

func PrintSlice2D(name string, source [][]int) {
	fmt.Println(name)
	for _, x := range source {
		fmt.Println(x)
	}
}

func init() {
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
		var x = UpFill(squareMask[sq])
		for j := 0; j < Rank(FlipSquare(sq)); j++ {
			x |= Left(x) | Right(x)
		}
		BB_WPAWN_SQUARE[sq] = x
	}

	for sq := 0; sq < 64; sq++ {
		var x = DownFill(squareMask[sq])
		for j := 0; j < Rank(sq); j++ {
			x |= Left(x) | Right(x)
		}
		BB_BPAWN_SQUARE[sq] = x
	}
}
