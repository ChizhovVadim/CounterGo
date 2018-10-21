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

func (s score) multK(k float64) score {
	return score{
		int(float64(s.opening) * k),
		int(float64(s.endgame) * k),
	}
}

func (s *score) Mix(phase int) int {
	return (s.opening*phase + s.endgame*(maxPhase-phase)) / maxPhase
}

func (s score) Neg() score {
	return score{-s.opening, -s.endgame}
}

type EvaluationService struct {
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
	rook7Th             score
	rookOpen            score
	rookSemiopen        score
	pawnIsolated        score
	pawnDoubled         score
	pawnDoubledIsolated score
	pawnConnected       score
	kingShelter         int
	kingAttack          int
	threat              score
	sideToMove          score
	pst                 pst
}

type pst struct {
	wn, bn, wb, bb, wq, bq, wk, bk [64]score
}

func NewEvaluationService() *EvaluationService {
	var srv = &EvaluationService{}
	srv.Init(srv.DefaultParams())
	return srv
}

func (srv *EvaluationService) DefaultParams() []int {
	// Error: 0.042270
	return []int{108, 356, 330, 384, 349, 519, 633, 924, 1297, 27, 35, 75, 0, 23, 15, 0, 19, 39, 0, 15, 7, -9, 0, 0, 0, -6, -2, 8, 14, 9, 11, 9, 10, 7, 4, 0, 1, 9, 8, 6, 0, 7, 6, 3, 7, 2, 6}
}

type evalValues struct {
	params []int
	index  int
}

func (ev *evalValues) Next() int {
	return ev.NextWithDefault(0)
}

func (ev *evalValues) NextWithDefault(defaultVal int) int {
	if len(ev.params) <= ev.index {
		ev.params = append(ev.params, defaultVal)
	}
	var value = ev.params[ev.index]
	ev.index++
	return value
}

func (ev *evalValues) NextScore() score {
	return score{ev.Next(), ev.Next()}
}

func limitValue(v, lower, upper int) int {
	if v < lower {
		return lower
	}
	if upper < v {
		return upper
	}
	return v
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func regularizationScore(s score) int {
	return absInt(s.opening) + absInt(s.endgame)
}

func regularizationSlice(source []score) int {
	var low = source[0]
	var high = low
	for i := range source {
		low.opening = Min(low.opening, source[i].opening)
		low.endgame = Min(low.endgame, source[i].endgame)
		high.opening = Max(high.opening, source[i].opening)
		high.endgame = Max(high.endgame, source[i].endgame)
	}
	return high.opening - low.opening + high.endgame - low.endgame
}

// should be coordinated with method Init
func (e *EvaluationService) Regularization() int {
	return regularizationScore(e.bishopPair) +
		regularizationScore(e.sideToMove) +
		regularizationScore(e.threat) +
		regularizationScore(e.rook7Th) +
		regularizationScore(e.rookOpen) +
		regularizationScore(e.rookSemiopen) +
		regularizationScore(e.pawnIsolated) +
		regularizationScore(e.pawnConnected) +
		regularizationScore(e.pawnDoubled) +
		regularizationScore(e.pawnDoubledIsolated) +
		//regularizationScore(e.kingShelter) +
		regularizationSlice(e.knightMobility[:]) +
		regularizationSlice(e.bishopMobility[:]) +
		regularizationSlice(e.rookMobility[:]) +
		regularizationSlice(e.queenMobility[:]) +
		regularizationSlice(e.pst.wn[:]) +
		regularizationSlice(e.pst.wb[:]) +
		regularizationSlice(e.pst.wq[:]) +
		regularizationSlice(e.pst.wk[:]) +
		absInt(e.materialPawn.opening) +
		regularizationScore(e.materialKnight) +
		regularizationScore(e.materialBishop) +
		regularizationScore(e.materialRook) +
		regularizationScore(e.materialQueen)
}

var (
	knightLine      = [8]int{0, 2, 3, 4, 4, 3, 2, 0}
	bishopLine      = [8]int{0, 1, 2, 3, 3, 2, 1, 0}
	kingLine        = [8]int{0, 2, 3, 4, 4, 3, 2, 0}
	kingFile        = [8]int{3, 4, 2, 0, 0, 2, 4, 3}
	kingRank        = [8]int{3, 2, 1, 0, 0, 0, 0, 0}
	pawnPassedBonus = [8]int{0, 0, 0, 2, 6, 12, 21, 0}
)

func (e *EvaluationService) Init(params []int) []int {
	var ev = evalValues{params, 0}

	e.materialPawn = score{ev.NextWithDefault(100), 100}
	e.materialKnight = score{ev.NextWithDefault(325), ev.NextWithDefault(325)}
	e.materialBishop = score{ev.NextWithDefault(325), ev.NextWithDefault(325)}
	e.materialRook = score{ev.NextWithDefault(500), ev.NextWithDefault(500)}
	e.materialQueen = score{ev.NextWithDefault(1000), ev.NextWithDefault(1000)}

	e.bishopPair = ev.NextScore()
	e.threat = ev.NextScore()
	e.sideToMove = ev.NextScore()
	e.rook7Th = ev.NextScore()
	e.rookOpen = ev.NextScore()
	e.rookSemiopen = ev.NextScore()
	e.pawnIsolated = ev.NextScore()
	e.pawnDoubled = ev.NextScore()
	e.pawnDoubledIsolated = ev.NextScore()
	e.pawnConnected = ev.NextScore()
	e.kingShelter = ev.Next()
	e.kingAttack = 25 * ev.Next()

	var knightCenter = ev.NextScore()
	var bishopCenter = ev.NextScore()
	var queenCenter = ev.NextScore()
	var kingCenter = ev.NextScore()

	for sq := 0; sq < 64; sq++ {
		var f = File(sq)
		var r = Rank(sq)
		e.pst.wn[sq] = knightCenter.multK(float64(knightLine[f] + knightLine[r]))
		e.pst.wb[sq] = bishopCenter.multK(float64(Min(bishopLine[f], bishopLine[r])))
		e.pst.wq[sq] = queenCenter.multK(float64(Min(bishopLine[f], bishopLine[r])))
		e.pst.wk[sq] = score{
			kingCenter.opening * (kingFile[f] + kingRank[r]),
			kingCenter.endgame * (kingLine[f] + kingLine[r]),
		}
	}
	e.pst.initBlack()

	initMobility(e.knightMobility[:], ev.NextScore())
	initMobility(e.bishopMobility[:], ev.NextScore())
	initMobility(e.rookMobility[:], ev.NextScore())
	initMobility(e.queenMobility[:], ev.NextScore())

	return ev.params
}

func initMobility(source []score, s score) {
	var total = s.multK(float64(len(source) - 1))
	for i := range source {
		var x = math.Sqrt(float64(i) / float64(len(source)-1))
		source[i] = total.multK(x)
	}
}

func NewExperimentEvaluationService() *EvaluationService {
	var srv = NewEvaluationService()
	srv.experimentSettings = true
	return srv
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

func (e *EvaluationService) Evaluate(p *Position) int {
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

	pawnScore.AddN(e.pawnConnected,
		PopCount(getConnectedPawns(p.Pawns&p.White))-PopCount(getConnectedPawns(p.Pawns&p.Black)))

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
		pieceScore.Add(e.pst.wn[sq])
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
		pieceScore.Add(e.pst.bn[sq])
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
		pieceScore.Add(e.pst.wb[sq])
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
		pieceScore.Add(e.pst.bb[sq])
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
		pieceScore.Add(e.pst.wq[sq])
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
		pieceScore.Add(e.pst.bq[sq])
		b = QueenAttacks(sq, allPieces)
		pieceScore.Sub(e.queenMobility[PopCount(b&bMobilityArea)])
		if (b & wkingZone) != 0 {
			battackNb++
			battackTot += kingAttackUnitQueen
		}
	}

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(sq)]
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
		bonus = pawnPassedBonus[Rank(FlipSquare(sq))]
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

	kingScore.Add(e.pst.wk[wkingSq])
	kingScore.Add(e.pst.bk[bkingSq])

	if (bq > 0 && br+bn+bb > 0) ||
		(br > 1 && bn+bb > 1) {
		kingScore.opening -= e.kingShelter*shelterWKingSquare(p, wkingSq) +
			e.kingAttack*battackTot*Max(1, battackTot-5)/150
	}

	if (wq > 0 && wr+wn+wb > 0) ||
		(wr > 1 && wn+wb > 1) {
		kingScore.opening += e.kingShelter*shelterBKingSquare(p, bkingSq) +
			e.kingAttack*wattackTot*Max(1, wattackTot-5)/150
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

	if p.WhiteMove {
		total.Add(e.sideToMove)
	} else {
		total.Sub(e.sideToMove)
	}

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

func getConnectedPawns(pawns uint64) uint64 {
	var wings = Left(pawns) | Right(pawns)
	return pawns & (wings | Up(wings) | Down(wings))
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

func (pst *pst) initBlack() {
	for sq := 0; sq < 64; sq++ {
		var flipSq = FlipSquare(sq)
		pst.bn[sq] = pst.wn[flipSq].Neg()
		pst.bb[sq] = pst.wb[flipSq].Neg()
		pst.bq[sq] = pst.wq[flipSq].Neg()
		pst.bk[sq] = pst.wk[flipSq].Neg()
	}
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
