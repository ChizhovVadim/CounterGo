package shell

import (
	"fmt"
	"time"

	"github.com/ChizhovVadim/CounterGo/engine"
)

func NewEngineA() UciEngine {
	var result = engine.NewEngine()
	result.Hash.Value = 16
	result.ExperimentSettings.Value = false
	result.Threads.Value = 1
	result.ClearTransTable = true
	result.Prepare()
	return result
}

func NewEngineB() UciEngine {
	var result = engine.NewEngine()
	result.Hash.Value = 16
	result.ExperimentSettings.Value = true
	result.Threads.Value = 1
	result.ClearTransTable = true
	result.Prepare()
	return result
}

const (
	GameResultNone = iota
	GameResultWhiteWins
	GameResultBlackWins
	GameResultDraw
)

var openings = []string{
	"rnbqkb1r/ppp2ppp/3p1n2/4N3/4P3/8/PPPP1PPP/RNBQKB1R w KQkq - 0 4",
	"r1bqk1nr/pppp1ppp/2n5/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
	"rnbqk2r/pppp1ppp/4pn2/8/1bPP4/2N5/PP2PPPP/R1BQKBNR w KQkq - 2 4",
	"rnbqkbnr/pp2pppp/2p5/8/3Pp3/2N5/PPP2PPP/R1BQKBNR w KQkq - 0 4",
	"rnbqkb1r/pp2pppp/2p2n2/3p4/2PP4/5N2/PP2PPPP/RNBQKB1R w KQkq - 2 4",
}

func RunTournament() {
	fmt.Println("Tournament started...")
	var numberOfGames = len(openings) * 2
	var playedGames = 0
	var engines = []struct {
		engine UciEngine
		wins   int
	}{
		{NewEngineA(), 0},
		{NewEngineB(), 0},
	}
	for i := 0; i < numberOfGames; i++ {
		var opening = openings[(i/2)%len(openings)]
		var pos = engine.NewPositionFromFEN(opening)

		var whiteEngineIndex = i % 2
		var blackEngineIndex = whiteEngineIndex ^ 1
		var res = PlayGame(engines[whiteEngineIndex].engine,
			engines[blackEngineIndex].engine, pos)
		playedGames++
		if res == GameResultWhiteWins {
			engines[whiteEngineIndex].wins++
		} else if res == GameResultBlackWins {
			engines[blackEngineIndex].wins++
		}

		fmt.Printf("Engine1: %v Engine2: %v Total games: %v\n",
			engines[0].wins, engines[1].wins, playedGames)
	}
	fmt.Println("Tournament finished.")
}

func PlayGame(engine1, engine2 UciEngine, initialPosition *engine.Position) int {
	var positions = []*engine.Position{initialPosition}
	var gameTime, isNodeLimits = 2 * 60 * 1000, false
	//var gameTime, isNodeLimits = 100 * 1000 * 1000, true
	var limits = engine.LimitsType{
		WhiteTime:    gameTime,
		BlackTime:    gameTime,
		IsNodeLimits: isNodeLimits,
	}
	for {
		var gameResult = ComputeGameResult(positions)
		if gameResult != GameResultNone {
			return gameResult
		}
		var searchParams = engine.SearchParams{
			Positions: positions,
			Limits:    limits,
		}
		var side = positions[len(positions)-1].WhiteMove
		var uciEngine UciEngine
		if side {
			uciEngine = engine1
		} else {
			uciEngine = engine2
		}
		var start = time.Now()
		var searchResult = uciEngine.Search(searchParams)
		var elapsed int
		if limits.IsNodeLimits {
			elapsed = int(searchResult.Nodes)
		} else {
			elapsed = int(time.Since(start) / time.Millisecond)
		}
		if side {
			limits.WhiteTime -= elapsed
			if limits.WhiteTime < 0 {
				return GameResultBlackWins
			}
		} else {
			limits.BlackTime -= elapsed
			if limits.BlackTime < 0 {
				return GameResultWhiteWins
			}
		}
		fmt.Println(searchResult.String())
		fmt.Printf("White: %v Black: %v\n", limits.WhiteTime, limits.BlackTime)
		var move = searchResult.MainLine[0]
		var newPos = &engine.Position{}
		var ok = positions[len(positions)-1].MakeMove(move, newPos)
		if !ok {
			panic("engine illegal move")
		}
		positions = append(positions, newPos)
		fmt.Println(newPos)
		PrintPosition(newPos)
	}
}

func ComputeGameResult(positions []*engine.Position) int {
	var position = positions[len(positions)-1]
	var ml = engine.GenerateLegalMoves(position)
	if len(ml) == 0 {
		if !position.IsCheck() {
			return GameResultDraw
		} else if position.WhiteMove {
			return GameResultBlackWins
		} else {
			return GameResultWhiteWins
		}
	} else if position.Rule50 >= 100 || IsRepetition(positions) {
		return GameResultDraw
	}
	return GameResultNone
}

func IsRepetition(positions []*engine.Position) bool {
	var current = positions[len(positions)-1]
	var repeats = 0

	for i := len(positions) - 2; i >= 0; i-- {
		if current.IsRepetition(positions[i]) {
			repeats++
			if repeats >= 3 {
				return true
			}
		}
		if positions[i].Rule50 == 0 {
			break
		}
	}

	return false
}
