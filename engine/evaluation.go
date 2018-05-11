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

type evaluator struct {
	sideToMoveBonus        int
	materialKnight         int
	materialBishop         int
	materialRook           int
	materialQueen          int
	materialBishopPair     score
	bishopMobility         []score
	rookMobility           []score
	pstKing                []score
	pawnMaterial           []int
	pstKnight              score
	pstQueen               score
	kingAttack             score
	bishopPawnsOnColor     score
	rook7Th                score
	rookOpen               score
	rookSemiopen           score
	kingPawnShiled         score
	minorOnStrongField     score
	pawnIsolated           score
	pawnDoubled            score
	pawnBackward           score
	pawnCenter             score
	pawnPassedAdvanceBonus score
	pawnPassedFreeBonus    score
	pawnPassedKingDistance score
	pawnPassedSquare       score
	threat                 score
}

func NewEvaluator() *evaluator {
	return &evaluator{
		sideToMoveBonus:        14,
		materialKnight:         400,
		materialBishop:         400,
		materialRook:           600,
		materialQueen:          1200,
		materialBishopPair:     score{15, 40},
		bishopMobility:         initMobility(13, -35, -35, 35, 35, 0.25),
		rookMobility:           initMobility(14, -25, -50, 25, 50, 0.25),
		pstKing:                initPstKing(-20, 10),
		pawnMaterial:           initProgressionSum2(8, 150, 100),
		pstKnight:              score{10, 10},
		pstQueen:               score{0, 6},
		kingAttack:             score{7, 0},
		bishopPawnsOnColor:     score{-4, -6},
		rook7Th:                score{30, 0},
		rookOpen:               score{20, 0},
		rookSemiopen:           score{10, 0},
		kingPawnShiled:         score{-10, 0},
		minorOnStrongField:     score{20, 0},
		pawnIsolated:           score{-15, -10},
		pawnDoubled:            score{-10, -10},
		pawnBackward:           score{-10, -10},
		pawnCenter:             score{15, 0},
		pawnPassedAdvanceBonus: score{4, 8},
		pawnPassedFreeBonus:    score{0, 1},
		pawnPassedKingDistance: score{0, 1},
		pawnPassedSquare:       score{0, 33},
		threat:                 score{50, 50},
	}
}

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
)

func (e *evaluator) Evaluate(p *Position) int {
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

	score.AddN(e.pawnIsolated,
		PopCount(getIsolatedPawns(p.Pawns&p.White))-
			PopCount(getIsolatedPawns(p.Pawns&p.Black)))

	score.AddN(e.pawnDoubled,
		PopCount(getDoubledPawns(p.Pawns&p.White))-
			PopCount(getDoubledPawns(p.Pawns&p.Black)))

	b = p.Pawns & p.White & (Rank4Mask | Rank5Mask | Rank6Mask)
	if (b & FileDMask) != 0 {
		score.Add(e.pawnCenter)
	}
	if (b & FileEMask) != 0 {
		score.Add(e.pawnCenter)
	}
	b = p.Pawns & p.Black & (Rank5Mask | Rank4Mask | Rank3Mask)
	if (b & FileDMask) != 0 {
		score.Sub(e.pawnCenter)
	}
	if (b & FileEMask) != 0 {
		score.Sub(e.pawnCenter)
	}

	score.AddN(e.pawnBackward,
		PopCount(p.Pawns&p.White&^AllWhitePawnAttacks(p.Pawns&p.White)&^FileFill(p.Pawns&p.Black))-
			PopCount(p.Pawns&p.Black&^AllBlackPawnAttacks(p.Pawns&p.Black)&^FileFill(p.Pawns&p.White)))

	var wStrongAttacks = AllWhitePawnAttacks(p.Pawns&p.White) & p.Black &^ p.Pawns
	var bStrongAttacks = AllBlackPawnAttacks(p.Pawns&p.Black) & p.White &^ p.Pawns
	score.AddN(e.threat, PopCount(wStrongAttacks)-PopCount(bStrongAttacks))

	var wkingZone = kingZone[wkingSq]
	var bkingZone = kingZone[bkingSq]

	var wMobilityArea = ^((p.Pawns & p.White) | AllBlackPawnAttacks(p.Pawns&p.Black))
	var bMobilityArea = ^((p.Pawns & p.Black) | AllWhitePawnAttacks(p.Pawns&p.White))

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		score.AddN(e.pstKnight, center[sq])
		if (KnightAttacks[sq] & bkingZone) != 0 {
			wattackTot += kingAttackUnitKnight
			wattackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		score.AddN(e.pstKnight, -center[sq])
		if (KnightAttacks[sq] & wkingZone) != 0 {
			battackTot += kingAttackUnitKnight
			battackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score.Add(e.bishopMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitBishop
			wattackNb++
		}

		if IsDarkSquare(sq) {
			score.AddN(e.bishopPawnsOnColor, PopCount(p.Pawns&p.White&darkSquares)-4)
		} else {
			score.AddN(e.bishopPawnsOnColor, PopCount(p.Pawns&p.White&^darkSquares)-4)
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		score.Sub(e.bishopMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackTot += kingAttackUnitBishop
			battackNb++
		}

		if IsDarkSquare(sq) {
			score.AddN(e.bishopPawnsOnColor, 4-PopCount(p.Pawns&p.Black&darkSquares))
		} else {
			score.AddN(e.bishopPawnsOnColor, 4-PopCount(p.Pawns&p.Black&^darkSquares))
		}
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 &&
			((p.Pawns&p.Black&Rank7Mask) != 0 || Rank(bkingSq) == Rank8) {
			score.Add(e.rook7Th)
		}
		b = RookAttacks(sq, allPieces)
		score.Add(e.rookMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitRook
			wattackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				score.Add(e.rookOpen)
			} else {
				score.Add(e.rookSemiopen)
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 &&
			((p.Pawns&p.White&Rank2Mask) != 0 || Rank(wkingSq) == Rank1) {
			score.Sub(e.rook7Th)
		}
		b = RookAttacks(sq, allPieces)
		score.Sub(e.rookMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackTot += kingAttackUnitRook
			battackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				score.Sub(e.rookOpen)
			} else {
				score.Sub(e.rookSemiopen)
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		wq++
		sq = FirstOne(x)
		score.AddN(e.pstQueen, center[sq])
		if (QueenAttacks(sq, allPieces) & bkingZone) != 0 {
			wattackTot += kingAttackUnitQueen
			wattackNb++
		}
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		score.AddN(e.pstQueen, -center[sq])
		if (QueenAttacks(sq, allPieces) & wkingZone) != 0 {
			battackTot += kingAttackUnitQueen
			battackNb++
		}
	}

	score.AddN(e.kingAttack, wattackTot*limitValue(wattackNb-1, 0, 3)-
		battackTot*limitValue(battackNb-1, 0, 3))

	var matIndexWhite = Min(32, (wn+wb)*3+wr*5+wq*10)
	var matIndexBlack = Min(32, (bn+bb)*3+br*5+bq*10)

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(sq)]
		score.AddN(e.pawnPassedAdvanceBonus, bonus)
		keySq = sq + 8
		score.AddN(e.pawnPassedKingDistance, bonus*(dist[keySq][bkingSq]*2-dist[keySq][wkingSq]))
		if (SquareMask[keySq] & p.Black) == 0 {
			score.AddN(e.pawnPassedFreeBonus, bonus)
		}

		if matIndexBlack == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				score.AddN(e.pawnPassedSquare, Rank(f1))
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(FlipSquare(sq))]
		score.AddN(e.pawnPassedAdvanceBonus, -bonus)
		keySq = sq - 8
		score.AddN(e.pawnPassedKingDistance, -bonus*(dist[keySq][wkingSq]*2-dist[keySq][bkingSq]))
		if (SquareMask[keySq] & p.White) == 0 {
			score.AddN(e.pawnPassedFreeBonus, -bonus)
		}

		if matIndexWhite == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				score.AddN(e.pawnPassedSquare, -Rank(FlipSquare(f1)))
			}
		}
	}

	score.AddN(e.kingPawnShiled, shelterWKingSquare(p, wkingSq)-shelterBKingSquare(p, bkingSq))
	score.Add(e.pstKing[wkingSq])
	score.Sub(e.pstKing[FlipSquare(bkingSq)])

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	var wMinorOnStrongFields = wStrongFields & (p.Knights | p.Bishops) & p.White
	var bMinorOnStrongFields = bStrongFields & (p.Knights | p.Bishops) & p.Black

	score.AddN(e.minorOnStrongField, PopCount(wMinorOnStrongFields)-
		PopCount(bMinorOnStrongFields))

	if wb >= 2 {
		score.Add(e.materialBishopPair)
	}
	if bb >= 2 {
		score.Sub(e.materialBishopPair)
	}

	var phase = matIndexWhite + matIndexBlack
	var result = (int(score.midgame)*phase + int(score.endgame)*(64-phase)) / 64

	result += e.pawnMaterial[Min(wp, 8)] - e.pawnMaterial[Min(bp, 8)]
	result += e.materialKnight*(wn-bn) +
		e.materialBishop*(wb-bb) +
		e.materialRook*(wr-br) +
		e.materialQueen*(wq-bq)

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

	if !p.IsCheck() {
		result += e.sideToMoveBonus
	}

	return result
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

func initMobility(maxMob, minMg, minEg, maxMg, maxEg int, ratio float64) []score {
	var q = math.Pow(ratio, 1/float64(maxMob-1))
	var b1 = (1 - q) / (1 - math.Pow(q, float64(maxMob)))

	var result = make([]score, 1+maxMob)
	var sum = 0.0
	for i := range result {
		if i > 0 {
			sum = b1 * (1 - math.Pow(q, float64(i))) / (1 - q)
		}
		var mg = float64(minMg) + float64(maxMg-minMg)*sum
		var eg = float64(minEg) + float64(maxEg-minEg)*sum
		result[i] = score{int32(mg), int32(eg)}
	}
	return result
}

func initProgressionSum2(n, first, last int) []int {
	var q = math.Pow(float64(last)/float64(first), 1/float64(n-1))
	var result = make([]int, n+1)
	var item = float64(first)
	var sum = item
	for i := 1; i <= n; i++ {
		result[i] = int(sum)
		item *= q
		sum += item
	}
	return result
}

func initPstKing(kingCentreOpening, kingCentreEndgame int) []score {
	var pstKing = make([]score, 64)
	for sq := range pstKing {
		pstKing[sq] = score{int32(kingCentreOpening * center_k[sq]), int32(kingCentreEndgame * center[sq])}
	}
	return pstKing
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
}
