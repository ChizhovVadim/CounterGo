package shell

import (
	"counter/engine"
	"runtime"
)

type UciOption struct {
	Name, Type                 string
	BoolValue                  *bool
	BoolDefault                bool
	IntValue                   *int
	IntDefault, IntMin, IntMax int
	OnChange                   func()
}

func NewBoolOption(name string, value *bool, defaultValue bool, onChange func()) *UciOption {
	var option = &UciOption{
		Name:        name,
		Type:        "check",
		BoolValue:   value,
		BoolDefault: defaultValue,
		OnChange:    onChange,
	}
	*value = defaultValue
	return option
}

func NewIntOption(name string, value *int, defaultValue, min, max int, onChange func()) *UciOption {
	var option = &UciOption{
		Name:       name,
		Type:       "spin",
		IntValue:   value,
		IntDefault: defaultValue,
		IntMin:     min,
		IntMax:     max,
		OnChange:   onChange,
	}
	*value = defaultValue
	return option
}

type CounterEngine struct {
	Hash               int
	ExperimentSettings bool
	ParallelSearch     bool
	options            []*UciOption
	searchFunc         func(engine.SearchParams) engine.SearchInfo
}

func NewCounterEngine() *CounterEngine {
	var result = &CounterEngine{}
	var onChange = func() {
		result.searchFunc = nil
	}
	result.options = []*UciOption{
		NewIntOption("Hash", &result.Hash, 4, 4, 512, onChange),
		NewBoolOption("ExperimentSettings", &result.ExperimentSettings, false, onChange),
		NewBoolOption("ParallelSearch", &result.ParallelSearch, true, onChange),
	}
	return result
}

func (*CounterEngine) GetInfo() (name, version, author string) {
	return "Counter", "1.99", "Vadim Chizhov"
}

func (eng *CounterEngine) GetOptions() []*UciOption {
	return eng.options
}

func (eng *CounterEngine) Search(searchParams engine.SearchParams) engine.SearchInfo {
	if eng.searchFunc == nil {
		var searchService = &engine.SearchService{
			MoveOrderService:      engine.NewMoveOrderService(),
			Evaluate:              engine.Evaluate,
			UseExperimentSettings: eng.ExperimentSettings,
		}
		if eng.Hash != 0 {
			searchService.TTable = engine.NewTranspositionTable(eng.Hash)
		}
		if eng.ParallelSearch {
			searchService.DegreeOfParallelism = runtime.NumCPU()
		} else {
			searchService.DegreeOfParallelism = 1
		}
		eng.searchFunc = searchService.Search
	}
	return eng.searchFunc(searchParams)
}
