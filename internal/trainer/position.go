package trainer

import "github.com/ChizhovVadim/CounterGo/pkg/common"

func toFeatures(pos *common.Position) []int16 {
	var buffer [64]int16
	var size int
	for x := pos.AllPieces(); x != 0; x &= x - 1 {
		var sq = common.FirstOne(x)
		var pt, side = pos.GetPieceTypeAndSide(sq)
		var piece12 = pt - common.Pawn
		if !side {
			piece12 += 6
		}
		var index = calculateNetInputIndex(piece12, sq)
		buffer[size] = int16(index)
		size++
	}

	var result = make([]int16, size)
	copy(result, buffer[:size])
	return result
}

func calculateNetInputIndex(piece12, square int) int16 {
	return int16(square ^ piece12<<6)
}

func mirrorInput(input []int16) []int16 {
	var buffer [64]int16
	var size int

	for _, index := range input {
		var sq = int(index % 64)
		var pt = int(index >> 6)
		if pt >= 6 {
			pt -= 6
		} else {
			pt += 6
		}
		var mirrorInput = calculateNetInputIndex(pt, sq^56)
		buffer[size] = mirrorInput
		size++
	}
	var result = make([]int16, size)
	copy(result, buffer[:size])
	return result
}
