package train

import (
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type EvalService struct {
	featureProvider IFeatureProvider
	model           IModel
}

func (e *EvalService) EvaluateProb(pos *common.Position) float64 {
	var features = e.featureProvider.ComputeFeatures(pos)
	return e.model.Forward(&features)
}

func NewEvalService(fp IFeatureProvider, filepath string) *EvalService {
	var model = NewModel(fp.FeatureSize(), 512)
	var err = model.LoadWeights(filepath)
	if err != nil {
		panic(err)
	}
	return &EvalService{
		featureProvider: fp,
		model:           model,
	}
}
