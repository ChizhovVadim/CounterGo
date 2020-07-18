package eval

import (
	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	pawnValue = 100
	maxPhase  = 24
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
	whiteOutpost = (FileCMask | FileDMask | FileEMask | FileFMask) & (Rank4Mask | Rank5Mask | Rank6Mask)
	blackOutpost = (FileCMask | FileDMask | FileEMask | FileFMask) & (Rank5Mask | Rank4Mask | Rank3Mask)
)

var (
	dist            [64][64]int
	whitePawnSquare [64]uint64
	blackPawnSquare [64]uint64
	kingZone        [64]uint64
	pstKingShield   [64]int
)

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
	black.pawnAttacks = AllBlackPawnAttacks(black.pawns)

	white.king = FirstOne(p.Kings & p.White)
	black.king = FirstOne(p.Kings & p.Black)

	white.kingAttacks = KingAttacks[white.king]
	black.kingAttacks = KingAttacks[black.king]

	white.kingZone = kingZone[white.king]
	black.kingZone = kingZone[black.king]

	white.mobilityArea = ^(white.pawns | black.pawnAttacks)
	black.mobilityArea = ^(black.pawns | white.pawnAttacks)

	// eval pieces

	for x = p.Knights & p.White; x != 0; x &= x - 1 {
		white.knightCount++
		sq = FirstOne(x)
		s.add(e.PST[sideWhite][Knight][sq])
		b = KnightAttacks[sq]
		s.add(e.KnightMobility[PopCount(b&white.mobilityArea)])
		white.knightAttacks |= b
		if (b & black.kingZone & white.mobilityArea) != 0 {
			white.kingAttackNb++
		}
	}

	for x = p.Knights & p.Black; x != 0; x &= x - 1 {
		black.knightCount++
		sq = FirstOne(x)
		s.add(e.PST[sideBlack][Knight][sq])
		b = KnightAttacks[sq]
		s.sub(e.KnightMobility[PopCount(b&black.mobilityArea)])
		black.knightAttacks |= b
		if (b & white.kingZone & black.mobilityArea) != 0 {
			black.kingAttackNb++
		}
	}

	for x = p.Bishops & p.White; x != 0; x &= x - 1 {
		white.bishopCount++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		s.add(e.BishopMobility[PopCount(b&white.mobilityArea)])
		white.bishopAttacks |= b
		if (b & black.kingZone & white.mobilityArea) != 0 {
			white.kingAttackNb++
		}
		s.addN(e.BishopRammedPawns, PopCount(sameColorSquares(sq)&white.pawns&Down(black.pawns)))
	}

	for x = p.Bishops & p.Black; x != 0; x &= x - 1 {
		black.bishopCount++
		sq = FirstOne(x)
		b = BishopAttacks(sq, allPieces)
		s.sub(e.BishopMobility[PopCount(b&black.mobilityArea)])
		black.bishopAttacks |= b
		if (b & white.kingZone & black.mobilityArea) != 0 {
			black.kingAttackNb++
		}
		s.addN(e.BishopRammedPawns, -PopCount(sameColorSquares(sq)&black.pawns&Up(white.pawns)))
	}

	for x = p.Rooks & p.White; x != 0; x &= x - 1 {
		white.rookCount++
		sq = FirstOne(x)
		if Rank(sq) == Rank7 &&
			((p.Pawns&p.Black&Rank7Mask) != 0 || Rank(black.king) == Rank8) {
			s.add(e.Rook7th)
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.White))
		s.add(e.RookMobility[PopCount(b&white.mobilityArea)])
		white.rookAttacks |= b
		if (b & black.kingZone & white.mobilityArea) != 0 {
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
		if Rank(sq) == Rank2 &&
			((p.Pawns&p.White&Rank2Mask) != 0 || Rank(white.king) == Rank1) {
			s.sub(e.Rook7th)
		}
		b = RookAttacks(sq, allPieces^(p.Rooks&p.Black))
		s.sub(e.RookMobility[PopCount(b&black.mobilityArea)])
		black.rookAttacks |= b
		if (b & white.kingZone & black.mobilityArea) != 0 {
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
		s.add(e.PST[sideWhite][Queen][sq])
		b = QueenAttacks(sq, allPieces)
		s.add(e.QueenMobility[PopCount(b&white.mobilityArea)])
		white.queenAttacks |= b
		if (b & black.kingZone & white.mobilityArea) != 0 {
			white.kingAttackNb++
		}
		s.addN(e.KingQueenTropism, dist[sq][black.king])
	}

	for x = p.Queens & p.Black; x != 0; x &= x - 1 {
		black.queenCount++
		sq = FirstOne(x)
		s.add(e.PST[sideBlack][Queen][sq])
		b = QueenAttacks(sq, allPieces)
		s.sub(e.QueenMobility[PopCount(b&black.mobilityArea)])
		black.queenAttacks |= b
		if (b & white.kingZone & black.mobilityArea) != 0 {
			black.kingAttackNb++
		}
		s.addN(e.KingQueenTropism, -dist[sq][white.king])
	}

	white.force = white.knightCount + white.bishopCount +
		2*white.rookCount + 4*white.queenCount
	black.force = black.knightCount + black.bishopCount +
		2*black.rookCount + 4*black.queenCount

	s.add(e.PST[sideWhite][King][white.king])
	s.add(e.PST[sideBlack][King][black.king])

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

	s.add(e.KingAttack[Min(len(e.KingAttack)-1, white.kingAttackNb)])
	s.sub(e.KingAttack[Min(len(e.KingAttack)-1, black.kingAttackNb)])

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

	// eval pawns

	s.addN(e.PawnWeak,
		PopCount(getWhiteWeakPawns(p))-
			PopCount(getBlackWeakPawns(p)))

	s.addN(e.PawnDoubled,
		PopCount(getIsolatedPawns(p.Pawns&p.White)&getDoubledPawns(p.Pawns&p.White))-
			PopCount(getIsolatedPawns(p.Pawns&p.Black)&getDoubledPawns(p.Pawns&p.Black)))

	s.addN(e.PawnDuo,
		PopCount(p.Pawns&p.White&(Left(p.Pawns&p.White)|Right(p.Pawns&p.White)))-
			PopCount(p.Pawns&p.Black&(Left(p.Pawns&p.Black)|Right(p.Pawns&p.Black))))

	s.addN(e.PawnProtected,
		PopCount(white.pawns&white.pawnAttacks)-
			PopCount(black.pawns&black.pawnAttacks))

	s.addN(e.MinorProtected,
		PopCount((p.Knights|p.Bishops)&p.White&white.pawnAttacks)-
			PopCount((p.Knights|p.Bishops)&p.Black&black.pawnAttacks))

	var wstrongFields = whiteOutpost &^ DownFill(black.pawnAttacks)
	var bstrongFields = blackOutpost &^ UpFill(white.pawnAttacks)

	s.addN(e.KnightOutpost,
		PopCount(p.Knights&p.White&wstrongFields)-
			PopCount(p.Knights&p.Black&bstrongFields))

	s.addN(e.PawnBlockedByOwnPiece,
		PopCount(p.Pawns&p.White&^white.kingZone&(Rank2Mask|Rank3Mask)&Down(p.White))-
			PopCount(p.Pawns&p.Black&^black.kingZone&(Rank7Mask|Rank6Mask)&Up(p.Black)))

	s.addN(e.PawnRammed,
		PopCount(p.Pawns&p.White&(Rank2Mask|Rank3Mask)&Down(p.Pawns&p.Black))-
			PopCount(p.Pawns&p.Black&(Rank7Mask|Rank6Mask)&Up(p.Pawns&p.White)))

	for x = getWhitePassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		var r = Rank(sq)
		s.add(e.PawnPassed[r])
		keySq = sq + 8
		s.addN(e.PawnPassedOppKing[r], dist[keySq][black.king])
		s.addN(e.PawnPassedOwnKing[r], dist[keySq][white.king])
		if (SquareMask[keySq] & p.Black) == 0 {
			s.add(e.PawnPassedFree[r])
		}

		if black.force == 0 {
			var f1 = sq
			if !p.WhiteMove {
				f1 -= 8
			}
			if (whitePawnSquare[f1] & p.Kings & p.Black) == 0 {
				s.addN(e.PawnPassedSquare, Rank(f1)-Rank1)
			}
		}
	}

	for x = getBlackPassedPawns(p); x != 0; x &= x - 1 {
		sq = FirstOne(x)
		var r = Rank(FlipSquare(sq))
		s.sub(e.PawnPassed[r])
		keySq = sq - 8
		s.addN(e.PawnPassedOppKing[r], -dist[keySq][white.king])
		s.addN(e.PawnPassedOwnKing[r], -dist[keySq][black.king])
		if (SquareMask[keySq] & p.White) == 0 {
			s.sub(e.PawnPassedFree[r])
		}

		if white.force == 0 {
			var f1 = sq
			if p.WhiteMove {
				f1 += 8
			}
			if (blackPawnSquare[f1] & p.Kings & p.White) == 0 {
				s.addN(e.PawnPassedSquare, Rank(f1)-Rank8)
			}
		}
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
	if phase > maxPhase {
		phase = maxPhase
	}

	// tempo bonus in endgame can prevent simple checkmates due to low pst values
	if phase > 6 {
		if p.WhiteMove {
			s.add(e.Tempo)
		} else {
			s.sub(e.Tempo)
		}
	}

	var result = (s.Mg*phase + s.Eg*(maxPhase-phase)) / maxPhase

	var ocb = white.force == 1 && black.force == 1 &&
		(p.Bishops&darkSquares) != 0 && (p.Bishops & ^darkSquares) != 0
	var whiteFactor = computeFactor(&white, &black, ocb)
	var blackFactor = computeFactor(&black, &white, ocb)

	//TODO (bug) Q VS R+3minors is not draw
	if result > 0 {
		result /= whiteFactor
	} else {
		result /= blackFactor
	}

	if !p.WhiteMove {
		result = -result
	}

	return result
}

func computeFactor(own, their *evalInfo, ocb bool) int {
	if own.force >= 6 {
		return 1
	}
	if own.pawnCount == 0 {
		if own.force <= 1 {
			return 16
		}
		if own.force == 2 && own.knightCount == 2 && their.pawnCount == 0 {
			return 16
		}
		if own.force-their.force <= 1 {
			return 4
		}
	} else if own.pawnCount == 1 {
		if own.force <= 1 && their.knightCount+their.bishopCount != 0 {
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

func getWhiteWeakPawns(p *Position) uint64 {
	var pawns = p.Pawns & p.White
	var supported = UpFill(Left(pawns) | Right(pawns))
	var weak = uint64(0)
	weak |= getIsolatedPawns(pawns)
	weak |= (Rank2Mask | Rank3Mask | Rank4Mask) & Down(AllBlackPawnAttacks(p.Pawns&p.Black)) &^ supported
	return pawns & weak

}

func getBlackWeakPawns(p *Position) uint64 {
	var pawns = p.Pawns & p.Black
	var supported = DownFill(Left(pawns) | Right(pawns))
	var weak = uint64(0)
	weak |= getIsolatedPawns(pawns)
	weak |= (Rank7Mask | Rank6Mask | Rank5Mask) & Up(AllWhitePawnAttacks(p.Pawns&p.White)) &^ supported
	return pawns & weak
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
		//kingZone[sq] = SquareMask[sq] | KingAttacks[sq]
		var x = MakeSquare(limit(File(sq), FileB, FileG), limit(Rank(sq), Rank2, Rank7))
		kingZone[sq] = SquareMask[x] | KingAttacks[x]
	}
	for sq := 0; sq < 64; sq++ {
		pstKingShield[sq] = computePstKingShield(sq)
	}
}
