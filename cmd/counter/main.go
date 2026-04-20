package main

import (
	"log"
	"runtime"
	"runtime/debug"

	"github.com/ChizhovVadim/CounterGo/pkg/engine"
	"github.com/ChizhovVadim/CounterGo/pkg/evalnn"
	"github.com/ChizhovVadim/CounterGo/pkg/uci"
)

/*
Counter Copyright (C) 2017-2026 Vadim Chizhov
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
You should have received a copy of the GNU General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

const (
	name        = "Counter"
	author      = "Vadim Chizhov"
	versionName = "5.5"
)

func main() {
	if info, ok := debug.ReadBuildInfo(); ok {
		log.Println("BuildInfo",
			"Path", info.Main.Path,
			"Version", info.Main.Version,
			"GoVersion", info.GoVersion,
		)
	}

	var weights, err = evalnn.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded nnue weights")
	var config = engine.NewConfig(func() engine.IUpdatableEvaluator {
		return evalnn.NewEvaluationService(weights, 1.0)
	})
	var eng = engine.New(config)
	var engAgent = uci.NewEngineAgent(name, author, versionName, eng,
		[]uci.Option{
			&uci.IntOption{Name: "Hash", Min: 4, Max: 1 << 16, Value: &eng.Config.Hash},
			&uci.IntOption{Name: "Threads", Min: 1, Max: runtime.NumCPU(), Value: &eng.Config.Threads},
			&uci.BoolOption{Name: "ExperimentSettings", Value: &eng.Config.ExperimentSettings},
		},
	)
	uci.Run(engAgent)
}
