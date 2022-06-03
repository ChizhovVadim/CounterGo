package eval

import (
	"math/bits"

	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	pawnValue = 100
)

const (
	darkSquares = uint64(0xAA55AA55AA55AA55)
)

func BoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func Several(bb uint64) bool {
	return bb&(bb-1) != 0
}

func OnlyOne(bb uint64) bool {
	return bb != 0 && !Several(bb)
}

func RelativeRankOf(colour, sq int) int {
	if colour == SideWhite {
		return Rank(sq)
	}
	return Rank8 - Rank(sq)
}

func RelativeSquare(colour, sq int) int {
	if colour == SideWhite {
		return sq
	}
	return FlipSquare(sq)
}

func Backmost(colour int, bb uint64) int {
	if colour == SideWhite {
		return bits.TrailingZeros64(bb)
	}
	return 63 - bits.LeadingZeros64(bb)
}

func squaresOfMatchingColour(sq int) uint64 {
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

func Abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func murmurMix(k, h uint64) uint64 {
	h ^= k
	h *= uint64(0xc6a4a7935bd1e995)
	return h ^ (h >> uint(51))
}

var rankMask = [RANK_NB]uint64{
	Rank1Mask, Rank2Mask, Rank3Mask, Rank4Mask, Rank5Mask, Rank6Mask, Rank7Mask, Rank8Mask,
}

var (
	distanceBetween    [SQUARE_NB][SQUARE_NB]int
	passedPawnMasks    [COLOUR_NB][SQUARE_NB]uint64
	adjacentFilesMasks [FILE_NB]uint64
	forwardFileMasks   [COLOUR_NB][SQUARE_NB]uint64
)

func init() {
	for i := 0; i < SQUARE_NB; i++ {
		for j := 0; j < SQUARE_NB; j++ {
			distanceBetween[i][j] = SquareDistance(i, j)
		}
	}

	for f := FileA; f <= FileH; f++ {
		adjacentFilesMasks[f] = Left(FileMask[f]) | Right(FileMask[f])
	}

	for sq := 0; sq < SQUARE_NB; sq++ {
		var x = SquareMask[sq]

		passedPawnMasks[SideWhite][sq] = UpFill(Up(Left(x) | Right(x) | x))
		passedPawnMasks[SideBlack][sq] = DownFill(Down(Left(x) | Right(x) | x))

		forwardFileMasks[SideWhite][sq] = UpFill(x)
		forwardFileMasks[SideBlack][sq] = DownFill(x)
	}
}
