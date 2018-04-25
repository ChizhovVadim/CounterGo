package engine

import (
	"math"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type score struct {
	midgame int32
	endgame int32
}

func (l *score) Add(r score) {
	l.midgame += r.midgame
	l.endgame += r.endgame
}

func (l *score) Sub(r score) {
	l.midgame -= r.midgame
	l.endgame -= r.endgame
}

func (l *score) AddN(r score, n int) {
	l.midgame += r.midgame * int32(n)
	l.endgame += r.endgame * int32(n)
}

var (
	materialPawn           = score{10000, 11000}
	materialKnight         = score{40000, 40000}
	materialBishop         = score{40000, 40000}
	materialRook           = score{60000, 60000}
	materialQueen          = score{120000, 120000}
	materialBishopPair     = score{1500, 4000}
	pstKnight              = score{1000, 1000}
	pstQueen               = score{0, 600}
	pstKingOpening         = score{-2000, 0}
	pstKingEndgame         = score{0, 1000}
	kingAttack             = score{700, 0}
	bishopMob              = score{35, 35}
	rookMob                = score{25, 50}
	rook7Th                = score{3000, 0}
	rookOpen               = score{2000, 0}
	rookSemiopen           = score{1000, 0}
	kingPawnShiled         = score{-1000, 0}
	minorOnStrongField     = score{2000, 0}
	pawnIsolated           = score{-1500, -1000}
	pawnDoubled            = score{-1000, -1000}
	pawnCenter             = score{1500, 0}
	pawnPassedAdvanceBonus = score{400, 800}
	pawnPassedFreeBonus    = score{0, 100}
	pawnPassedKingDistance = score{0, 100}
	pawnPassedSquare       = score{0, 3300}
	threat                 = score{5000, 5000}
)

const PawnValue = 100
const darkSquares uint64 = 0xAA55AA55AA55AA55

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

	pawnPassedBonus = [8]int{0, 0, 0, 2, 6, 12, 21, 0}
)

const (
	kingAttackUnitKnight = 2
	kingAttackUnitBishop = 2
	kingAttackUnitRook   = 3
	kingAttackUnitQueen  = 4
)

var (
	dist            [64][64]int
	whitePawnSquare [64]uint64
	blackPawnSquare [64]uint64
	kingZone        [64]uint64
	bishopMobility  []int
	rookMobility    []int
)

func Evaluate(p *Position) int {
	var (
		x, b                                         uint64
		sq, keySq, bonus                             int
		wn, bn, wb, bb, wr, br, wq, bq               int
		score                                        score
		wattackTot, wattackNb, battackTot, battackNb int
	)
	var allPieces = p.White | p.Black
	var wkingSq = FirstOne(p.Kings & p.White)
	var bkingSq = FirstOne(p.Kings & p.Black)
	var wp = PopCount(p.Pawns & p.White)
	var bp = PopCount(p.Pawns & p.Black)

	score.AddN(pawnIsolated,
		PopCount(getIsolatedPawns(p.Pawns&p.White))-
			PopCount(getIsolatedPawns(p.Pawns&p.Black)))

	score.AddN(pawnDoubled,
		PopCount(getDoubledPawns(p.Pawns&p.White))-
			PopCount(getDoubledPawns(p.Pawns&p.Black)))

	b = p.Pawns & p.White & (Rank4Mask | Rank5Mask | Rank6Mask)
	if (b & FileDMask) != 0 {
		score.Add(pawnCenter)
	}
	if (b & FileEMask) != 0 {
		score.Add(pawnCenter)
	}
	b = p.Pawns & p.Black & (Rank5Mask | Rank4Mask | Rank3Mask)
	if (b & FileDMask) != 0 {
		score.Sub(pawnCenter)
	}
	if (b & FileEMask) != 0 {
		score.Sub(pawnCenter)
	}

	var wStrongAttacks = AllWhitePawnAttacks(p.Pawns&p.White) & p.Black &^ p.Pawns
	var bStrongAttacks = AllBlackPawnAttacks(p.Pawns&p.Black) & p.White &^ p.Pawns
	score.AddN(threat, PopCount(wStrongAttacks)-PopCount(bStrongAttacks))

	var wkingZone = kingZone[wkingSq]
	var bkingZone = kingZone[bkingSq]

	var wMobilityArea = ^((p.Pawns & p.White) | AllBlackPawnAttacks(p.Pawns&p.Black))
	var bMobilityArea = ^((p.Pawns & p.Black) | AllWhitePawnAttacks(p.Pawns&p.White))

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		score.AddN(pstKnight, center[sq])
		if (KnightAttacks[sq] & bkingZone) != 0 {
			wattackTot += kingAttackUnitKnight
			wattackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		score.AddN(pstKnight, -center[sq])
		if (KnightAttacks[sq] & wkingZone) != 0 {
			battackTot += kingAttackUnitKnight
			battackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score.AddN(bishopMob, bishopMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitBishop
			wattackNb++
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score.AddN(bishopMob, -bishopMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackTot += kingAttackUnitBishop
			battackNb++
		}
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 &&
			((p.Pawns&p.Black&Rank7Mask) != 0 || Rank(bkingSq) == Rank8) {
			score.Add(rook7Th)
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		score.AddN(rookMob, rookMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitRook
			wattackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				score.Add(rookOpen)
			} else {
				score.Add(rookSemiopen)
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 &&
			((p.Pawns&p.White&Rank2Mask) != 0 || Rank(wkingSq) == Rank1) {
			score.Sub(rook7Th)
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		score.AddN(rookMob, -rookMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackTot += kingAttackUnitRook
			battackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				score.Sub(rookOpen)
			} else {
				score.Sub(rookSemiopen)
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		wq++
		sq = FirstOne(x)
		score.AddN(pstQueen, center[sq])
		if (QueenAttacks(sq, allPieces) & bkingZone) != 0 {
			wattackTot += kingAttackUnitQueen
			wattackNb++
		}
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		score.AddN(pstQueen, -center[sq])
		if (QueenAttacks(sq, allPieces) & wkingZone) != 0 {
			battackTot += kingAttackUnitQueen
			battackNb++
		}
	}

	score.AddN(kingAttack, wattackTot*limitValue(wattackNb-1, 0, 3)-
		battackTot*limitValue(battackNb-1, 0, 3))

	var matIndexWhite = Min(32, (wn+wb)*3+wr*5+wq*10)
	var matIndexBlack = Min(32, (bn+bb)*3+br*5+bq*10)

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(sq)]
		score.AddN(pawnPassedAdvanceBonus, bonus)
		keySq = sq + 8
		score.AddN(pawnPassedKingDistance, bonus*(dist[keySq][bkingSq]*2-dist[keySq][wkingSq]))
		if (SquareMask[keySq] & p.Black) == 0 {
			score.AddN(pawnPassedFreeBonus, bonus)
		}

		if matIndexBlack == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				score.AddN(pawnPassedSquare, Rank(f1))
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(FlipSquare(sq))]
		score.AddN(pawnPassedAdvanceBonus, -bonus)
		keySq = sq - 8
		score.AddN(pawnPassedKingDistance, -bonus*(dist[keySq][wkingSq]*2-dist[keySq][bkingSq]))
		if (SquareMask[keySq] & p.White) == 0 {
			score.AddN(pawnPassedFreeBonus, -bonus)
		}

		if matIndexWhite == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				score.AddN(pawnPassedSquare, -Rank(FlipSquare(f1)))
			}
		}
	}

	score.AddN(kingPawnShiled, shelterWKingSquare(p, wkingSq)-shelterBKingSquare(p, bkingSq))
	score.AddN(pstKingOpening, center_k[wkingSq]-center_k[FlipSquare(bkingSq)])
	score.AddN(pstKingEndgame, center[wkingSq]-center[bkingSq])

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	var wMinorOnStrongFields = wStrongFields & (p.Knights | p.Bishops) & p.White
	var bMinorOnStrongFields = bStrongFields & (p.Knights | p.Bishops) & p.Black

	score.AddN(minorOnStrongField, PopCount(wMinorOnStrongFields)-
		PopCount(bMinorOnStrongFields))

	score.AddN(materialPawn, wp-bp)
	score.AddN(materialKnight, wn-bn)
	score.AddN(materialBishop, wb-bb)
	score.AddN(materialRook, wr-br)
	score.AddN(materialQueen, wq-bq)
	if wb >= 2 {
		score.Add(materialBishopPair)
	}
	if bb >= 2 {
		score.Sub(materialBishopPair)
	}

	var phase = matIndexWhite + matIndexBlack
	var result = (int(score.midgame)*phase + int(score.endgame)*(64-phase)) / 64

	if wp == 0 && result > 0 {
		if wn+wb <= 1 && wr+wq == 0 {
			result /= 16
		} else if wn == 2 && wb+wr+wq == 0 && bp == 0 {
			result /= 16
		} else if (wn+wb+2*wr+4*wq)-(bn+bb+2*br+4*bq) <= 1 {
			result /= 4
		}
	}

	if bp == 0 && result < 0 {
		if bn+bb <= 1 && br+bq == 0 {
			result /= 16
		} else if bn == 2 && bb+br+bq == 0 && wp == 0 {
			result /= 16
		} else if (bn+bb+2*br+4*bq)-(wn+wb+2*wr+4*wq) <= 1 {
			result /= 4
		}
	}

	if (p.Knights|p.Rooks|p.Queens) == 0 &&
		wb == 1 && bb == 1 && AbsDelta(wp, bp) <= 2 &&
		(p.Bishops&darkSquares) != 0 &&
		(p.Bishops & ^darkSquares) != 0 {
		result /= 2
	}

	if !p.WhiteMove {
		result = -result
	}
	return result / 100
}

func limitValue(v, min, max int) int {
	if v <= min {
		return min
	}
	if v >= max {
		return max
	}
	return v
}

func getDoubledPawns(pawns uint64) uint64 {
	return DownFill(Down(pawns)) & pawns
}

func getIsolatedPawns(pawns uint64) uint64 {
	return ^FileFill(Left(pawns)|Right(pawns)) & pawns
}

func getWhitePassedPawns(p *Position) uint64 {
	var allFrontSpans = DownFill(Down(p.Black & p.Pawns))
	allFrontSpans |= Right(allFrontSpans) | Left(allFrontSpans)
	return p.White & p.Pawns &^ allFrontSpans
}

func getBlackPassedPawns(p *Position) uint64 {
	var allFrontSpans = UpFill(Up(p.White & p.Pawns))
	allFrontSpans |= Right(allFrontSpans) | Left(allFrontSpans)
	return p.Black & p.Pawns &^ allFrontSpans
}

func shelterWKingSquare(p *Position, square int) int {
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
		} else {
			penalty += 3
		}
	}
	return Max(0, penalty-1)
}

func shelterBKingSquare(p *Position, square int) int {
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
		} else {
			penalty += 3
		}
	}
	return Max(0, penalty-1)
}

func initProgressionSum(size int, ratio float64, min, max int) []int {
	var n = size - 1
	var q = math.Pow(ratio, 1/float64(n-1))
	var b1 = (1 - q) / (1 - math.Pow(q, float64(n)))

	var result = make([]int, size)
	for i := range result {
		var sum float64
		if i > 0 {
			sum = b1 * (1 - math.Pow(q, float64(i))) / (1 - q)
		}
		result[i] = min + int(float64(max-min)*sum)
	}
	return result
}

func init() {
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			dist[i][j] = SquareDistance(i, j)
		}
	}
	for sq := 0; sq < 64; sq++ {
		var x = UpFill(SquareMask[sq])
		for j := 0; j < Rank(FlipSquare(sq)); j++ {
			x |= Left(x) | Right(x)
		}
		whitePawnSquare[sq] = x
	}
	for sq := 0; sq < 64; sq++ {
		var x = DownFill(SquareMask[sq])
		for j := 0; j < Rank(sq); j++ {
			x |= Left(x) | Right(x)
		}
		blackPawnSquare[sq] = x
	}
	for sq := range kingZone {
		var keySq = MakeSquare(limitValue(File(sq), FileB, FileG), limitValue(Rank(sq), Rank2, Rank7))
		kingZone[sq] = SquareMask[keySq] | KingAttacks[keySq]
	}
	bishopMobility = initProgressionSum(13+1, 0.25, -100, 100)
	rookMobility = initProgressionSum(14+1, 0.25, -100, 100)
}
