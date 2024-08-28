package train

import (
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Sample struct {
	TuneEntry
	Target float32
}

type TuneEntry struct {
	Features []domain.FeatureInfo
}

type IFeatureProvider interface {
	ComputeFeatures(pos *common.Position) TuneEntry
	FeatureSize() int
}
