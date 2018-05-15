package engine

import (
	"fmt"

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

type evaluationService struct {
	TraceEnabled           bool
	bishopPair             int
	bishopMobility         []score
	rookMobility           []score
	pstKing                []score
	pstKnight              score
	pstQueen               score
	kingAttack             score
	rook7Th                score
	rookOpen               score
	rookSemiopen           score
	kingPawnShiled         score
	minorOnStrongField     score
	pawnIsolated           score
	pawnDoubled            score
	pawnCenter             score
	pawnPassedAdvanceBonus score
	pawnPassedFreeBonus    score
	pawnPassedKingDistance score
	pawnPassedSquare       score
	threat                 int
}

func NewEvaluationService() *evaluationService {
	return &evaluationService{
		bishopPair:             50,
		bishopMobility:         initMobility(13, -30, -30, 35, 35),
		rookMobility:           initMobility(14, -14, -28, 14, 28),
		pstKing:                initPstKing(-20, 10),
		pstKnight:              score{10, 10},
		pstQueen:               score{0, 6},
		kingAttack:             score{7, 0},
		rook7Th:                score{20, 40},
		rookOpen:               score{20, 20},
		rookSemiopen:           score{10, 10},
		kingPawnShiled:         score{-10, 0},
		minorOnStrongField:     score{20, 0},
		pawnIsolated:           score{-10, -20},
		pawnDoubled:            score{-10, -20},
		pawnCenter:             score{15, 0},
		pawnPassedAdvanceBonus: score{4, 8},
		pawnPassedFreeBonus:    score{0, 1},
		pawnPassedKingDistance: score{0, 1},
		pawnPassedSquare:       score{0, 33},
		threat:                 50,
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

func (e *evaluationService) Evaluate(p *Position) int {
	var (
		x, b                                         uint64
		sq, keySq, bonus                             int
		wn, bn, wb, bb, wr, br, wq, bq               int
		pieceScore, kingScore, pawnScore             score
		wattackTot, wattackNb, battackTot, battackNb int
	)
	var allPieces = p.White | p.Black
	var wkingSq = FirstOne(p.Kings & p.White)
	var bkingSq = FirstOne(p.Kings & p.Black)
	var wp = PopCount(p.Pawns & p.White)
	var bp = PopCount(p.Pawns & p.Black)

	pawnScore.AddN(e.pawnIsolated,
		PopCount(getIsolatedPawns(p.Pawns&p.White))-
			PopCount(getIsolatedPawns(p.Pawns&p.Black)))

	pawnScore.AddN(e.pawnDoubled,
		PopCount(getDoubledPawns(p.Pawns&p.White))-
			PopCount(getDoubledPawns(p.Pawns&p.Black)))

	b = p.Pawns & p.White & (Rank4Mask | Rank5Mask | Rank6Mask)
	if (b & FileDMask) != 0 {
		pawnScore.Add(e.pawnCenter)
	}
	if (b & FileEMask) != 0 {
		pawnScore.Add(e.pawnCenter)
	}
	b = p.Pawns & p.Black & (Rank5Mask | Rank4Mask | Rank3Mask)
	if (b & FileDMask) != 0 {
		pawnScore.Sub(e.pawnCenter)
	}
	if (b & FileEMask) != 0 {
		pawnScore.Sub(e.pawnCenter)
	}

	var wStrongAttacks = AllWhitePawnAttacks(p.Pawns&p.White) & p.Black &^ p.Pawns
	var bStrongAttacks = AllBlackPawnAttacks(p.Pawns&p.Black) & p.White &^ p.Pawns
	var threatScore = e.threat * (PopCount(wStrongAttacks) - PopCount(bStrongAttacks))

	var wkingZone = kingZone[wkingSq]
	var bkingZone = kingZone[bkingSq]

	var wMobilityArea = ^((p.Pawns & p.White) | AllBlackPawnAttacks(p.Pawns&p.Black))
	var bMobilityArea = ^((p.Pawns & p.Black) | AllWhitePawnAttacks(p.Pawns&p.White))

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		pieceScore.AddN(e.pstKnight, center[sq])
		if (KnightAttacks[sq] & bkingZone) != 0 {
			wattackTot += kingAttackUnitKnight
			wattackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		pieceScore.AddN(e.pstKnight, -center[sq])
		if (KnightAttacks[sq] & wkingZone) != 0 {
			battackTot += kingAttackUnitKnight
			battackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		pieceScore.Add(e.bishopMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitBishop
			wattackNb++
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		pieceScore.Sub(e.bishopMobility[PopCount(b&bMobilityArea)])
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
			pieceScore.Add(e.rook7Th)
		}
		b = RookAttacks(sq, allPieces)
		pieceScore.Add(e.rookMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackTot += kingAttackUnitRook
			wattackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				pieceScore.Add(e.rookOpen)
			} else {
				pieceScore.Add(e.rookSemiopen)
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 &&
			((p.Pawns&p.White&Rank2Mask) != 0 || Rank(wkingSq) == Rank1) {
			pieceScore.Sub(e.rook7Th)
		}
		b = RookAttacks(sq, allPieces)
		pieceScore.Sub(e.rookMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackTot += kingAttackUnitRook
			battackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				pieceScore.Sub(e.rookOpen)
			} else {
				pieceScore.Sub(e.rookSemiopen)
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		wq++
		sq = FirstOne(x)
		pieceScore.AddN(e.pstQueen, center[sq])
		if (QueenAttacks(sq, allPieces) & bkingZone) != 0 {
			wattackTot += kingAttackUnitQueen
			wattackNb++
		}
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		pieceScore.AddN(e.pstQueen, -center[sq])
		if (QueenAttacks(sq, allPieces) & wkingZone) != 0 {
			battackTot += kingAttackUnitQueen
			battackNb++
		}
	}

	kingScore.AddN(e.kingAttack, wattackTot*limitValue(wattackNb-1, 0, 3)-
		battackTot*limitValue(battackNb-1, 0, 3))

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(sq)]
		pawnScore.AddN(e.pawnPassedAdvanceBonus, bonus)
		keySq = sq + 8
		pawnScore.AddN(e.pawnPassedKingDistance, bonus*(dist[keySq][bkingSq]*2-dist[keySq][wkingSq]))
		if (SquareMask[keySq] & p.Black) == 0 {
			pawnScore.AddN(e.pawnPassedFreeBonus, bonus)
		}

		if bn+bb+br+bq == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				pawnScore.AddN(e.pawnPassedSquare, Rank(f1))
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(FlipSquare(sq))]
		pawnScore.AddN(e.pawnPassedAdvanceBonus, -bonus)
		keySq = sq - 8
		pawnScore.AddN(e.pawnPassedKingDistance, -bonus*(dist[keySq][wkingSq]*2-dist[keySq][bkingSq]))
		if (SquareMask[keySq] & p.White) == 0 {
			pawnScore.AddN(e.pawnPassedFreeBonus, -bonus)
		}

		if wn+wb+wr+wq == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				pawnScore.AddN(e.pawnPassedSquare, -Rank(FlipSquare(f1)))
			}
		}
	}

	kingScore.AddN(e.kingPawnShiled, shelterWKingSquare(p, wkingSq)-shelterBKingSquare(p, bkingSq))
	kingScore.Add(e.pstKing[wkingSq])
	kingScore.Sub(e.pstKing[FlipSquare(bkingSq)])

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	var wMinorOnStrongFields = wStrongFields & (p.Knights | p.Bishops) & p.White
	var bMinorOnStrongFields = bStrongFields & (p.Knights | p.Bishops) & p.Black

	pieceScore.AddN(e.minorOnStrongField, PopCount(wMinorOnStrongFields)-
		PopCount(bMinorOnStrongFields))

	var phase = Min(24, wn+bn+wb+bb+2*(wr+br)+4*(wq+bq))
	var result = (int(pieceScore.midgame+pawnScore.midgame+kingScore.midgame)*phase +
		int(pieceScore.endgame+pawnScore.endgame+kingScore.endgame)*(24-phase)) / 24

	var pm = 2*(wn-bn+wb-bb) + 3*(wr-br) + 6*(wq-bq)
	var materialScore = 100*(wp-bp) + 150*pm + 50*limitValue(pm, -1, 1)
	if wb >= 2 {
		materialScore += e.bishopPair
	}
	if bb >= 2 {
		materialScore -= e.bishopPair
	}
	result += materialScore
	result += threatScore

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

	if e.TraceEnabled {
		fmt.Println("Pawns:", pawnScore)
		fmt.Println("Pieces:", pieceScore)
		fmt.Println("King:", kingScore)
		fmt.Println("Material:", materialScore)
		fmt.Println("Threats:", threatScore)
		fmt.Println("TOTAL:", result)
	}

	if !p.WhiteMove {
		result = -result
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
		} else if (mask & Rank4Mask) != 0 {
			penalty += 2
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
		} else if (mask & Rank5Mask) != 0 {
			penalty += 2
		} else {
			penalty += 3
		}
	}
	return Max(0, penalty-1)
}

func initMobility(maxMob, minMg, minEg, maxMg, maxEg int) []score {
	var result = make([]score, 1+maxMob)
	for i := range result {
		var mg = lirp(i, 0, maxMob, minMg, maxMg)
		var eg = lirp(i, 0, maxMob, minEg, maxEg)
		result[i] = score{int32(mg), int32(eg)}
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

func lirp(x, x_min, x_max, y_min, y_max int) int {
	if x > x_max {
		x = x_max
	} else if x < x_min {
		x = x_min
	}
	return ((y_max-y_min)*(x-x_min)+(x_max-x_min)/2)/(x_max-x_min) + y_min
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
		var keySq = MakeSquare(limitValue(File(sq), FileB, FileG), Rank(sq))
		kingZone[sq] = SquareMask[keySq] | KingAttacks[keySq]
	}
}
