package eval

import (
	"math"
	"math/bits"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	darkSquares = uint64(0xAA55AA55AA55AA55)
)

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

func relativeUp(colour int, b uint64) uint64 {
	if colour == SideWhite {
		return Up(b)
	}
	return Down(b)
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

func onlyOne(bb uint64) bool {
	return bb != 0 && !MoreThanOne(bb)
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
var ranks = [8]uint64{Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask, Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask}

var forwardFileMasks [COLOUR_NB][SQUARE_NB]uint64
var pawnConnectedMask [COLOUR_NB][SQUARE_NB]uint64
var passedPawnMasks [COLOUR_NB][SQUARE_NB]uint64
var kingShieldMasks [COLOUR_NB][SQUARE_NB]uint64
var distanceBetween [SQUARE_NB][SQUARE_NB]int
var kingAreaMasks [COLOUR_NB][SQUARE_NB]uint64
var adjacentFilesMask [FILE_NB]uint64
var upperRankMasks [COLOUR_NB][RANK_NB]uint64 //TODO
var forwardRanksMasks [COLOUR_NB][RANK_NB]uint64
var outpostSquareMasks [COLOUR_NB][SQUARE_NB]uint64
var sqrtInt [32]int

func init() {
	for i := 0; i < SQUARE_NB; i++ {
		for j := 0; j < SQUARE_NB; j++ {
			distanceBetween[i][j] = SquareDistance(i, j)
		}
	}

	for i := range sqrtInt {
		sqrtInt[i] = int(math.Round(float64(i)))
	}

	for f := FileA; f <= FileH; f++ {
		adjacentFilesMask[f] = Left(FileMask[f]) | Right(FileMask[f])
	}
	for r := Rank1; r <= Rank8; r++ {
		upperRankMasks[SideWhite][r] = UpFill(Up(ranks[r]))
		upperRankMasks[SideBlack][r] = DownFill(Down(ranks[r]))

		forwardRanksMasks[SideWhite][r] = UpFill(ranks[r])
		forwardRanksMasks[SideBlack][r] = DownFill(ranks[r])
	}

	for sq := 0; sq < SQUARE_NB; sq++ {
		var x = SquareMask[sq]

		forwardFileMasks[SideWhite][sq] = UpFill(x)
		forwardFileMasks[SideBlack][sq] = DownFill(x)

		pawnConnectedMask[SideWhite][sq] = Left(x) | Right(x) | Down(Left(x)|Right(x))
		pawnConnectedMask[SideBlack][sq] = Left(x) | Right(x) | Up(Left(x)|Right(x))

		passedPawnMasks[SideWhite][sq] = UpFill(Up(Left(x) | Right(x) | x))
		passedPawnMasks[SideBlack][sq] = DownFill(Down(Left(x) | Right(x) | x))

		outpostSquareMasks[SideWhite][sq] = passedPawnMasks[SideWhite][sq] & ^FileMask[File(sq)]
		outpostSquareMasks[SideBlack][sq] = passedPawnMasks[SideBlack][sq] & ^FileMask[File(sq)]

		kingShieldMasks[SideWhite][sq] = UpFill(Left(x) | Right(x) | x)
		kingShieldMasks[SideBlack][sq] = DownFill(Left(x) | Right(x) | x)

		var kingZoneSq = MakeSquare(limit(File(sq), FileB, FileG), limit(Rank(sq), Rank2, Rank7))
		//var kingZoneSq = sq
		kingAreaMasks[SideWhite][sq] = KingAttacks[kingZoneSq] | SquareMask[kingZoneSq]
		kingAreaMasks[SideBlack][sq] = kingAreaMasks[SideWhite][sq]
	}
}
