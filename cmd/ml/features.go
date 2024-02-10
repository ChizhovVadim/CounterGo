package main

import (
	"sort"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

// На самом деле признаков 736, тк пешки не могут быть на крайних горизонталях. Переделать?
func toFeatures768(pos *common.Position) []domain.FeatureInfo {
	var result = make([]domain.FeatureInfo, 0, common.PopCount(pos.AllPieces()))
	for x := pos.AllPieces(); x != 0; x &= x - 1 {
		var sq = common.FirstOne(x)
		var pt, side = pos.GetPieceTypeAndSide(sq)
		var piece12 = pt - common.Pawn
		if !side {
			piece12 += 6
		}
		var index = int16(sq ^ piece12<<6)
		result = append(result, domain.FeatureInfo{
			Index: index,
			Value: 1,
		})
	}
	return result
}

func sortFeatures(features []domain.FeatureInfo) {
	sort.Slice(features, func(i, j int) bool {
		return features[i].Index < features[j].Index
	})
}

// предусловие: признаки отсортированы по порядковому номеру признака
func compareFeatures(l, r []domain.FeatureInfo) int {
	var n = len(l)
	if m := len(r); m < n {
		n = m
	}
	for i := 0; i < n; i++ {
		if l[i].Index < r[i].Index {
			return -1
		}
		if l[i].Index > r[i].Index {
			return 1
		}
		if l[i].Value < r[i].Value {
			return -1
		}
		if l[i].Value > r[i].Value {
			return 1
		}
	}
	if len(l) > len(r) {
		return -1
	}
	if len(l) < len(r) {
		return 1
	}
	return 0
}
