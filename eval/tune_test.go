package eval

import (
	"testing"

	"github.com/ChizhovVadim/CounterGo/tuner"
)

const tuneFile = "/home/vadim/chess/tuner/quiet-labeled.epd"

func TestTune(t *testing.T) {
	var err = tuner.RunTune(tuneFile, func() tuner.Evaluator {
		return NewEvaluationService()
	})
	if err != nil {
		t.Error(err)
	}
}
