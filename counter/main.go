package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/engine"
	counter "github.com/ChizhovVadim/CounterGo/eval/counter"
	"github.com/ChizhovVadim/CounterGo/uci"
	//pesto "github.com/ChizhovVadim/CounterGo/eval/pesto"
	//weiss "github.com/ChizhovVadim/CounterGo/eval/weiss"
)

/*
Counter Copyright (C) 2017-2022 Vadim Chizhov
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
You should have received a copy of the GNU General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

const (
	name   = "Counter"
	author = "Vadim Chizhov"
)

var (
	versionName = "dev"
	buildDate   = "(null)"
	gitRevision = "(null)"
	flgEval     string
)

func main() {
	flag.StringVar(&flgEval, "eval", "counter", "specifies evaluation function")
	flag.Parse()

	var logger = log.New(os.Stderr, "", log.LstdFlags)

	logger.Println(name,
		"VersionName", versionName,
		"BuildDate", buildDate,
		"GitRevision", gitRevision,
		"RuntimeVersion", runtime.Version())

	var evalBuilder = evalBuilder(flgEval)
	var engine = engine.NewEngine(evalBuilder)

	var protocol = uci.New(name, author, versionName, engine,
		[]uci.Option{
			&uci.IntOption{Name: "Hash", Min: 4, Max: 1 << 16, Value: &engine.Hash},
			&uci.IntOption{Name: "Threads", Min: 1, Max: runtime.NumCPU(), Value: &engine.Threads},
			&uci.BoolOption{Name: "ExperimentSettings", Value: &engine.ExperimentSettings},
		},
	)

	uci.RunCli(logger, protocol)
}

func evalBuilder(name string) func() engine.Evaluator {
	return func() engine.Evaluator {
		if name == "counter" {
			return counter.NewEvaluationService()
		}
		/*if name == "pesto" {
			return pesto.NewEvaluationService()
		}
		if name == "weiss" {
			return weiss.NewEvaluationService()
		}*/
		panic(fmt.Errorf("bad eval %v", name))
	}
}
