package tuner

import (
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Sample struct {
	domain.TuneEntry
	Target float32
}

type IFeatureProvider interface {
	ComputeFeatures(pos *common.Position) domain.TuneEntry
	FeatureSize() int
}
