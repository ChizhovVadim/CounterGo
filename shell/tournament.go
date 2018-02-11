package shell

import (
	"fmt"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
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
	gameResultNone = iota
	gameResultWhiteWins
	gameResultBlackWins
	gameResultDraw
)

var openings = []string{
	"rn1q1rk1/1p2ppbp/p1p2np1/3p4/2PP2b1/1PNBPN2/P4PPP/R1BQ1RK1 w - - 1 9 ",
	"r1b1kb1r/1pq2ppp/p1nppn2/8/3NP1P1/2N4P/PPP2PB1/R1BQK2R w KQkq - 2 9 ",
	"r1bq1rk1/pp1nppbp/3p1np1/8/P2p1B2/4PN1P/1PP1BPP1/RN1Q1RK1 w - - 0 9 ",
	"r1bqk2r/p3bpp1/1pn1pn1p/2pp4/3P3B/2PBPN2/PP1N1PPP/R2QK2R w KQkq - 0 9 ",
	"r2qk2r/p1pp1ppp/b1p2n2/8/2P5/6P1/PP1QPP1P/RN2KB1R w KQkq - 1 9 ",
	"r1bqr1k1/pppp1ppp/2n2n2/2bN4/2P1p2N/6P1/PP1PPPBP/R1BQ1RK1 w - - 6 9 ",
	"r1bq1rk1/1p3ppp/2n1pn2/p1bp4/2P5/P3PN2/1P1NBPPP/R1BQK2R w KQ - 2 9 ",
	"r1bq1rk1/ppp1p1bp/3p1np1/n2P1p2/2P5/2N2NP1/PP2PPBP/R1BQ1RK1 w - - 1 9 ",
	"r1bq1rk1/ppp2ppp/2np1n2/4p3/2P5/2PP1NP1/P3PPBP/R1BQ1RK1 w - - 1 9",
	"r1bqk1nr/1pp2pbp/p2p2p1/1N1P4/2PpP3/8/PP3PPP/R1BQKB1R w KQkq - 0 9 ",
	"rn1qkb1r/pbpp1p2/1p2p2p/6p1/2PP4/2N1PNn1/PP3PPP/R2QKB1R w KQkq - 0 9 ",
	"rn1qk2r/pp2bppp/2p2nb1/3p4/3N4/3P2P1/PP2PPBP/RNBQ1RK1 w kq - 3 9 ",
	"rn1q1rk1/pbp1bppp/1p1pp3/7n/2PP4/2N1PNB1/PPQ2PPP/R3KB1R w KQ - 0 9 ",
	"r1bq1rk1/bpp2ppp/p1np1n2/4p3/B3P3/2PP1N2/PP1N1PPP/R1BQ1RK1 w - - 2 9 ",
	"r3kbnr/pp1b1ppp/1q2p3/3pP3/3n4/2P2N2/PP3PPP/R1BQKB1R w KQkq - 0 9 ",
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
		var pos = common.NewPositionFromFEN(opening)

		var whiteEngineIndex = i % 2
		var blackEngineIndex = whiteEngineIndex ^ 1
		var res = playGame(engines[whiteEngineIndex].engine,
			engines[blackEngineIndex].engine, pos)
		playedGames++
		if res == gameResultWhiteWins {
			engines[whiteEngineIndex].wins++
		} else if res == gameResultBlackWins {
			engines[blackEngineIndex].wins++
		}

		fmt.Printf("Engine1: %v Engine2: %v Total games: %v\n",
			engines[0].wins, engines[1].wins, playedGames)
	}
	fmt.Println("Tournament finished.")
}

func playGame(engine1, engine2 UciEngine, initialPosition *common.Position) int {
	const Second = 1000
	var timeControl = struct {
		main, inc, moves int
		//}{60 * Second, 1 * Second, 0}
	}{60 * Second, 0 * Second, 40}
	var positions = []*common.Position{initialPosition}
	var limits = common.LimitsType{
		WhiteTime: timeControl.main,
		BlackTime: timeControl.main,
		MovesToGo: timeControl.moves,
	}
	for {
		var gameResult = computeGameResult(positions)
		if gameResult != gameResultNone {
			return gameResult
		}
		var searchParams = common.SearchParams{
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
		var elapsed = int(time.Since(start) / time.Millisecond)
		if side {
			limits.WhiteTime -= elapsed
			if limits.WhiteTime < 0 {
				return gameResultBlackWins
			}
			limits.WhiteTime += timeControl.inc
		} else {
			limits.BlackTime -= elapsed
			if limits.BlackTime < 0 {
				return gameResultWhiteWins
			}
			limits.BlackTime += timeControl.inc
			if timeControl.moves > 0 {
				limits.MovesToGo--
				if limits.MovesToGo == 0 {
					limits.WhiteTime = timeControl.main
					limits.BlackTime = timeControl.main
					limits.MovesToGo = timeControl.moves
				}
			}
		}
		PrintSearchInfo(searchResult)
		fmt.Printf("White: %v Black: %v\n", limits.WhiteTime, limits.BlackTime)
		var move = searchResult.MainLine[0]
		var newPos = &common.Position{}
		var ok = positions[len(positions)-1].MakeMove(move, newPos)
		if !ok {
			panic("engine illegal move")
		}
		positions = append(positions, newPos)
		fmt.Println(newPos)
		PrintPosition(newPos)
	}
}

func computeGameResult(positions []*common.Position) int {
	var position = positions[len(positions)-1]
	var ml = position.GenerateLegalMoves()
	if len(ml) == 0 {
		if !position.IsCheck() {
			return gameResultDraw
		} else if position.WhiteMove {
			return gameResultBlackWins
		} else {
			return gameResultWhiteWins
		}
	} else if position.Rule50 >= 100 || isRepetition(positions) {
		return gameResultDraw
	}
	return gameResultNone
}

func isRepetition(positions []*common.Position) bool {
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
