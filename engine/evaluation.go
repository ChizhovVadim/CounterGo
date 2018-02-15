package engine

import (
	"math"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	FeatureMaterialPawn = iota
	FeatureMaterialKnight
	FeatureMaterialBishop
	FeatureMaterialRook
	FeatureMaterialQueen
	FeatureMaterialBishopPair
	FeaturePstKnight
	FeaturePstQueen
	FeaturePstKingOpening
	FeaturePstKingEndgame
	FeatureKingAttack
	FeatureBishopMobility
	FeatureRookMobility
	FeatureRook7Th
	FeatureRookOpen
	FeatureRookSemiopen
	FeatureKingPawnShiled
	FeatureMinorOnStrongField
	FeaturePawnIsolated
	FeaturePawnCenter
	FeaturePawnPassedBonus
	FeaturePawnPassedFreeBonus
	FeaturePawnPassedKingDistance
	FeaturePawnPassedSquare
	FeatureSize
)

var FeatureNames = []string{
	"MaterialPawn",
	"MaterialKnight",
	"MaterialBishop",
	"MaterialRook",
	"MaterialQueen",
	"MaterialBishopPair",
	"PstKnight",
	"PstQueen",
	"PstKingOpening",
	"PstKingEndgame",
	"KingAttack",
	"BishopMobility",
	"RookMobility",
	"Rook7Th",
	"RookOpen",
	"RookSemiopen",
	"KingPawnShiled",
	"MinorOnStrongField",
	"PawnIsolated",
	"PawnCenter",
	"PawnPassedBonus",
	"PawnPassedFreeBonus",
	"PawnPassedKingDistance",
	"PawnPassedSquare",
}

type EvalInfo struct {
	Features                               [FeatureSize]int
	Phase                                  int
	wp, bp, wn, bn, wb, bb, wr, br, wq, bq int
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

	pawnPassedBonus  = [8]int{0, 0, 0, 2, 6, 12, 21, 0}
	kingAttackWeight = [...]int{0, 0, 2, 3, 4}
	featureWeights   = [FeatureSize * 2]int{7562, 8403, 32752, 33879, 32721, 34956, 46452, 49740, 109394, 96886, 1445, 3578, 1061, 1347, 0, 687, -891, 0, 0, -488, 215, 0, 60, 8, 10, 105, 2165, 2933, 1472, 0, 641, 0, -505, 0, 1829, 0, -1702, -658, 3063, 0, 217, 349, 0, 107, 0, 56, 0, 0}
)

const (
	kingAttackUnitKnight = 2
	kingAttackUnitBishop = 2
	kingAttackUnitRook   = 3
	kingAttackUnitQueen  = 4
)

var (
	dist            [64][64]int
	bishopMobility  [13 + 1]int
	rookMobility    [14 + 1]int
	whitePawnSquare [64]uint64
	blackPawnSquare [64]uint64
)

func Evaluate(p *Position) int {
	var ei EvalInfo
	EvaluateFeatures(p, &ei)

	var opening, endgame int
	for i, x := range ei.Features {
		opening += x * featureWeights[i*2]
		endgame += x * featureWeights[i*2+1]
	}

	var score = (opening*ei.Phase + endgame*(64-ei.Phase)) / 64

	if ei.wp == 0 && score > 0 {
		if ei.wn+ei.wb <= 1 && ei.wr+ei.wq == 0 {
			score /= 16
		} else if ei.wn == 2 && ei.wb+ei.wr+ei.wq == 0 && ei.bp == 0 {
			score /= 16
		} else if (ei.wn+ei.wb+2*ei.wr+4*ei.wq)-(ei.bn+ei.bb+2*ei.br+4*ei.bq) <= 1 {
			score /= 4
		}
	}

	if ei.bp == 0 && score < 0 {
		if ei.bn+ei.bb <= 1 && ei.br+ei.bq == 0 {
			score /= 16
		} else if ei.bn == 2 && ei.bb+ei.br+ei.bq == 0 && ei.wp == 0 {
			score /= 16
		} else if (ei.bn+ei.bb+2*ei.br+4*ei.bq)-(ei.wn+ei.wb+2*ei.wr+4*ei.wq) <= 1 {
			score /= 4
		}
	}

	if (p.Knights|p.Rooks|p.Queens) == 0 &&
		ei.wb == 1 && ei.bb == 1 && AbsDelta(ei.wp, ei.bp) <= 2 &&
		(p.Bishops&darkSquares) != 0 &&
		(p.Bishops & ^darkSquares) != 0 {
		score /= 2
	}

	if !p.WhiteMove {
		score = -score
	}

	return score / 100
}

func EvaluateFeatures(p *Position, ei *EvalInfo) {
	var (
		x, b             uint64
		sq, keySq, bonus int

		allPieces = p.White | p.Black
		wkingSq   = FirstOne(p.Kings & p.White)
		bkingSq   = FirstOne(p.Kings & p.Black)
	)

	ei.wp = PopCount(p.Pawns & p.White)
	ei.bp = PopCount(p.Pawns & p.Black)

	ei.Features[FeaturePawnIsolated] = PopCount(getIsolatedPawns(p.Pawns&p.White)) -
		PopCount(getIsolatedPawns(p.Pawns&p.Black))

	b = p.Pawns & p.White & (Rank4Mask | Rank5Mask | Rank6Mask)
	if (b & FileDMask) != 0 {
		ei.Features[FeaturePawnCenter]++
	}
	if (b & FileEMask) != 0 {
		ei.Features[FeaturePawnCenter]++
	}
	b = p.Pawns & p.Black & (Rank5Mask | Rank4Mask | Rank3Mask)
	if (b & FileDMask) != 0 {
		ei.Features[FeaturePawnCenter]--
	}
	if (b & FileEMask) != 0 {
		ei.Features[FeaturePawnCenter]--
	}

	var wkingMoves = KingAttacks[wkingSq]
	var bkingMoves = KingAttacks[bkingSq]
	var wattackTot, wattackNb, battackTot, battackNb int

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		ei.wn++
		sq = FirstOne(x)
		ei.Features[FeaturePstKnight] += center[sq]
		if (KnightAttacks[sq] & bkingMoves) != 0 {
			wattackTot += kingAttackUnitKnight
			wattackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		ei.bn++
		sq = FirstOne(x)
		ei.Features[FeaturePstKnight] -= center[sq]
		if (KnightAttacks[sq] & wkingMoves) != 0 {
			battackTot += kingAttackUnitKnight
			battackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		ei.wb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		ei.Features[FeatureBishopMobility] += bishopMobility[PopCount(b)]
		if (b & bkingMoves) != 0 {
			wattackTot += kingAttackUnitBishop
			wattackNb++
		}
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		ei.bb++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		ei.Features[FeatureBishopMobility] -= bishopMobility[PopCount(b)]
		if (b & wkingMoves) != 0 {
			battackTot += kingAttackUnitBishop
			battackNb++
		}
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		ei.wr++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 {
			ei.Features[FeatureRook7Th]++
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		ei.Features[FeatureRookMobility] += rookMobility[PopCount(b)]
		if (b & bkingMoves) != 0 {
			wattackTot += kingAttackUnitRook
			wattackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.White) == 0 {
			if (b & p.Pawns) == 0 {
				ei.Features[FeatureRookOpen]++
			} else {
				ei.Features[FeatureRookSemiopen]++
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		ei.br++
		sq = FirstOne(x)
		if Rank(sq) == Rank2 {
			ei.Features[FeatureRook7Th]--
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		ei.Features[FeatureRookMobility] -= rookMobility[PopCount(b)]
		if (b & wkingMoves) != 0 {
			battackTot += kingAttackUnitRook
			battackNb++
		}
		b = FileMask[File(sq)]
		if (b & p.Pawns & p.Black) == 0 {
			if (b & p.Pawns) == 0 {
				ei.Features[FeatureRookOpen]--
			} else {
				ei.Features[FeatureRookSemiopen]--
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		ei.wq++
		sq = FirstOne(x)
		ei.Features[FeaturePstQueen] += center[sq]
		if (QueenAttacks(sq, allPieces) & bkingMoves) != 0 {
			wattackTot += kingAttackUnitQueen
			wattackNb++
		}
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		ei.bq++
		sq = FirstOne(x)
		ei.Features[FeaturePstQueen] -= center[sq]
		if (QueenAttacks(sq, allPieces) & wkingMoves) != 0 {
			battackTot += kingAttackUnitQueen
			battackNb++
		}
	}

	ei.Features[FeatureKingAttack] = wattackTot*kingAttackWeight[Min(wattackNb, len(kingAttackWeight)-1)] -
		battackTot*kingAttackWeight[Min(battackNb, len(kingAttackWeight)-1)]

	var matIndexWhite = Min(32, (ei.wn+ei.wb)*3+ei.wr*5+ei.wq*10)
	var matIndexBlack = Min(32, (ei.bn+ei.bb)*3+ei.br*5+ei.bq*10)

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(sq)]
		ei.Features[FeaturePawnPassedBonus] += bonus
		keySq = sq + 8
		ei.Features[FeaturePawnPassedKingDistance] += bonus * (dist[keySq][bkingSq]*2 - dist[keySq][wkingSq])
		if (SquareMask[keySq] & allPieces) == 0 {
			ei.Features[FeaturePawnPassedFreeBonus] += bonus
		}

		if matIndexBlack == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				ei.Features[FeaturePawnPassedSquare] += Rank(f1) / 6
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		bonus = pawnPassedBonus[Rank(FlipSquare(sq))]
		ei.Features[FeaturePawnPassedBonus] -= bonus
		keySq = sq - 8
		ei.Features[FeaturePawnPassedKingDistance] -= bonus * (dist[keySq][wkingSq]*2 - dist[keySq][bkingSq])
		if (SquareMask[keySq] & allPieces) == 0 {
			ei.Features[FeaturePawnPassedFreeBonus] -= bonus
		}

		if matIndexWhite == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				ei.Features[FeaturePawnPassedSquare] -= Rank(FlipSquare(f1)) / 6
			}
		}
	}

	ei.Features[FeatureKingPawnShiled] += shelterWKingSquare(p, wkingSq)
	ei.Features[FeatureKingPawnShiled] -= shelterBKingSquare(p, bkingSq)

	ei.Features[FeaturePstKingOpening] += center_k[wkingSq]
	ei.Features[FeaturePstKingOpening] -= center_k[FlipSquare(bkingSq)]

	ei.Features[FeaturePstKingEndgame] += center[wkingSq]
	ei.Features[FeaturePstKingEndgame] -= center[bkingSq]

	var wStrongFields = AllWhitePawnAttacks(p.Pawns&p.White) &^
		DownFill(AllBlackPawnAttacks(p.Pawns&p.Black)) & 0xffffffff00000000

	var bStrongFields = AllBlackPawnAttacks(p.Pawns&p.Black) &^
		UpFill(AllWhitePawnAttacks(p.Pawns&p.White)) & 0x00000000ffffffff

	var wMinorOnStrongFields = wStrongFields & (p.Knights | p.Bishops) & p.White
	var bMinorOnStrongFields = bStrongFields & (p.Knights | p.Bishops) & p.Black

	ei.Features[FeatureMinorOnStrongField] = PopCount(wMinorOnStrongFields) -
		PopCount(bMinorOnStrongFields)

	ei.Features[FeatureMaterialPawn] = ei.wp - ei.bp
	ei.Features[FeatureMaterialKnight] = ei.wn - ei.bn
	ei.Features[FeatureMaterialBishop] = ei.wb - ei.bb
	if ei.wb >= 2 {
		ei.Features[FeatureMaterialBishopPair]++
	}
	if ei.bb >= 2 {
		ei.Features[FeatureMaterialBishopPair]--
	}
	ei.Features[FeatureMaterialRook] = ei.wr - ei.br
	ei.Features[FeatureMaterialQueen] = ei.wq - ei.bq

	ei.Phase = matIndexWhite + matIndexBlack
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
	return penalty
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
	return penalty
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
	for i := range bishopMobility {
		bishopMobility[i] = int(math.Sqrt(float64(i)/13)*200 - 100) // [-100;+100]
	}
	for i := range rookMobility {
		rookMobility[i] = int(math.Sqrt(float64(i)/14)*200 - 100) // [-100;+100]
	}
	featureWeights[FeaturePawnPassedSquare*2+1] = 20000
}
