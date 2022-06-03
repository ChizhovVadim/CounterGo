package eval

import (
	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	pawnValue = 100
)

const (
	minorPhase = 4
	rookPhase  = 6
	queenPhase = 12
	totalPhase = 2 * (4*minorPhase + 2*rookPhase + queenPhase)
)

type EvaluationService struct {
	Weights
}

func NewEvaluationService() *EvaluationService {
	var es = &EvaluationService{}
	es.Weights.init()
	return es
}

func computePstKingShield(sq int) int {
	switch sq {
	case SquareG2, SquareB2:
		return 4
	case SquareH2, SquareH3, SquareG3, SquareF2,
		SquareA2, SquareA3, SquareB3, SquareC2:
		return 3
	default:
		return 2
	}
}

const (
	darkSquares  = uint64(0xAA55AA55AA55AA55)
	whiteOutpost = (Rank4Mask | Rank5Mask | Rank6Mask)
	blackOutpost = (Rank5Mask | Rank4Mask | Rank3Mask)
)

var (
	dist            [64][64]int
	whitePawnSquare [64]uint64
	blackPawnSquare [64]uint64
	kingZone        [64]uint64
	pstKingShield   [64]int
)

//TODO?
var passedBonus = [8]int{0, 1, 1, 3, 6, 9, 12, 0}

type evalInfo struct {
	pawns         uint64
	pawnCount     int
	knightCount   int
	bishopCount   int
	rookCount     int
	queenCount    int
	pawnAttacks   uint64
	knightAttacks uint64
	bishopAttacks uint64
	rookAttacks   uint64
	queenAttacks  uint64
	kingAttacks   uint64
	attacks       uint64
	attacksByTwo  uint64
	mobilityArea  uint64
	kingZone      uint64
	king          int
	force         int
	kingAttackNb  int
}

func (e *EvaluationService) Evaluate(p *Position) int {
	var (
		x, b         uint64
		sq           int
		keySq        int
		white, black evalInfo
		s            Score
	)

	// init

	var allPieces = p.White | p.Black

	white.pawns = p.Pawns & p.White
	black.pawns = p.Pawns & p.Black

	white.pawnCount = PopCount(white.pawns)
	black.pawnCount = PopCount(black.pawns)

	white.pawnAttacks = AllWhitePawnAttacks(white.pawns)
	white.attacksByTwo |= white.attacks & white.pawnAttacks
	white.attacks |= white.pawnAttacks

	black.pawnAttacks = AllBlackPawnAttacks(black.pawns)
	black.attacksByTwo |= black.attacks & black.pawnAttacks
	black.attacks |= black.pawnAttacks

	white.king = FirstOne(p.Kings & p.White)
	black.king = FirstOne(p.Kings & p.Black)

	white.kingAttacks = KingAttacks[white.king]
	white.attacksByTwo |= white.attacks & white.kingAttacks
	white.attacks |= white.kingAttacks

	black.kingAttacks = KingAttacks[black.king]
	black.attacksByTwo |= black.attacks & black.kingAttacks
	black.attacks |= black.kingAttacks

	white.kingZone = kingZone[white.king]
	black.kingZone = kingZone[black.king]

	white.mobilityArea = ^(white.pawns | black.pawnAttacks)
	black.mobilityArea = ^(black.pawns | white.pawnAttacks)

	// eval pieces
	var passed uint64

	for x = p.Pawns & p.White; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		s.add(e.PST[SideWhite][Pawn][sq])
		if SquareMask[sq+8]&allPieces == 0 {
			s.add(e.PawnMobility)
		}
		if adjacentFilesMask[File(sq)]&white.pawns == 0 {
			s.add(e.PawnIsolated)
		}
		if FileMask[File(sq)]&^SquareMask[sq]&white.pawns != 0 {
			s.add(e.PawnDoubled)
		}
		if pawnConnectedMask[SideWhite][sq]&white.pawns != 0 {
			s.add(e.PawnConnected[sq])
		}
		if pawnPassedMask[SideWhite][sq]&black.pawns == 0 &&
			upperRanks[Rank(sq)]&FileMask[File(sq)]&white.pawns == 0 {
			passed |= SquareMask[sq]
		}
	}

	for x = p.Pawns & p.Black; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		s.add(e.PST[SideBlack][Pawn][sq])
		if SquareMask[sq-8]&allPieces == 0 {
			s.sub(e.PawnMobility)
		}
		if adjacentFilesMask[File(sq)]&black.pawns == 0 {
			s.sub(e.PawnIsolated)
		}
		if FileMask[File(sq)]&^SquareMask[sq]&black.pawns != 0 {
			s.sub(e.PawnDoubled)
		}
		if pawnConnectedMask[SideBlack][sq]&black.pawns != 0 {
			s.sub(e.PawnConnected[FlipSquare(sq)])
		}
		if pawnPassedMask[SideBlack][sq]&white.pawns == 0 &&
			lowerRanks[Rank(sq)]&FileMask[File(sq)]&black.pawns == 0 {
			passed |= SquareMask[sq]
		}
	}

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		white.knightCount++
		sq = FirstOne(x)
		s.add(e.PST[SideWhite][Knight][sq])
		b = KnightAttacks[sq]
		s.add(e.KnightMobility[PopCount(b&white.mobilityArea)])
		white.attacksByTwo |= white.attacks & b
		white.attacks |= b
		white.knightAttacks |= b
		if (b & black.kingZone) != 0 {
			white.kingAttackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		black.knightCount++
		sq = FirstOne(x)
		s.add(e.PST[SideBlack][Knight][sq])
		b = KnightAttacks[sq]
		s.sub(e.KnightMobility[PopCount(b&black.mobilityArea)])
		black.attacksByTwo |= black.attacks & b
		black.attacks |= b
		black.knightAttacks |= b
		if (b & white.kingZone) != 0 {
			black.kingAttackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		white.bishopCount++
		sq = FirstOne(x)
		s.add(e.PST[SideWhite][Bishop][sq])
		b = BishopAttacks(sq, allPieces)
		s.add(e.BishopMobility[PopCount(b&white.mobilityArea)])
		white.attacksByTwo |= white.attacks & b
		white.attacks |= b
		white.bishopAttacks |= b
		if (b & black.kingZone) != 0 {
			white.kingAttackNb++
		}
		s.addN(e.BishopRammedPawns, PopCount(sameColorSquares(sq)&white.pawns&Down(black.pawns)))
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		black.bishopCount++
		sq = FirstOne(x)
		s.add(e.PST[SideBlack][Bishop][sq])
		b = BishopAttacks(sq, allPieces)
		s.sub(e.BishopMobility[PopCount(b&black.mobilityArea)])
		black.attacksByTwo |= black.attacks & b
		black.attacks |= b
		black.bishopAttacks |= b
		if (b & white.kingZone) != 0 {
			black.kingAttackNb++
		}
		s.addN(e.BishopRammedPawns, -PopCount(sameColorSquares(sq)&black.pawns&Up(white.pawns)))
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		white.rookCount++
		sq = FirstOne(x)
		s.add(e.PST[SideWhite][Rook][sq])
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		s.add(e.RookMobility[PopCount(b&white.mobilityArea)])
		white.attacksByTwo |= white.attacks & b
		white.attacks |= b
		white.rookAttacks |= b
		if (b & black.kingZone) != 0 {
			white.kingAttackNb++
		}
		b = FileMask[File(sq)]
		if (b & white.pawns) == 0 {
			if (b & p.Pawns) == 0 {
				s.add(e.RookOpen)
			} else {
				s.add(e.RookSemiopen)
			}
		}
	}

	for x = p.Rooks & p.Black; x != 0; x &= x - 1 {
		black.rookCount++
		sq = FirstOne(x)
		s.add(e.PST[SideBlack][Rook][sq])
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		s.sub(e.RookMobility[PopCount(b&black.mobilityArea)])
		black.attacksByTwo |= black.attacks & b
		black.attacks |= b
		black.rookAttacks |= b
		if (b & white.kingZone) != 0 {
			black.kingAttackNb++
		}
		b = FileMask[File(sq)]
		if (b & black.pawns) == 0 {
			if (b & p.Pawns) == 0 {
				s.sub(e.RookOpen)
			} else {
				s.sub(e.RookSemiopen)
			}
		}
	}

	for x = p.Queens & p.White; x != 0; x &= x - 1 {
		white.queenCount++
		sq = FirstOne(x)
		s.add(e.PST[SideWhite][Queen][sq])
		b = QueenAttacks(sq, allPieces)
		s.add(e.QueenMobility[PopCount(b&white.mobilityArea)])
		white.attacksByTwo |= white.attacks & b
		white.attacks |= b
		white.queenAttacks |= b
		if (b & black.kingZone) != 0 {
			white.kingAttackNb++
		}
		s.addN(e.KingQueenTropism, dist[sq][black.king])
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		black.queenCount++
		sq = FirstOne(x)
		s.add(e.PST[SideBlack][Queen][sq])
		b = QueenAttacks(sq, allPieces)
		s.sub(e.QueenMobility[PopCount(b&black.mobilityArea)])
		black.attacksByTwo |= black.attacks & b
		black.attacks |= b
		black.queenAttacks |= b
		if (b & white.kingZone) != 0 {
			black.kingAttackNb++
		}
		s.addN(e.KingQueenTropism, -dist[sq][white.king])
	}

	white.force = minorPhase*(white.knightCount+white.bishopCount) +
		rookPhase*white.rookCount + queenPhase*white.queenCount
	black.force = minorPhase*(black.knightCount+black.bishopCount) +
		rookPhase*black.rookCount + queenPhase*black.queenCount

	s.add(e.PST[SideWhite][King][white.king])
	s.add(e.PST[SideBlack][King][black.king])

	var kingShield = 0
	for x = white.pawns & kingShieldMask[File(white.king)] &^ lowerRanks[Rank(white.king)]; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		kingShield += pstKingShield[sq]
	}
	for x = black.pawns & kingShieldMask[File(black.king)] &^ upperRanks[Rank(black.king)]; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		kingShield -= pstKingShield[FlipSquare(sq)]
	}
	s.addN(e.KingShelter, kingShield)

	{
		weakForBlack := white.attacks & ^black.attacksByTwo & (^black.attacks | black.queenAttacks | black.kingAttacks)

		safe := ^p.White & (^black.attacks | (weakForBlack & white.attacksByTwo))

		knightThreats := KnightAttacks[black.king]
		bishopThreats := BishopAttacks(black.king, allPieces)
		rookThreats := RookAttacks(black.king, allPieces)
		queenThreats := bishopThreats | rookThreats

		knightChecks := knightThreats & safe & white.knightAttacks
		bishopChecks := bishopThreats & safe & white.bishopAttacks
		rookChecks := rookThreats & safe & white.rookAttacks
		queenChecks := queenThreats & safe & white.queenAttacks

		var kingSafety int
		if white.queenCount == 0 {
			kingSafety -= 100
		}
		kingSafety += e.KingSafetyAttackers * white.kingAttackNb
		kingSafety += e.KingSafetyWeakSquares * PopCount(black.kingZone&weakForBlack)
		kingSafety += e.KingSafetyQueenCheck * PopCount(queenChecks)
		kingSafety += e.KingSafetyRookCheck * PopCount(rookChecks)
		kingSafety += e.KingSafetyBishopCheck * PopCount(bishopChecks)
		kingSafety += e.KingSafetyKnightCheck * PopCount(knightChecks)
		kingSafety = Max(kingSafety, 0)
		s.Mg += kingSafety * kingSafety / 720
		s.Eg += kingSafety / 20
	}

	{
		weakForWhite := black.attacks & ^white.attacksByTwo & (^white.attacks | white.queenAttacks | white.kingAttacks)

		safe := ^p.Black & (^white.attacks | (weakForWhite & black.attacksByTwo))

		knightThreats := KnightAttacks[white.king]
		bishopThreats := BishopAttacks(white.king, allPieces)
		rookThreats := RookAttacks(white.king, allPieces)
		queenThreats := bishopThreats | rookThreats

		knightChecks := knightThreats & safe & black.knightAttacks
		bishopChecks := bishopThreats & safe & black.bishopAttacks
		rookChecks := rookThreats & safe & black.rookAttacks
		queenChecks := queenThreats & safe & black.queenAttacks

		var kingSafety int
		if black.queenCount == 0 {
			kingSafety -= 100
		}
		kingSafety += e.KingSafetyAttackers * black.kingAttackNb
		kingSafety += e.KingSafetyWeakSquares * PopCount(white.kingZone&weakForWhite)
		kingSafety += e.KingSafetyQueenCheck * PopCount(queenChecks)
		kingSafety += e.KingSafetyRookCheck * PopCount(rookChecks)
		kingSafety += e.KingSafetyBishopCheck * PopCount(bishopChecks)
		kingSafety += e.KingSafetyKnightCheck * PopCount(knightChecks)
		kingSafety = Max(kingSafety, 0)
		s.Mg -= kingSafety * kingSafety / 720
		s.Eg -= kingSafety / 20
	}

	// eval threats

	s.addN(e.ThreatPawn,
		PopCount(white.pawnAttacks&p.Black&^(p.Pawns|p.Queens))-
			PopCount(black.pawnAttacks&p.White&^(p.Pawns|p.Queens)))

	s.addN(e.ThreatForPawn,
		PopCount((white.rookAttacks|white.kingAttacks)&black.pawns&^black.pawnAttacks)-
			PopCount((black.rookAttacks|black.kingAttacks)&p.White&p.Pawns&^white.pawnAttacks))

	s.addN(e.ThreatPiece,
		PopCount((white.knightAttacks|white.bishopAttacks|white.rookAttacks)&p.Black&(p.Knights|p.Bishops|p.Rooks))-
			PopCount((black.knightAttacks|black.bishopAttacks|black.rookAttacks)&p.White&(p.Knights|p.Bishops|p.Rooks)))

	s.addN(e.ThreatPieceForQueen,
		PopCount((white.pawnAttacks|white.knightAttacks|white.bishopAttacks|white.rookAttacks)&p.Black&p.Queens)-
			PopCount((black.pawnAttacks|black.knightAttacks|black.bishopAttacks|black.rookAttacks)&p.White&p.Queens))

	s.addN(e.MinorProtected,
		PopCount((p.Knights|p.Bishops)&p.White&white.pawnAttacks)-
			PopCount((p.Knights|p.Bishops)&p.Black&black.pawnAttacks))

	var wstrongFields = whiteOutpost &^ DownFill(black.pawnAttacks)
	var bstrongFields = blackOutpost &^ UpFill(white.pawnAttacks)

	s.addN(e.KnightOutpost,
		PopCount(p.Knights&p.White&wstrongFields)-
			PopCount(p.Knights&p.Black&bstrongFields))

	for x = passed & p.White; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		var r = Rank(sq)
		var bonus Score
		bonus.add(e.PawnPassed)
		keySq = sq + 8
		bonus.addN(e.PawnPassedKingDist, 2*dist[keySq][black.king]-dist[keySq][white.king])
		if (SquareMask[keySq] & allPieces) == 0 {
			bonus.add(e.PawnPassedCanMove)
		}
		if (SquareMask[keySq] & black.attacks) == 0 {
			bonus.add(e.PawnPassedSafeMove)
		}

		if black.force == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				bonus.add(e.PawnPassedSquare)
			}
		}

		bonus.Mg = Max(bonus.Mg, 0)
		bonus.Eg = Max(bonus.Eg, 0)

		s.addRatio(bonus, passedBonus[r], passedBonus[Rank7])
	}

	for x = passed & p.Black; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		var r = Rank(FlipSquare(sq))
		var bonus Score
		bonus.add(e.PawnPassed)
		keySq = sq - 8
		bonus.addN(e.PawnPassedKingDist, 2*dist[keySq][white.king]-dist[keySq][black.king])
		if (SquareMask[keySq] & allPieces) == 0 {
			bonus.add(e.PawnPassedCanMove)
		}
		if (SquareMask[keySq] & white.attacks) == 0 {
			bonus.add(e.PawnPassedSafeMove)
		}

		if white.force == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				bonus.add(e.PawnPassedSquare)
			}
		}

		bonus.Mg = Max(bonus.Mg, 0)
		bonus.Eg = Max(bonus.Eg, 0)

		s.addRatio(bonus, -passedBonus[r], passedBonus[Rank7])
	}

	// eval material

	s.addN(e.PawnMaterial, white.pawnCount-black.pawnCount)
	s.addN(e.KnightMaterial, white.knightCount-black.knightCount)
	s.addN(e.BishopMaterial, white.bishopCount-black.bishopCount)
	s.addN(e.RookMaterial, white.rookCount-black.rookCount)
	s.addN(e.QueenMaterial, white.queenCount-black.queenCount)
	if white.bishopCount >= 2 {
		s.add(e.BishopPairMaterial)
	}
	if black.bishopCount >= 2 {
		s.sub(e.BishopPairMaterial)
	}

	// mix score

	var phase = white.force + black.force
	if phase > totalPhase {
		phase = totalPhase
	}

	// tempo bonus in endgame can prevent simple checkmates due to low pst values
	if phase > queenPhase+rookPhase {
		if p.WhiteMove {
			s.add(e.Tempo)
		} else {
			s.sub(e.Tempo)
		}
	}

	var result = (s.Mg*phase + s.Eg*(totalPhase-phase)) / totalPhase

	var ocb = white.force == minorPhase && black.force == minorPhase &&
		(p.Bishops&darkSquares) != 0 && (p.Bishops & ^darkSquares) != 0

	if result > 0 {
		result /= computeFactor(&white, &black, ocb)
	} else {
		result /= computeFactor(&black, &white, ocb)
	}

	if !p.WhiteMove {
		result = -result
	}

	return result
}

func computeFactor(own, their *evalInfo, ocb bool) int {
	if own.force >= queenPhase+rookPhase {
		return 1
	}
	if own.pawnCount == 0 {
		if own.force <= minorPhase {
			return 16
		}
		if own.force == 2*minorPhase && own.knightCount == 2 && their.pawnCount == 0 {
			return 16
		}
		if own.force-their.force <= minorPhase {
			return 4
		}
	} else if own.pawnCount == 1 {
		if own.force <= minorPhase && their.knightCount+their.bishopCount != 0 {
			return 8
		}
		if own.force == their.force && their.knightCount+their.bishopCount != 0 {
			return 2
		}
	} else if ocb && own.pawnCount-their.pawnCount <= 2 {
		return 2
	}
	return 1
}

func sameColorSquares(sq int) uint64 {
	if IsDarkSquare(sq) {
		return darkSquares
	}
	return ^darkSquares
}

func limit(v, min, max int) int {
	if v <= min {
		return min
	}
	if v >= max {
		return max
	}
	return v
}

var (
	upperRanks [8]uint64
	lowerRanks [8]uint64
)

var (
	ranks          = [8]uint64{Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask, Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask}
	kingShieldMask = [8]uint64{
		FileAMask | FileBMask | FileCMask,
		FileAMask | FileBMask | FileCMask,
		FileAMask | FileBMask | FileCMask,
		FileCMask | FileDMask | FileEMask,
		FileDMask | FileEMask | FileFMask,
		FileFMask | FileGMask | FileHMask,
		FileFMask | FileGMask | FileHMask,
		FileFMask | FileGMask | FileHMask}
	pawnConnectedMask [2][64]uint64
	pawnPassedMask    [2][64]uint64
	adjacentFilesMask [8]uint64
)

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
	for i := Rank7; i >= Rank1; i-- {
		upperRanks[i] = upperRanks[i+1] | ranks[i+1]
	}
	for i := Rank2; i <= Rank8; i++ {
		lowerRanks[i] = lowerRanks[i-1] | ranks[i-1]
	}
	for sq := range kingZone {
		var x = MakeSquare(limit(File(sq), FileB, FileG), limit(Rank(sq), Rank2, Rank7))
		kingZone[sq] = SquareMask[x] | KingAttacks[x]
	}
	for sq := 0; sq < 64; sq++ {
		pstKingShield[sq] = computePstKingShield(sq)
	}
	for f := FileA; f <= FileH; f++ {
		adjacentFilesMask[f] = Left(FileMask[f]) | Right(FileMask[f])
	}
	for sq := 0; sq < 64; sq++ {
		var x = SquareMask[sq]

		pawnConnectedMask[SideWhite][sq] = Left(x) | Right(x) | Down(Left(x)|Right(x))
		pawnConnectedMask[SideBlack][sq] = Left(x) | Right(x) | Up(Left(x)|Right(x))

		pawnPassedMask[SideWhite][sq] = UpFill(Up(Left(x) | Right(x) | x))
		pawnPassedMask[SideBlack][sq] = DownFill(Down(Left(x) | Right(x) | x))
	}
}
