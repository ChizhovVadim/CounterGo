package main

import (
	"fmt"
	"runtime"

	"github.com/ChizhovVadim/CounterGo/engine"
	"github.com/ChizhovVadim/CounterGo/eval"
	"github.com/ChizhovVadim/CounterGo/uci"
)

/*
Counter Copyright (C) 2017-2020 Vadim Chizhov
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
You should have received a copy of the GNU General Public License along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

var (
	versionName = "dev"
	buildDate   = "(null)"
	gitRevision = "(null)"
)

func main() {
	fmt.Println("Counter",
		"VersionName", versionName,
		"BuildDate", buildDate,
		"GitRevision", gitRevision,
		"RuntimeVersion", runtime.Version())

	uci.Run(engine.NewEngine(func() engine.Evaluator {
		return eval.NewEvaluationService()
	}, versionName))
}
