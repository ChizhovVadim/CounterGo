package eval

import (
	"math"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	sideWhite = 0
	sideBlack = 1
)

type Score struct {
	Mg int
	Eg int
}

func (s *Score) add(v Score) {
	s.Mg += v.Mg
	s.Eg += v.Eg
}

func (s *Score) sub(v Score) {
	s.Mg -= v.Mg
	s.Eg -= v.Eg
}

func (s *Score) addN(v Score, n int) {
	s.Mg += v.Mg * n
	s.Eg += v.Eg * n
}

type Weights struct {
	PawnMaterial          Score
	KnightMaterial        Score
	BishopMaterial        Score
	RookMaterial          Score
	QueenMaterial         Score
	BishopPairMaterial    Score
	PawnWeak              Score
	PawnDoubled           Score
	PawnDuo               Score
	PawnProtected         Score
	PawnPassed            [8]Score
	PawnPassedFree        [8]Score
	PawnPassedOppKing     [8]Score
	PawnPassedOwnKing     [8]Score
	PawnPassedSquare      Score
	ThreatPawn            Score
	ThreatForPawn         Score
	ThreatPiece           Score
	ThreatPieceForQueen   Score
	Rook7th               Score
	RookOpen              Score
	RookSemiopen          Score
	KingShelter           Score
	KingAttack            [4]Score
	KingQueenTropism      Score
	BishopRammedPawns     Score
	MinorProtected        Score
	KnightOutpost         Score
	PawnBlockedByOwnPiece Score
	PawnRammed            Score
	Tempo                 Score
	KnightMobility        [9]Score
	BishopMobility        [14]Score
	RookMobility          [15]Score
	QueenMobility         [28]Score
	PST                   [2][8][64]Score `json:"-"`
}

func (w *Weights) init() {
	// Error: 0.055735
	var autoGeneratedWeights = []int{97, 363, 320, 394, 363, 516, 619, 1368, 1180, 40, 60, 8, 11, -11, -2, -9, 12, 9, 3, 12, 9, 11, 12, 12, 19, 69, 6, -7, 24, 25, 21, 65, -8, -16, -9, -12, -11, 8, 2, 13, 12, 1, 12, 51, -10, 16, -3, 14, 43, 158, -3, -31, -8, -16, 7, 23, 29, 2, -22, 5, 1, -16, 33, 40, 50, 22}
	w.Apply(autoGeneratedWeights)
}

func (w *Weights) Apply(weights []int) []int {
	var wh = &weightHolder{weights: weights, index: 0}

	w.PawnMaterial = Score{wh.withDefault(100).next(), 100}
	w.KnightMaterial = Score{wh.withDefault(325).next(), wh.withDefault(325).next()}
	w.BishopMaterial = Score{wh.withDefault(325).next(), wh.withDefault(325).next()}
	w.RookMaterial = Score{wh.withDefault(500).next(), wh.withDefault(500).next()}
	w.QueenMaterial = Score{wh.withDefault(1000).next(), wh.withDefault(1000).next()}
	w.BishopPairMaterial = Score{wh.withDefault(50).next(), wh.withDefault(50).next()}

	var knightPst = wh.nextScore()
	var queenPst = wh.nextScore()
	var kingPst = wh.nextScore()
	var (
		knightLine = [8]int{0, 2, 3, 4, 4, 3, 2, 0}
		bishopLine = [8]int{0, 1, 2, 3, 3, 2, 1, 0}
		kingLine   = [8]int{0, 2, 3, 4, 4, 3, 2, 0}
	)

	for sq := 0; sq < 64; sq++ {
		var f = File(sq)
		var r = Rank(sq)
		w.PST[sideWhite][Knight][sq] = Score{
			Mg: (knightLine[f] + knightLine[r]) * knightPst.Mg,
			Eg: (knightLine[f] + knightLine[r]) * knightPst.Eg,
		}
		w.PST[sideWhite][Queen][sq] = Score{
			Mg: Min(bishopLine[f], bishopLine[r]) * queenPst.Mg,
			Eg: Min(bishopLine[f], bishopLine[r]) * queenPst.Eg,
		}
		w.PST[sideWhite][King][sq] = Score{
			Mg: (Min(dist[sq][SquareG1], dist[sq][SquareB1])) * kingPst.Mg,
			Eg: (kingLine[f] + kingLine[r]) * kingPst.Eg,
		}
	}

	for pieceType := Pawn; pieceType <= King; pieceType++ {
		for sq := 0; sq < 64; sq++ {
			w.PST[sideBlack][pieceType][sq] = negScore(w.PST[sideWhite][pieceType][FlipSquare(sq)])
		}
	}

	initMobility(w.KnightMobility[:], wh.nextScore())
	initMobility(w.BishopMobility[:], wh.nextScore())
	initMobility(w.RookMobility[:], wh.nextScore())
	initMobility(w.QueenMobility[:], wh.nextScore())

	w.ThreatPawn = wh.nextScore()
	w.ThreatForPawn = wh.nextScore()
	w.ThreatPiece = wh.nextScore()
	w.ThreatPieceForQueen = wh.nextScore()

	w.PawnWeak = wh.nextScore()
	w.PawnDoubled = wh.nextScore()
	w.PawnDuo = wh.nextScore()
	w.PawnProtected = wh.nextScore()
	w.Rook7th = wh.nextScore()
	w.RookOpen = wh.nextScore()
	w.RookSemiopen = wh.nextScore()
	w.KingShelter = Score{wh.next(), 0}

	for i := 2; i < len(w.KingAttack); i++ {
		w.KingAttack[i] = Score{wh.next(), 0}
	}

	w.KingQueenTropism = wh.nextScore()
	w.BishopRammedPawns = wh.nextScore()
	w.MinorProtected = wh.nextScore()
	w.KnightOutpost = wh.nextScore()
	w.PawnBlockedByOwnPiece = wh.nextScore()
	w.PawnRammed = wh.nextScore()

	var pawnPassed = wh.nextScore()
	var pawnPassedFree = wh.next()
	var pawnPassedOppKing = wh.next()
	var pawnPassedBonus = [8]float64{0, 0.1, 0.13, 0.16, 0.28, 0.68, 1.0, 0}
	//var pawnPassedBonus = [8]int{0, 0, 0, 2, 6, 12, 21, 0}
	var pawnPassedStepPrice = mean(pawnPassedBonus[1:7])
	for i := 0; i < 8; i++ {
		var r = pawnPassedBonus[i] / pawnPassedStepPrice
		w.PawnPassed[i] = makeScore(float64(pawnPassed.Mg)*r,
			float64(pawnPassed.Eg)*r)
		if i >= Rank4 {
			w.PawnPassedFree[i] = makeScore(0,
				float64(pawnPassedFree)*r)
			w.PawnPassedOppKing[i] = makeScore(0,
				float64(pawnPassedOppKing)*r)
			w.PawnPassedOwnKing[i] = makeScore(0,
				float64(pawnPassedOppKing)*r/-2.5)
		}
	}
	w.PawnPassedSquare = Score{0, 33}
	w.Tempo = Score{8, 8}

	return wh.weights
}

func mean(source []float64) float64 {
	var sum = 0.0
	var count = 0
	for _, x := range source {
		sum += x
		count++
	}
	return sum / float64(count)
}

func negScore(s Score) Score {
	return Score{-s.Mg, -s.Eg}
}

func makeScore(mg, eg float64) Score {
	return Score{
		Mg: int(math.Round(mg)),
		Eg: int(math.Round(eg)),
	}
}

func initMobility(source []Score, weight Score) {
	for i := range source {
		var k = math.Sqrt(float64(i) / float64(len(source)-1))
		source[i] = makeScore(k*float64(weight.Mg)*float64(weight.Mg),
			k*float64(weight.Eg)*float64(weight.Eg))
	}
}

type weightHolder struct {
	weights []int
	index   int
}

func (wh *weightHolder) withDefault(v int) *weightHolder {
	if wh.index >= len(wh.weights) {
		wh.weights = append(wh.weights, v)
	}
	return wh
}

func (wh *weightHolder) next() int {
	if wh.index >= len(wh.weights) {
		wh.weights = append(wh.weights, 0)
	}
	var value = wh.weights[wh.index]
	wh.index++
	return value
}

func (wh *weightHolder) nextScore() Score {
	return Score{wh.next(), wh.next()}
}
