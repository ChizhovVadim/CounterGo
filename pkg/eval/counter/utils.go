package eval

import (
	"math/bits"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	darkSquares = uint64(0xAA55AA55AA55AA55)
)

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func onlyOne(bb uint64) bool {
	return bb != 0 && !MoreThanOne(bb)
}

func sameColorSquares(sq int) uint64 {
	if IsDarkSquare(sq) {
		return darkSquares
	}
	return ^darkSquares
}

func relativeSq32(side, sq int) int {
	if side == SideBlack {
		sq = FlipSquare(sq)
	}
	var f = File(sq)
	if f >= FileE {
		f = FileH - f
	}
	return f + 4*Rank(sq)
}

func relativeRankOf(colour, sq int) int {
	if colour == SideWhite {
		return Rank(sq)
	}
	return Rank8 - Rank(sq)
}

func file4(sq int) int {
	var f = File(sq)
	if f >= FileE {
		f = FileH - f
	}
	return f
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

func backmost(colour int, bb uint64) int {
	if colour == SideWhite {
		return bits.TrailingZeros64(bb)
	}
	return 63 - bits.LeadingZeros64(bb)
}

func murmurMix(k, h uint64) uint64 {
	h ^= k
	h *= uint64(0xc6a4a7935bd1e995)
	return h ^ (h >> uint(51))
}

var outpostSquares = [COLOUR_NB]uint64{
	(Rank4Mask | Rank5Mask | Rank6Mask),
	(Rank5Mask | Rank4Mask | Rank3Mask),
}

var rankMasks = [RANK_NB]uint64{Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask, Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask}

var pawnConnectedMask [COLOUR_NB][SQUARE_NB]uint64
var pawnPassedMask [COLOUR_NB][SQUARE_NB]uint64
var outpostSquareMasks [COLOUR_NB][SQUARE_NB]uint64
var kingShieldMasks [COLOUR_NB][SQUARE_NB]uint64
var forwardFileMasks [COLOUR_NB][SQUARE_NB]uint64
var kingAreaMasks [COLOUR_NB][SQUARE_NB]uint64
var adjacentFilesMask [FILE_NB]uint64
var forwardRanksMasks [COLOUR_NB][RANK_NB]uint64
var distanceBetween [SQUARE_NB][SQUARE_NB]int

func init() {
	for i := 0; i < SQUARE_NB; i++ {
		for j := 0; j < SQUARE_NB; j++ {
			distanceBetween[i][j] = SquareDistance(i, j)
		}
	}

	for f := FileA; f <= FileH; f++ {
		adjacentFilesMask[f] = Left(FileMask[f]) | Right(FileMask[f])
	}
	for r := Rank1; r <= Rank8; r++ {
		forwardRanksMasks[SideWhite][r] = UpFill(rankMasks[r])
		forwardRanksMasks[SideBlack][r] = DownFill(rankMasks[r])
	}

	for sq := 0; sq < 64; sq++ {
		var x = SquareMask[sq]

		pawnConnectedMask[SideWhite][sq] = Left(x) | Right(x) | Down(Left(x)|Right(x))
		pawnConnectedMask[SideBlack][sq] = Left(x) | Right(x) | Up(Left(x)|Right(x))

		pawnPassedMask[SideWhite][sq] = UpFill(Up(Left(x) | Right(x) | x))
		pawnPassedMask[SideBlack][sq] = DownFill(Down(Left(x) | Right(x) | x))

		outpostSquareMasks[SideWhite][sq] = pawnPassedMask[SideWhite][sq] & ^FileMask[File(sq)]
		outpostSquareMasks[SideBlack][sq] = pawnPassedMask[SideBlack][sq] & ^FileMask[File(sq)]

		kingShieldMasks[SideWhite][sq] = UpFill(Left(x) | Right(x) | x)
		kingShieldMasks[SideBlack][sq] = DownFill(Left(x) | Right(x) | x)

		forwardFileMasks[SideWhite][sq] = UpFill(x)
		forwardFileMasks[SideBlack][sq] = DownFill(x)

		var kingZoneSq = MakeSquare(limit(File(sq), FileB, FileG), limit(Rank(sq), Rank2, Rank7))
		//var kingZoneSq = sq
		kingAreaMasks[SideWhite][sq] = KingAttacks[kingZoneSq] | SquareMask[kingZoneSq]
		kingAreaMasks[SideBlack][sq] = kingAreaMasks[SideWhite][sq]
	}
}
