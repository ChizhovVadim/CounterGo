package shell

import (
	"CounterGo/engine"
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
	Threads            int
	ExperimentSettings bool
	options            []*UciOption
	searchFunc         func(engine.SearchParams) engine.SearchInfo
}

func NewCounterEngine() *CounterEngine {
	var result = &CounterEngine{}
	var onChange = func() {
		result.Reset()
	}
	var numCPUs = runtime.NumCPU()
	result.options = []*UciOption{
		NewIntOption("Hash", &result.Hash, 4, 4, 512, onChange),
		NewIntOption("Threads", &result.Threads, numCPUs, 1, numCPUs, onChange),
		NewBoolOption("ExperimentSettings", &result.ExperimentSettings, false, onChange),
	}
	return result
}

func (eng *CounterEngine) Reset() {
	eng.searchFunc = nil
}

func (*CounterEngine) GetInfo() (name, version, author string) {
	return "Counter", "2.0.1", "Vadim Chizhov"
}

func (eng *CounterEngine) GetOptions() []*UciOption {
	return eng.options
}

func (eng *CounterEngine) Search(searchParams engine.SearchParams) engine.SearchInfo {
	if eng.searchFunc == nil {
		var searchService = &engine.SearchService{
			MoveOrderService:      engine.NewMoveOrderService(),
			Evaluate:              engine.Evaluate,
			DegreeOfParallelism:   eng.Threads,
			UseExperimentSettings: eng.ExperimentSettings,
		}
		if eng.Hash != 0 {
			searchService.TTable = engine.NewTranspositionTable(eng.Hash)
		}
		eng.searchFunc = searchService.Search
	}
	return eng.searchFunc(searchParams)
}
