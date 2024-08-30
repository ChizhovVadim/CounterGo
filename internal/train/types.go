package train

import (
	"math/rand"

	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/ml"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type Sample struct {
	input  Input
	target float32
}

type Input struct {
	Features []domain.FeatureInfo
}

type IFeatureProvider interface {
	ComputeFeatures(pos *common.Position) Input
	FeatureSize() int
}

type IModel interface {
	InitWeights(rand *rand.Rand)
	LoadWeights(path string) error
	SaveWeights(path string) error
	Forward(input *Input) float64
	Train(sample *Sample, cost ml.IModelCost)
	ApplyGradients()
	AddGradients(mainModel IModel) //for concurrent learning
	Clone() IModel                 //for concurrent learning
}
