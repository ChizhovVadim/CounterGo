package eval

import (
	"errors"

	"github.com/ChizhovVadim/CounterGo/internal/domain"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const totalPhase = 24

var errAddComplexFeature = errors.New("errAddComplexFeature")

type EvaluationService struct {
	tuning        bool
	score         Score
	kingpawnTable []kingPawnEntry
	features      []int
	weights       []Score
	pieceCount    [COLOUR_NB][PIECE_NB]int
	phase         int
	attackedBy    [COLOUR_NB][PIECE_NB]uint64
	kingSq        [COLOUR_NB]int
}

type kingPawnEntry struct {
	wpawns, bpawns uint64
	wking, bking   int
	score          Score
	passed         uint64
}

func NewEvaluationService() *EvaluationService {
	var e = &EvaluationService{
		kingpawnTable: make([]kingPawnEntry, 1<<16),
		features:      make([]int, totalFeatureSize),
		weights:       make([]Score, totalFeatureSize),
	}
	e.initWeights()
	return e
}

func (e *EvaluationService) EnableTuning() {
	e.tuning = true
}

func (e *EvaluationService) initWeights() {
	if 2*len(e.features) != len(w) {
		return
	}
	for i := range e.weights {
		e.weights[i] = S(w[2*i], w[2*i+1])
	}
}

func (e *EvaluationService) Evaluate(p *Position) int {
	e.init(p)

	var pawnKingKey = murmurMix(p.Pawns&p.White,
		murmurMix(p.Pawns&p.Black,
			murmurMix(p.Kings&p.White,
				p.Kings&p.Black)))
	var pke = &e.kingpawnTable[pawnKingKey%uint64(len(e.kingpawnTable))]
	if e.tuning ||
		!(pke.wpawns == p.Pawns&p.White &&
			pke.bpawns == p.Pawns&p.Black &&
			pke.wking == e.kingSq[SideWhite] &&
			pke.bking == e.kingSq[SideBlack]) {
		pke.wpawns = p.Pawns & p.White
		pke.bpawns = p.Pawns & p.Black
		pke.wking = e.kingSq[SideWhite]
		pke.bking = e.kingSq[SideBlack]

		e.evalKingAndPawns(p)
		pke.score = e.score
	} else {
		e.score = pke.score
	}

	e.evalFirstPass(p)

	e.addFeature(fThreatPawn,
		PopCount(p.White&^p.Pawns&e.attackedBy[SideBlack][Pawn])-
			PopCount(p.Black&^p.Pawns&e.attackedBy[SideWhite][Pawn]))

	e.addFeature(fPawnValue, e.pieceCount[SideWhite][Pawn]-e.pieceCount[SideBlack][Pawn])
	e.addFeature(fKnightValue, e.pieceCount[SideWhite][Knight]-e.pieceCount[SideBlack][Knight])
	e.addFeature(fBishopValue, e.pieceCount[SideWhite][Bishop]-e.pieceCount[SideBlack][Bishop])
	e.addFeature(fRookValue, e.pieceCount[SideWhite][Rook]-e.pieceCount[SideBlack][Rook])
	e.addFeature(fQueenValue, e.pieceCount[SideWhite][Queen]-e.pieceCount[SideBlack][Queen])

	if p.WhiteMove {
		e.addFeature(fTempo, 1)
	} else {
		e.addFeature(fTempo, -1)
	}

	var phase = e.pieceCount[SideWhite][Knight] + e.pieceCount[SideBlack][Knight] +
		e.pieceCount[SideWhite][Bishop] + e.pieceCount[SideBlack][Bishop] +
		2*(e.pieceCount[SideWhite][Rook]+e.pieceCount[SideBlack][Rook]) +
		4*(e.pieceCount[SideWhite][Queen]+e.pieceCount[SideBlack][Queen])
	if phase > totalPhase {
		phase = totalPhase
	}
	e.phase = phase

	var result = (e.score.Mg()*phase + e.score.Eg()*(totalPhase-phase)) / (totalPhase * 100)
	var strongSide int
	if result > 0 {
		strongSide = SideWhite
	} else {
		strongSide = SideBlack
	}
	result = result * e.computeFactor(strongSide, p) / scaleNormal

	if !p.WhiteMove {
		result = -result
	}

	return result
}

func (e *EvaluationService) init(p *Position) {
	e.score = S(0, 0)

	for pt := Pawn; pt <= King; pt++ {
		e.pieceCount[SideWhite][pt] = 0
		e.pieceCount[SideBlack][pt] = 0
	}

	e.pieceCount[SideWhite][Pawn] = PopCount(p.Pawns & p.White)
	e.pieceCount[SideBlack][Pawn] = PopCount(p.Pawns & p.Black)

	e.kingSq[SideWhite] = FirstOne(p.Kings & p.White)
	e.kingSq[SideBlack] = FirstOne(p.Kings & p.Black)

	e.attackedBy[SideWhite][Pawn] = AllWhitePawnAttacks(p.Pawns & p.White)
	e.attackedBy[SideBlack][Pawn] = AllBlackPawnAttacks(p.Pawns & p.Black)
}

func (e *EvaluationService) evalKingAndPawns(p *Position) {
	var x uint64
	var sq int

	for side := SideWhite; side <= SideBlack; side++ {
		var sign int
		var forward int
		if side == SideWhite {
			sign = 1
			forward = 8
		} else {
			sign = -1
			forward = -8
		}
		var US = side
		var THEM = side ^ 1
		var friendlyPawns = p.Colours(US) & p.Pawns
		var enemyPawns = p.Colours(THEM) & p.Pawns

		for x = friendlyPawns; x != 0; x &= x - 1 {
			sq = FirstOne(x)
			e.addComplexFeature(fPawnPST, relativeSq32(side, sq), sign)

			if PawnAttacksNew(THEM, sq)&friendlyPawns != 0 {
				e.addComplexFeature(fPawnProtected, relativeSq32(side, sq), sign)
			}
			if adjacentFilesMask[File(sq)]&ranks[Rank(sq)]&friendlyPawns != 0 {
				e.addComplexFeature(fPawnDuo, relativeSq32(side, sq), sign)
			}

			if adjacentFilesMask[File(sq)]&friendlyPawns == 0 {
				e.addFeature(fPawnIsolated, sign)
			}
			if FileMask[File(sq)]&^SquareMask[sq]&friendlyPawns != 0 {
				e.addFeature(fPawnDoubled, sign)
			}

			var stoppers = enemyPawns & passedPawnMasks[side][sq]
			// passed pawn
			if stoppers == 0 && upperRankMasks[US][Rank(sq)]&FileMask[File(sq)]&p.Pawns == 0 {
				var r = Max(0, relativeRankOf(side, sq)-Rank3)
				e.addComplexFeature(fPassedPawn, r, sign)
				var keySq = sq + forward
				e.addComplexFeature(fPassedEnemyKing, 8*r+distanceBetween[keySq][e.kingSq[THEM]], sign)
				e.addComplexFeature(fPassedOwnKing, 8*r+distanceBetween[keySq][e.kingSq[US]], sign)
			}
		}

		{
			// KING
			sq = e.kingSq[US]
			e.addComplexFeature(fKingPST, relativeSq32(side, sq), sign)

			for x = kingShieldMasks[US][sq] & friendlyPawns; x != 0; x &= x - 1 {
				var sq = FirstOne(x)
				e.addPst12(fKingShield, side, sq, sign)
			}

			/*for file := Max(FileA, File(sq)-1); file <= Min(FileH, File(sq)+1); file++ {
				var ours = friendly & p.Pawns & FileMask[file] & forwardRanksMasks[US][Rank(sq)]
				var ourDist int
				if ours == 0 {
					ourDist = 7
				} else {
					ourDist = Rank(sq) - Rank(Backmost(US, ours))
					if ourDist < 0 {
						ourDist = -ourDist
					}
				}
				e.addComplexFeature(fKingShield, 8*file+ourDist, sign)
			}*/
		}
	}
}

func (e *EvaluationService) evalFirstPass(p *Position) {
	var x uint64
	var sq int

	for side := SideWhite; side <= SideBlack; side++ {
		var sign int
		if side == SideWhite {
			sign = 1
		} else {
			sign = -1
		}
		var US = side
		var friendly = p.Colours(US)

		for x = p.Knights & friendly; x != 0; x &= x - 1 {
			sq = FirstOne(x)
			e.pieceCount[US][Knight]++
			e.addComplexFeature(fKnightPST, relativeSq32(side, sq), sign)
		}

		for x = p.Bishops & friendly; x != 0; x &= x - 1 {
			sq = FirstOne(x)
			e.pieceCount[US][Bishop]++
			e.addComplexFeature(fBishopPST, relativeSq32(side, sq), sign)
		}

		for x = p.Rooks & friendly; x != 0; x &= x - 1 {
			sq = FirstOne(x)
			e.pieceCount[US][Rook]++
			e.addComplexFeature(fRookPST, relativeSq32(side, sq), sign)
		}

		for x = p.Queens & friendly; x != 0; x &= x - 1 {
			sq = FirstOne(x)
			e.pieceCount[US][Queen]++
			e.addComplexFeature(fQueenPST, relativeSq32(side, sq), sign)
		}

		if e.pieceCount[US][Bishop] >= 2 {
			e.addFeature(fBishopPair, sign)
		}
	}
}

const (
	scaleNormal = 128
)

const (
	QueenSideBB = FileAMask | FileBMask | FileCMask | FileDMask
	KingSideBB  = FileEMask | FileFMask | FileGMask | FileHMask
)

func (e *EvaluationService) computeFactor(own int, p *Position) int {
	var them = own ^ 1
	var ownPawns = e.pieceCount[own][Pawn]
	if ownPawns <= 1 {
		var ownForce = computeForce(e, own)
		var theirForce = computeForce(e, own^1)
		if ownPawns == 0 {
			if ownForce <= 4 {
				return scaleNormal * 1 / 16
			}
			if ownForce-theirForce <= 4 {
				return scaleNormal * 1 / 4
			}
		} else if ownPawns == 1 {
			var theirMinor = e.pieceCount[them][Knight]+e.pieceCount[them][Bishop] != 0
			if ownForce <= 4 && theirMinor {
				return scaleNormal * 1 / 8
			}
			if ownForce == theirForce && theirMinor {
				return scaleNormal * 1 / 2
			}
		}
	}
	var strong = p.Colours(own)

	var strongPawnCount = e.pieceCount[own][Pawn]
	var x = 8 - strongPawnCount
	var pawnScale = 128 - x*x

	if strong&p.Pawns&QueenSideBB == 0 ||
		strong&p.Pawns&KingSideBB == 0 {
		pawnScale -= 20
	}

	//var pawnScale = scaleNormal

	if e.pieceCount[SideWhite][Bishop] == 1 &&
		e.pieceCount[SideBlack][Bishop] == 1 &&
		onlyOne(p.Bishops&darkSquares) {
		if p.Knights|p.Rooks|p.Queens == 0 {
			pawnScale = Min(pawnScale, scaleNormal*1/2)
		}
	}

	return pawnScale
}

func computeForce(e *EvaluationService, side int) int {
	return 4*(e.pieceCount[side][Knight]+e.pieceCount[side][Bishop]) +
		6*e.pieceCount[side][Rook] +
		12*e.pieceCount[side][Queen]
}

func (e *EvaluationService) StartingWeights() []float64 {
	var material = []float64{100, 100, 325, 325, 325, 325, 500, 500, 1000, 1000}
	var result = make([]float64, 2*totalFeatureSize)
	copy(result, material)
	return result
}

func (e *EvaluationService) ComputeFeatures(pos *Position) domain.TuneEntry {
	for i := range e.features {
		e.features[i] = 0
	}
	e.Evaluate(pos)
	var size int
	for _, v := range e.features {
		if v != 0 {
			size++
		}
	}
	var features = make([]domain.FeatureInfo, 0, size)
	for i, v := range e.features {
		if v != 0 {
			features = append(features, domain.FeatureInfo{Index: int16(i), Value: int16(v)})
		}
	}
	var result = domain.TuneEntry{
		Features:         features,
		MgPhase:          float32(e.phase) / totalPhase,
		WhiteStrongScale: float32(e.computeFactor(SideWhite, pos)) / scaleNormal,
		BlackStrongScale: float32(e.computeFactor(SideBlack, pos)) / scaleNormal,
	}
	result.EgPhase = 1 - result.MgPhase
	return result
}

func (e *EvaluationService) addPst12(feature, side, sq, value int) {
	e.addComplexFeature(feature, file4(sq), value)
	e.addComplexFeature(feature, 4+relativeRankOf(side, sq), value)
}

func (e *EvaluationService) addFeature(feature, value int) {
	e.addComplexFeature(feature, 0, value)
}

func (e *EvaluationService) addComplexFeature(feature, featureIndex, value int) {
	var info = &infos[feature]
	var index = info.StartIndex + featureIndex
	var w = e.weights[index]
	//e.score.mg += value * w.mg
	//e.score.eg += value * w.eg
	e.score += Score(value) * w
	if e.tuning {
		if featureIndex >= info.Size {
			panic(errAddComplexFeature)
		}
		e.features[index] += value
	}
}
