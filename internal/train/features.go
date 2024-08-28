package train

import (
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Feature768Provider struct{}

// На самом деле признаков 736, тк пешки не могут быть на крайних горизонталях. Переделать?
func (p *Feature768Provider) FeatureSize() int { return 768 }

func (p *Feature768Provider) ComputeFeatures(pos *common.Position) TuneEntry {
	var input = make([]domain.FeatureInfo, 0, common.PopCount(pos.AllPieces()))
	for x := pos.AllPieces(); x != 0; x &= x - 1 {
		var sq = common.FirstOne(x)
		var pt, side = pos.GetPieceTypeAndSide(sq)
		var piece12 = pt - common.Pawn
		if !side {
			piece12 += 6
		}
		var index = int16(sq ^ piece12<<6)
		input = append(input, domain.FeatureInfo{
			Index: index,
			Value: 1,
		})
	}
	return TuneEntry{
		Features: input,
	}
}
