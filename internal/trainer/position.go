package trainer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func FromFen(fen string) ([]int16, error) {
	length := 0
	spacesCount := 0
	for _, ch := range fen {
		if ch == ' ' {
			if spacesCount >= 1 {
				break
			}
			spacesCount += 1
		}

		if ch == 'w' {
			length++
		} else if spacesCount == 0 && ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') {
			length++
		}
	}

	var input = make([]int16, 0, length)

	var tokens = strings.Split(fen, " ")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("parse fen failed %v", fen)
	}

	var i = 0
	for _, ch := range tokens[0] {
		if unicode.IsDigit(ch) {
			var n, _ = strconv.Atoi(string(ch))
			i += n
		} else if unicode.IsLetter(ch) {
			var piece = pieceFromRune(ch)
			if piece >= 0 {
				var square = i ^ 56
				input = append(input, calculateNetInputIndex(piece, square))
				i++
			}
		}
	}

	if tokens[1] == "w" {
		input = append(input, 768)
	}

	if len(input) != length {
		return nil, fmt.Errorf("parse fen len fail %v %v %v", fen, len(input), length)
	}

	return input, nil
}

func calculateNetInputIndex(piece12, square int) int16 {
	return int16(square ^ piece12<<6)
}

func pieceFromRune(r rune) int {
	switch r {
	case 'P':
		return 0
	case 'N':
		return 1
	case 'B':
		return 2
	case 'R':
		return 3
	case 'Q':
		return 4
	case 'K':
		return 5
	case 'p':
		return 6
	case 'n':
		return 7
	case 'b':
		return 8
	case 'r':
		return 9
	case 'q':
		return 10
	case 'k':
		return 11
	}
	return -1
}

func MirrorInput(input []int16, buffer []int16) []int16 {
	var result = buffer[:0]
	var whiteMove = false
	for _, index := range input {
		if index == 768 {
			whiteMove = true
		} else {
			var sq = int(index % 64)
			var pt = int(index >> 6)
			if pt >= 6 {
				pt -= 6
			} else {
				pt += 6
			}
			var mirrorInput = calculateNetInputIndex(pt, sq^56)
			result = append(result, mirrorInput)
		}
	}
	if !whiteMove {
		result = append(result, 768)
	}
	return result
}
