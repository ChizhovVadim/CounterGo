package engine

import (
	"fmt"
	"math"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	PawnValue = 100
	maxPhase  = 24
)

type score struct {
	opening int
	endgame int
}

func (l *score) Add(r score) {
	l.opening += r.opening
	l.endgame += r.endgame
}

func (l *score) Sub(r score) {
	l.opening -= r.opening
	l.endgame -= r.endgame
}

func (l *score) AddN(r score, n int) {
	l.opening += r.opening * n
	l.endgame += r.endgame * n
}

func (s *score) Mix(phase int) int {
	return (int(s.opening)*phase + int(s.endgame)*(maxPhase-phase)) / maxPhase
}

type evaluationService struct {
	TraceEnabled        bool
	experimentSettings  bool
	materialPawn        score
	materialKnight      score
	materialBishop      score
	materialRook        score
	materialQueen       score
	bishopPair          score
	knightMobility      [1 + 8]score
	bishopMobility      [1 + 13]score
	rookMobility        [1 + 14]score
	queenMobility       [1 + 27]score
	pstKnight           [64]score
	pstBishop           [64]score
	pstQueen            [64]score
	pstKing             [64]score
	kingOpeningPenalty  [64]int
	rook7Th             score
	rookOpen            score
	rookSemiopen        score
	pawnIsolated        score
	pawnDoubled         score
	pawnDoubledIsolated score
	pawnPassedBonus     [8]int
	threat              score
}

func NewEvaluationService() *evaluationService {
	var srv = &evaluationService{
		materialPawn:        score{83, 100},
		materialKnight:      score{333, 350},
		materialBishop:      score{333, 350},
		materialRook:        score{500, 525},
		materialQueen:       score{1000, 1050},
		bishopPair:          score{25, 50},
		rook7Th:             score{15, 35},
		rookOpen:            score{35, 15},
		rookSemiopen:        score{2, 10},
		pawnIsolated:        score{-10, -10},
		pawnDoubled:         score{-10, -10},
		pawnDoubledIsolated: score{-30, -30},
		pawnPassedBonus:     [8]int{0, 0, 0, 2, 6, 12, 21, 0},
		threat:              score{50, 50},
	}
	const (
		KnightMobility    = 40
		KnightCenter      = 60
		BishopCenter      = 30
		QueenMobility     = 90
		QueenCenter       = 20
		KingCenterEndgame = 100
	)
	var (
		KnightLine = [8]int{-3, -1, 0, 1, 1, 0, -1, -3}
		BishopLine = [8]int{-1, 0, 1, 2, 2, 1, 0, -1}
		QueenLine  = [8]int{-1, 0, 1, 2, 2, 1, 0, -1}
		KingLine   = [8]int{-3, -1, 0, 1, 1, 0, -1, -3}
	)
	for sq := 0; sq < 64; sq++ {
		var f = File(sq)
		var r = Rank(sq)
		srv.pstKnight[sq] = score{
			KnightCenter * (KnightLine[f] + KnightLine[r]) / 8,
			KnightCenter * (KnightLine[f] + KnightLine[r]) / 8,
		}
		srv.pstBishop[sq] = score{
			BishopCenter * Min(BishopLine[f], BishopLine[r]) / 3,
			BishopCenter * Min(BishopLine[f], BishopLine[r]) / 3,
		}
		srv.pstQueen[sq] = score{0, QueenCenter * (QueenLine[f] + QueenLine[r]) / 6}
		srv.kingOpeningPenalty[sq] = Min(dist[sq][SquareB1], dist[sq][SquareG1])
		srv.pstKing[sq] = score{0, KingCenterEndgame * (KingLine[f] + KingLine[r]) / 8}
	}

	var b = math.Log(float64(KnightMobility)/float64(QueenMobility)) / math.Log(float64(8)/float64(27))
	var kernel = func(x int) int {
		return int(QueenMobility * math.Pow(float64(x)/27, b))
	}

	var mobilityBase = -kernel(13) / 2
	for m := range srv.knightMobility {
		srv.knightMobility[m] = score{
			kernel(m) + mobilityBase,
			kernel(m) + mobilityBase,
		}
	}
	for m := range srv.bishopMobility {
		srv.bishopMobility[m] = score{
			kernel(m) + mobilityBase,
			kernel(m) + mobilityBase,
		}
	}
	for m := range srv.rookMobility {
		srv.rookMobility[m] = score{
			kernel(m)/2 + mobilityBase,
			kernel(m) + mobilityBase,
		}
	}
	for m := range srv.queenMobility {
		srv.queenMobility[m] = score{
			kernel(m)*2/3 + mobilityBase,
			kernel(m) + mobilityBase,
		}
	}

	return srv
}

func NewExperimentEvaluationService() *evaluationService {
	var srv = NewEvaluationService()
	srv.experimentSettings = true
	return srv
}

func initPowerKernel(fullX, halfX float64) func(int) float64 {
	var b = math.Log(0.5) / math.Log(halfX/fullX)
	return func(x int) float64 {
		return math.Pow(float64(x)/fullX, b)
	}
}

const (
	darkSquares uint64 = 0xAA55AA55AA55AA55
)

const (
	kingAttackUnitKnight = 3
	kingAttackUnitBishop = 2
	kingAttackUnitRook   = 4
	kingAttackUnitQueen  = 6
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

	pawnScore.AddN(e.pawnDoubledIsolated,
		PopCount(getIsolatedPawns(p.Pawns&p.White)&getDoubledPawns(p.Pawns&p.White))-
			PopCount(getIsolatedPawns(p.Pawns&p.Black)&getDoubledPawns(p.Pawns&p.Black)))

	var wpawnAttacks = AllWhitePawnAttacks(p.Pawns & p.White)
	var bpawnAttacks = AllBlackPawnAttacks(p.Pawns & p.Black)

	var wkingZone = kingZone[wkingSq]
	var bkingZone = kingZone[bkingSq]

	if (wpawnAttacks & bkingZone) != 0 {
		wattackNb++
	}
	if (bpawnAttacks & wkingZone) != 0 {
		battackNb++
	}

	var threatScore score
	threatScore.AddN(e.threat,
		PopCount(wpawnAttacks&p.Black&^p.Pawns)-
			PopCount(bpawnAttacks&p.White&^p.Pawns))

	var wMobilityArea = p.Black | (^allPieces & ^bpawnAttacks)
	var bMobilityArea = p.White | (^allPieces & ^wpawnAttacks)

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		wn++
		sq = FirstOne(x)
		pieceScore.Add(e.pstKnight[sq])
		b = KnightAttacks[sq]
		pieceScore.Add(e.knightMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackNb++
			wattackTot += kingAttackUnitKnight
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		bn++
		sq = FirstOne(x)
		pieceScore.Sub(e.pstKnight[sq])
		b = KnightAttacks[sq]
		pieceScore.Sub(e.knightMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackNb++
			battackTot += kingAttackUnitKnight
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		wb++
		sq = FirstOne(x)
		pieceScore.Add(e.pstBishop[sq])
		b = BishopAttacks(sq, allPieces)
		pieceScore.Add(e.bishopMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackNb++
			wattackTot += kingAttackUnitBishop
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		bb++
		sq = FirstOne(x)
		pieceScore.Sub(e.pstBishop[sq])
		b = BishopAttacks(sq, allPieces)
		pieceScore.Sub(e.bishopMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackNb++
			battackTot += kingAttackUnitBishop
		}
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 &&
			((p.Pawns&p.Black&Rank7Mask) != 0 || Rank(bkingSq) == Rank8) {
			pieceScore.Add(e.rook7Th)
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		//b = RookAttacks(sq, allPieces)
		pieceScore.Add(e.rookMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackNb++
			wattackTot += kingAttackUnitRook
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
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		//b = RookAttacks(sq, allPieces)
		pieceScore.Sub(e.rookMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackNb++
			battackTot += kingAttackUnitRook
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
		pieceScore.Add(e.pstQueen[sq])
		b = QueenAttacks(sq, allPieces)
		pieceScore.Add(e.queenMobility[PopCount(b&wMobilityArea)])
		if (b & bkingZone) != 0 {
			wattackNb++
			wattackTot += kingAttackUnitQueen
		}
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		bq++
		sq = FirstOne(x)
		pieceScore.Sub(e.pstQueen[sq])
		b = QueenAttacks(sq, allPieces)
		pieceScore.Sub(e.queenMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackNb++
			battackTot += kingAttackUnitQueen
		}
	}

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = e.pawnPassedBonus[Rank(sq)]
		pawnScore.opening += 4 * bonus
		pawnScore.endgame += 8 * bonus
		keySq = sq + 8
		pawnScore.endgame += (dist[keySq][bkingSq]*2 - dist[keySq][wkingSq]) * bonus
		if (SquareMask[keySq] & p.Black) == 0 {
			pawnScore.endgame += bonus
		}

		if bn+bb+br+bq == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				pawnScore.endgame += 200 * Rank(f1) / Rank7
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = e.pawnPassedBonus[Rank(FlipSquare(sq))]
		pawnScore.opening -= 4 * bonus
		pawnScore.endgame -= 8 * bonus
		keySq = sq - 8
		pawnScore.endgame -= (dist[keySq][wkingSq]*2 - dist[keySq][bkingSq]) * bonus
		if (SquareMask[keySq] & p.White) == 0 {
			pawnScore.endgame -= bonus
		}

		if wn+wb+wr+wq == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				pawnScore.endgame -= 200 * (Rank8 - Rank(f1)) / Rank7
			}
		}
	}

	kingScore.endgame += e.pstKing[wkingSq].endgame
	if (bq > 0 && br+bn+bb > 0) ||
		(br > 1 && bn+bb > 1) {
		kingDanger := 15*shelterWKingSquare(p, wkingSq) +
			10*e.kingOpeningPenalty[wkingSq] +
			2*battackTot*Max(1, battackTot-5)
		kingScore.opening -= kingDanger
	}

	kingScore.endgame -= e.pstKing[FlipSquare(bkingSq)].endgame
	if (wq > 0 && wr+wn+wb > 0) ||
		(wr > 1 && wn+wb > 1) {
		kingDanger := 15*shelterBKingSquare(p, bkingSq) +
			10*e.kingOpeningPenalty[FlipSquare(bkingSq)] +
			2*wattackTot*Max(1, wattackTot-5)
		kingScore.opening += kingDanger
	}

	var materialScore score
	materialScore.AddN(e.materialPawn, wp-bp)
	materialScore.AddN(e.materialKnight, wn-bn)
	materialScore.AddN(e.materialBishop, wb-bb)
	materialScore.AddN(e.materialRook, wr-br)
	materialScore.AddN(e.materialQueen, wq-bq)
	if wb >= 2 {
		materialScore.Add(e.bishopPair)
	}
	if bb >= 2 {
		materialScore.Sub(e.bishopPair)
	}

	var total = pawnScore
	total.Add(pieceScore)
	total.Add(kingScore)
	total.Add(materialScore)
	total.Add(threatScore)

	var phase = wn + bn + wb + bb + 2*(wr+br) + 4*(wq+bq)
	if phase > maxPhase {
		phase = maxPhase
	}
	var result = total.Mix(phase)

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
		fmt.Println("Total:", total)
		fmt.Println("Total Evaluation:", result)
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
	return p.Pawns & p.White &^
		DownFill(Down(Left(p.Pawns&p.Black)|p.Pawns|Right(p.Pawns&p.Black)))
}

func getBlackPassedPawns(p *Position) uint64 {
	return p.Pawns & p.Black &^
		UpFill(Up(Left(p.Pawns&p.White)|p.Pawns|Right(p.Pawns&p.White)))
}

func shelterWKingSquare(p *Position, square int) int {
	var file = File(square)
	if file <= FileC {
		file = FileB
	} else if file >= FileF {
		file = FileG
	}
	var penalty = 0
	for i := 0; i < 3; i++ {
		var mask = FileMask[file+i-1] & p.White & p.Pawns
		if (mask & Rank2Mask) == 0 {
			penalty++
			if (mask & Rank3Mask) == 0 {
				penalty++
				if mask == 0 {
					penalty++
				}
			}
		}
	}
	if penalty == 1 {
		penalty = 0
	}
	return penalty
}

func shelterBKingSquare(p *Position, square int) int {
	var file = File(square)
	if file <= FileC {
		file = FileB
	} else if file >= FileF {
		file = FileG
	}
	var penalty = 0
	for i := 0; i < 3; i++ {
		var mask = FileMask[file+i-1] & p.Black & p.Pawns
		if (mask & Rank7Mask) == 0 {
			penalty++
			if (mask & Rank6Mask) == 0 {
				penalty++
				if mask == 0 {
					penalty++
				}
			}
		}
	}
	if penalty == 1 {
		penalty = 0
	}
	return penalty
}

func BitboardToString(b uint64) string {
	result := ""
	for x := b; x != 0; x &= x - 1 {
		sq := FirstOne(x)
		if result != "" {
			result += ","
		}
		result += SquareName(sq)
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
		kingZone[sq] = SquareMask[sq] | KingAttacks[sq]
	}
}
