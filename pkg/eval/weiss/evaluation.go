package eval

import (
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

// Weiss 2.0 eval. CCRL 40/15 3220 ELO.
// From here: https://github.com/TerjeKir/weiss
type EvaluationService struct {
	Tuning bool
	Weights
	pawnKingTable      []pawnKingEntry
	occupied           uint64
	passedPawns        uint64
	pawnAttacks        [COLOUR_NB]uint64
	pieceCount         [COLOUR_NB][PIECE_NB]int
	kingSquare         [COLOUR_NB]int
	mobilityAreas      [COLOUR_NB]uint64
	kingAreas          [COLOUR_NB]uint64
	kingAttackPower    [COLOUR_NB]int
	kingAttackersCount [COLOUR_NB]int
}

type pawnKingEntry struct {
	pawns       [COLOUR_NB]uint64
	kingSquare  [COLOUR_NB]int
	eval        Score
	passedPawns uint64
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.pawnKingTable = make([]pawnKingEntry, 1<<16)
	es.Weights.init()
	return es
}

func (e *EvaluationService) EnableTuning() *EvaluationService {
	e.Tuning = true
	return e
}

func (e *EvaluationService) Evaluate(p *Position) int {
	var eval Score

	initEval(e, p)

	var pawnKingKey = murmurMix(p.Pawns&p.White,
		murmurMix(p.Pawns&p.Black,
			murmurMix(p.Kings&p.White,
				p.Kings&p.Black)))
	var pke = &e.pawnKingTable[pawnKingKey%uint64(len(e.pawnKingTable))]
	if e.Tuning ||
		!(pke.pawns[SideWhite] == p.Pawns&p.White &&
			pke.pawns[SideBlack] == p.Pawns&p.Black &&
			pke.kingSquare[SideWhite] == e.kingSquare[SideWhite] &&
			pke.kingSquare[SideBlack] == e.kingSquare[SideBlack]) {
		pke.pawns[SideWhite] = p.Pawns & p.White
		pke.pawns[SideBlack] = p.Pawns & p.Black
		pke.kingSquare[SideWhite] = e.kingSquare[SideWhite]
		pke.kingSquare[SideBlack] = e.kingSquare[SideBlack]
		pke.passedPawns = 0
		pke.eval = evaluateKingPawns(pke, &e.Weights, SideWhite) -
			evaluateKingPawns(pke, &e.Weights, SideBlack)
	}
	eval += pke.eval
	e.passedPawns = pke.passedPawns

	eval += e.NBBehindPawn * Score(
		PopCount((p.Knights|p.Bishops)&p.White&Down(p.Pawns))-
			PopCount((p.Knights|p.Bishops)&p.Black&Up(p.Pawns)))

	eval += evaluatePieces(e, p, SideWhite) - evaluatePieces(e, p, SideBlack)

	eval += evaluateKings(e, p, SideWhite) - evaluateKings(e, p, SideBlack)
	eval += evaluatePassed(e, p, SideWhite) - evaluatePassed(e, p, SideBlack)
	eval += evaluateThreats(e, p, SideWhite) - evaluateThreats(e, p, SideBlack)

	eval += e.PawnValue * Score(e.pieceCount[SideWhite][Pawn]-e.pieceCount[SideBlack][Pawn])
	eval += e.KnightValue * Score(e.pieceCount[SideWhite][Knight]-e.pieceCount[SideBlack][Knight])
	eval += e.BishopValue * Score(e.pieceCount[SideWhite][Bishop]-e.pieceCount[SideBlack][Bishop])
	eval += e.RookValue * Score(e.pieceCount[SideWhite][Rook]-e.pieceCount[SideBlack][Rook])
	eval += e.QueenValue * Score(e.pieceCount[SideWhite][Queen]-e.pieceCount[SideBlack][Queen])
	if e.pieceCount[SideWhite][Bishop] >= 2 {
		eval += e.BishopPair
	}
	if e.pieceCount[SideBlack][Bishop] >= 2 {
		eval -= e.BishopPair
	}

	var factor = computeFactor(e, p, eval)

	var phase = 4*(e.pieceCount[SideWhite][Queen]+e.pieceCount[SideBlack][Queen]) +
		2*(e.pieceCount[SideWhite][Rook]+e.pieceCount[SideBlack][Rook]) +
		1*(e.pieceCount[SideWhite][Knight]+e.pieceCount[SideBlack][Knight]+
			e.pieceCount[SideWhite][Bishop]+e.pieceCount[SideBlack][Bishop])
	//TODO Weiss full compability. I think better phase = Min(phase, 24)
	phase = (phase*256 + 12) / 24

	var result = (eval.Middle()*phase +
		eval.End()*(256-phase)*factor/scaleFactorNormal) / 256

	if !p.WhiteMove {
		result = -result
	}

	const Tempo = 15

	return result + Tempo
}

const scaleFactorNormal = 128

const (
	QueenSideBB = FileAMask | FileBMask | FileCMask | FileDMask
	KingSideBB  = FileEMask | FileFMask | FileGMask | FileHMask
)

func computeFactor(e *EvaluationService, p *Position, eval Score) int {
	var strongSide int
	var strong uint64
	if eval.End() > 0 {
		strongSide = SideWhite
		strong = p.White
	} else {
		strongSide = SideBlack
		strong = p.Black
	}

	var strongPawnCount = e.pieceCount[strongSide][Pawn]
	var x = 8 - strongPawnCount
	var pawnScale = 128 - x*x

	if strong&p.Pawns&QueenSideBB == 0 ||
		strong&p.Pawns&KingSideBB == 0 {
		pawnScale -= 20
	}

	if e.pieceCount[SideWhite][Bishop] == 1 &&
		e.pieceCount[SideBlack][Bishop] == 1 &&
		OnlyOne(p.Bishops&darkSquares) {

		//TODO Weiss full compability. I think Q+Q VS N+N should not ocb scale

		var whiteNonPawnCount = PopCount(p.White &^ (p.Pawns | p.Kings))
		var blackNonPawnCount = PopCount(p.Black &^ (p.Pawns | p.Kings))
		if whiteNonPawnCount == blackNonPawnCount &&
			whiteNonPawnCount <= 2 &&
			blackNonPawnCount <= 2 {
			//TODO Weiss full compability. whiteNonPawnCount <= 1??
			if whiteNonPawnCount == 1 {
				pawnScale = Min(pawnScale, 64)
			} else {
				pawnScale = Min(pawnScale, 96)
			}
		}
	}

	return pawnScale
}

func initEval(e *EvaluationService, p *Position) {
	e.kingAttackPower[SideWhite] = -30
	e.kingAttackPower[SideBlack] = -30

	e.kingAttackersCount[SideWhite] = 0
	e.kingAttackersCount[SideBlack] = 0

	var occ = p.AllPieces()
	e.occupied = occ
	e.passedPawns = 0

	for pt := Pawn; pt <= King; pt++ {
		e.pieceCount[SideWhite][pt] = 0
		e.pieceCount[SideBlack][pt] = 0
	}

	e.kingSquare[SideWhite] = p.KingSq(true)
	e.kingSquare[SideBlack] = p.KingSq(false)

	e.pawnAttacks[SideWhite] = AllWhitePawnAttacks(p.Pawns & p.White)
	e.pawnAttacks[SideBlack] = AllBlackPawnAttacks(p.Pawns & p.Black)

	e.pieceCount[SideWhite][Pawn] = PopCount(p.Pawns & p.White)
	e.pieceCount[SideBlack][Pawn] = PopCount(p.Pawns & p.Black)

	e.mobilityAreas[SideWhite] = ^(e.pawnAttacks[SideBlack] | p.Pawns&p.White&(Rank2Mask|Down(occ)))
	e.mobilityAreas[SideBlack] = ^(e.pawnAttacks[SideWhite] | p.Pawns&p.Black&(Rank7Mask|Up(occ)))

	e.kingAreas[SideWhite] = KingAttacks[e.kingSquare[SideWhite]]
	e.kingAreas[SideBlack] = KingAttacks[e.kingSquare[SideBlack]]
}

func evaluateKingPawns(e *pawnKingEntry, w *Weights, colour int) Score {
	var US, THEM = colour, colour ^ 1

	var sq int
	var eval Score

	var myPawns = e.pawns[US]
	var enemyPawns = e.pawns[THEM]
	var kingSq = e.kingSquare[US]

	var forward int
	if colour == SideWhite {
		forward = 8
	} else {
		forward = -8
	}

	for temp := myPawns; temp != 0; temp &= temp - 1 {
		sq = FirstOne(temp)
		eval += w.PSQT[US][Pawn][sq]

		var neighbors = myPawns & adjacentFilesMasks[File(sq)]
		var stoppers = enemyPawns & passedPawnMasks[US][sq]
		var support = myPawns & PawnAttacksNew(THEM, sq)

		// passed pawn
		if stoppers == 0 {
			var rank = RelativeRankOf(US, sq)

			eval += w.PassedPawn[rank]
			//TODO Weiss full compability. bug here, better:
			//if support != 0 {
			if support&temp != 0 {
				eval += w.PassedDefended[rank]
			}

			var keySq = sq + forward

			if rank > Rank3 {
				var dist = distanceBetween[keySq][e.kingSquare[US]]
				eval += Score(dist) * w.PassedDistUs[rank]

				dist = distanceBetween[keySq][e.kingSquare[THEM]]
				eval += Score(dist*(rank-Rank3)) * w.PassedDistThem
			}

			e.passedPawns |= SquareMask[sq]
		}

		if neighbors == 0 {
			eval += w.PawnIsolated
		}

		if PawnAttacksNew(US, sq)&myPawns != 0 {
			eval += w.PawnSupport
		}

		if SquareMask[sq+forward]&myPawns != 0 {
			eval += w.PawnDoubled
		}

		if forwardFileMasks[US][sq]&enemyPawns == 0 && support == 0 {
			eval += w.PawnOpen
		}

		if Left(SquareMask[sq])&myPawns != 0 {
			eval += w.PawnPhalanx[RelativeRankOf(US, sq)]
		}
	}

	eval += w.PSQT[US][King][kingSq]

	if KingAttacks[kingSq]&enemyPawns != 0 {
		eval += w.KingAtkPawn
	}

	return eval
}

func evaluatePieces(e *EvaluationService, p *Position, colour int) Score {
	var US, THEM = colour, colour ^ 1

	var sq int
	var eval Score
	var attacks uint64

	var friendly = p.Colours(US)
	var myPawns = p.Pawns & friendly
	var enemyPawns = p.Pawns & p.Colours(THEM)

	for temp := p.Knights & friendly; temp != 0; temp &= temp - 1 {
		e.pieceCount[US][Knight]++
		sq = FirstOne(temp)
		eval += e.PSQT[US][Knight][sq]

		attacks = KnightAttacks[sq]

		eval += e.KnightMobility[PopCount(e.mobilityAreas[US]&attacks)]

		var kingAttacks = attacks & e.kingAreas[THEM] & e.mobilityAreas[US]
		//TODO Weiss full compability. I think for all pieces N/B/R/Q better checks &^= friendly
		var checks = attacks & KnightAttacks[e.kingSquare[THEM]] & e.mobilityAreas[US]
		if kingAttacks|checks != 0 {
			e.kingAttackPower[THEM] += e.SafetyAttackPower[Knight]*PopCount(kingAttacks) +
				e.SafetyCheckPower[Knight]*PopCount(checks)
			e.kingAttackersCount[THEM] += 1
		}
	}

	var xRayOcc = e.occupied &^ (p.Queens | p.Bishops&friendly)

	for temp := p.Bishops & friendly; temp != 0; temp &= temp - 1 {
		e.pieceCount[US][Bishop]++
		sq = FirstOne(temp)
		eval += e.PSQT[US][Bishop][sq]

		attacks = BishopAttacks(sq, xRayOcc)

		eval += e.BishopMobility[PopCount(e.mobilityAreas[US]&attacks)]

		var kingAttacks = attacks & e.kingAreas[THEM] & e.mobilityAreas[US]
		var checks = attacks & BishopAttacks(e.kingSquare[THEM], e.occupied) & e.mobilityAreas[US]
		if kingAttacks|checks != 0 {
			e.kingAttackPower[THEM] += e.SafetyAttackPower[Bishop]*PopCount(kingAttacks) +
				e.SafetyCheckPower[Bishop]*PopCount(checks)
			e.kingAttackersCount[THEM] += 1
		}
	}

	xRayOcc = e.occupied &^ (p.Queens | p.Rooks&friendly)

	for temp := p.Rooks & friendly; temp != 0; temp &= temp - 1 {
		e.pieceCount[US][Rook]++
		sq = FirstOne(temp)
		eval += e.PSQT[US][Rook][sq]

		attacks = RookAttacks(sq, xRayOcc)

		if myPawns&forwardFileMasks[US][sq] == 0 {
			eval += e.RookFile[BoolToInt(enemyPawns&forwardFileMasks[US][sq] == 0)]
		}

		eval += e.RookMobility[PopCount(e.mobilityAreas[US]&attacks)]

		var kingAttacks = attacks & e.kingAreas[THEM] & e.mobilityAreas[US]
		var checks = attacks & RookAttacks(e.kingSquare[THEM], e.occupied) & e.mobilityAreas[US]
		if kingAttacks|checks != 0 {
			e.kingAttackPower[THEM] += e.SafetyAttackPower[Rook]*PopCount(kingAttacks) +
				e.SafetyCheckPower[Rook]*PopCount(checks)
			e.kingAttackersCount[THEM] += 1
		}
	}

	xRayOcc = e.occupied &^ (p.Queens | (p.Bishops|p.Rooks)&friendly)

	for temp := p.Queens & friendly; temp != 0; temp &= temp - 1 {
		e.pieceCount[US][Queen]++
		sq = FirstOne(temp)
		eval += e.PSQT[US][Queen][sq]

		//TODO Weiss full compability. I think better for bishopAttacks ignore bishops and for rookAttacks ignore rooks
		attacks = QueenAttacks(sq, xRayOcc)

		eval += e.QueenMobility[PopCount(e.mobilityAreas[US]&attacks)]

		var kingAttacks = attacks & e.kingAreas[THEM] & e.mobilityAreas[US]
		var checks = attacks & QueenAttacks(e.kingSquare[THEM], e.occupied) & e.mobilityAreas[US]
		if kingAttacks|checks != 0 {
			e.kingAttackPower[THEM] += e.SafetyAttackPower[Queen]*PopCount(kingAttacks) +
				e.SafetyCheckPower[Queen]*PopCount(checks)
			e.kingAttackersCount[THEM] += 1
		}
	}

	return eval
}

var countModifier = [...]int{0, 0, 64, 96, 113, 120, 124, 128}
var safeLine = [COLOUR_NB]uint64{Rank1Mask, Rank8Mask}

func evaluateKings(e *EvaluationService, p *Position, colour int) Score {
	var US = colour

	var eval Score

	var count = PopCount(QueenAttacks(e.kingSquare[US], p.Colours(US)|p.Pawns) &^ safeLine[US])
	eval += e.KingLineDanger[count]

	e.kingAttackPower[US] += (count - 3) * 8

	var safety = e.kingAttackPower[US] *
		countModifier[Min(e.kingAttackersCount[US], len(countModifier)-1)] / countModifier[len(countModifier)-1]

	//TODO Weiss full compability. safety should be clearly positive
	//eval -= S(Max(0, safety), 0)
	eval -= S(safety, 0)

	return eval
}

func evaluatePassed(e *EvaluationService, p *Position, colour int) Score {
	var US, THEM = colour, colour ^ 1

	var eval Score
	var sq int

	var occupied = p.AllPieces()

	var myPassers = e.passedPawns & p.Colours(US)

	for temp := myPassers; temp != 0; temp &= temp - 1 {
		sq = FirstOne(temp)

		var keySq int
		if colour == SideWhite {
			keySq = sq + 8
		} else {
			keySq = sq - 8
		}

		if RelativeRankOf(US, sq) > Rank3 {
			if SquareMask[keySq]&occupied != 0 {
				eval += e.PassedBlocked[RelativeRankOf(US, sq)]
			}

			if p.Rooks&p.Colours(US)&forwardFileMasks[THEM][sq] != 0 {
				eval += e.PassedRookBack
			}
		}
	}

	return eval
}

func evaluateThreats(e *EvaluationService, p *Position, colour int) Score {
	var US, THEM = colour, colour ^ 1

	var eval Score
	var count int

	var friendly = p.Colours(US)

	count = PopCount(^p.Pawns & friendly & e.pawnAttacks[THEM])
	eval += Score(count) * e.ThreatByPawn

	var pawnPushAttacks uint64
	if colour == SideWhite {
		pawnPushAttacks = AllBlackPawnAttacks(Down(p.Pawns&p.Black) &^ p.AllPieces())
	} else {
		pawnPushAttacks = AllWhitePawnAttacks(Up(p.Pawns&p.White) &^ p.AllPieces())
	}
	count = PopCount(^p.Pawns & friendly & pawnPushAttacks)
	eval += Score(count) * e.ThreatByPawnPush

	return eval
}
