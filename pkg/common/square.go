package common

import "strings"

const (
	FileA = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	Rank1 = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

const SquareNone = -1

const (
	SquareA1 = iota
	SquareB1
	SquareC1
	SquareD1
	SquareE1
	SquareF1
	SquareG1
	SquareH1
	SquareA2
	SquareB2
	SquareC2
	SquareD2
	SquareE2
	SquareF2
	SquareG2
	SquareH2
	SquareA3
	SquareB3
	SquareC3
	SquareD3
	SquareE3
	SquareF3
	SquareG3
	SquareH3
	SquareA4
	SquareB4
	SquareC4
	SquareD4
	SquareE4
	SquareF4
	SquareG4
	SquareH4
	SquareA5
	SquareB5
	SquareC5
	SquareD5
	SquareE5
	SquareF5
	SquareG5
	SquareH5
	SquareA6
	SquareB6
	SquareC6
	SquareD6
	SquareE6
	SquareF6
	SquareG6
	SquareH6
	SquareA7
	SquareB7
	SquareC7
	SquareD7
	SquareE7
	SquareF7
	SquareG7
	SquareH7
	SquareA8
	SquareB8
	SquareC8
	SquareD8
	SquareE8
	SquareF8
	SquareG8
	SquareH8
)

func FlipSquare(sq int) int {
	return sq ^ 56
}

func File(sq int) int {
	return sq & 7
}

func Rank(sq int) int {
	return sq >> 3
}

func IsDarkSquare(sq int) bool {
	return (File(sq) & 1) == (Rank(sq) & 1)
}

func AbsDelta(x, y int) int {
	if x > y {
		return x - y
	}
	return y - x
}

func FileDistance(sq1, sq2 int) int {
	return AbsDelta(File(sq1), File(sq2))
}

func RankDistance(sq1, sq2 int) int {
	return AbsDelta(Rank(sq1), Rank(sq2))
}

func SquareDistance(sq1, sq2 int) int {
	return Max(FileDistance(sq1, sq2), RankDistance(sq1, sq2))
}

func MakeSquare(file, rank int) int {
	return (rank << 3) | file
}

const (
	fileNames = "abcdefgh"
	rankNames = "12345678"
)

func SquareName(sq int) string {
	var file = fileNames[File(sq)]
	var rank = rankNames[Rank(sq)]
	return string(file) + string(rank)
}

func ParseSquare(s string) int {
	if s == "-" {
		return SquareNone
	}
	var file = strings.Index(fileNames, s[0:1])
	var rank = strings.Index(rankNames, s[1:2])
	return MakeSquare(file, rank)
}
