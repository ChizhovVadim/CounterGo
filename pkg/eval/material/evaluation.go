package eval

import (
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type EvaluationService struct{}

func NewEvaluationService() *EvaluationService {
	return &EvaluationService{}
}

func (e *EvaluationService) Evaluate(p *common.Position) int {
	var eval = 100*(common.PopCount(p.Pawns&p.White)-common.PopCount(p.Pawns&p.Black)) +
		400*(common.PopCount(p.Knights&p.White)-common.PopCount(p.Knights&p.Black)) +
		400*(common.PopCount(p.Bishops&p.White)-common.PopCount(p.Bishops&p.Black)) +
		600*(common.PopCount(p.Rooks&p.White)-common.PopCount(p.Rooks&p.Black)) +
		1200*(common.PopCount(p.Queens&p.White)-common.PopCount(p.Queens&p.Black))
	if !p.WhiteMove {
		eval = -eval
	}
	return eval
}
