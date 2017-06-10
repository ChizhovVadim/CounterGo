package shell

import (
	"CounterGo/engine"
	"fmt"
	"time"
)

type TournamentEngine interface {
	Search(searchParams engine.SearchParams) engine.SearchInfo
}

func NewEngineA() TournamentEngine {
	return &engine.SearchService{
		MoveOrderService:      engine.NewMoveOrderService(),
		Evaluate:              engine.Evaluate,
		TimeControlStrategy:   engine.TimeControlBasic,
		DegreeOfParallelism:   1,
		TTable:                engine.NewTranspositionTable(4),
		UseExperimentSettings: false,
	}
}

func NewEngineB() TournamentEngine {
	return &engine.SearchService{
		MoveOrderService:      engine.NewMoveOrderService(),
		Evaluate:              engine.Evaluate,
		TimeControlStrategy:   engine.TimeControlBasic,
		DegreeOfParallelism:   1,
		TTable:                engine.NewTranspositionTable(4),
		UseExperimentSettings: true,
	}
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
	var playedGames, engine1Wins, engine2Wins int
	for i := 0; i < numberOfGames; i++ {
		var opening = openings[(i/2)%len(openings)]
		var pos = engine.NewPositionFromFEN(opening)

		if (i % 2) == 0 {
			var res = PlayGame(NewEngineA(), NewEngineB(), pos)
			playedGames++
			if res == GameResultWhiteWins {
				engine1Wins++
			} else if res == GameResultBlackWins {
				engine2Wins++
			}
		} else {
			var res = PlayGame(NewEngineB(), NewEngineA(), pos)
			playedGames++
			if res == GameResultWhiteWins {
				engine2Wins++
			} else if res == GameResultBlackWins {
				engine1Wins++
			}
		}

		fmt.Printf("Engine1: %v Engine2: %v Total games: %v\n",
			engine1Wins, engine2Wins, playedGames)
	}
	fmt.Println("Tournament finished.")
}

func PlayGame(engine1, engine2 TournamentEngine, initialPosition *engine.Position) int {
	var positions = []*engine.Position{initialPosition}
	var gameTime = 3 * 60 * 1000
	var limits = engine.LimitsType{
		WhiteTime: gameTime,
		BlackTime: gameTime,
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
		var uciEngine TournamentEngine
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
				return GameResultBlackWins
			}
		} else {
			limits.BlackTime -= elapsed
			if limits.BlackTime < 0 {
				return GameResultWhiteWins
			}
		}
		fmt.Println(searchResult.String())
		fmt.Printf("White: %v Black: %v\n",
			time.Duration(limits.WhiteTime)*time.Microsecond, time.Duration(limits.BlackTime)*time.Microsecond)
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
	var ml = &engine.MoveList{}
	ml.GenerateMoves(position)
	ml.FilterLegalMoves(position)
	if ml.Count == 0 {
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
