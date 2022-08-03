package common

import (
	"strings"
	"unicode"
)

func Min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func Max(l, r int) int {
	if l > r {
		return l
	}
	return r
}

func let(ok bool, yes, no int) int {
	if ok {
		return yes
	}
	return no
}

func parsePiece(ch rune) coloredPiece {
	var side = unicode.IsUpper(ch)
	var spiece = string(unicode.ToLower(ch))
	var i = strings.Index("pnbrqk", spiece)
	if i < 0 {
		return coloredPiece{Empty, false}
	}
	return coloredPiece{i + Pawn, side}
}
