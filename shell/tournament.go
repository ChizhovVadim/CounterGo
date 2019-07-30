package shell

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ChizhovVadim/CounterGo/common"
	"github.com/ChizhovVadim/CounterGo/engine"
)

func NewEngineA() UciEngine {
	var result = engine.NewEngine()
	result.Hash.Value = 64
	result.ExperimentSettings = false
	result.Prepare()
	return result
}

func NewEngineB() UciEngine {
	var result = engine.NewEngine()
	result.Hash.Value = 64
	//result.Threads.Value = 4
	result.ExperimentSettings = true
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
		var pos, err = common.NewPositionFromFEN(opening)
		if err != nil {
			panic(err)
		}

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
		computeStat(engines[1].wins, engines[0].wins, playedGames-engines[1].wins-engines[0].wins)
	}
	fmt.Println("Tournament finished.")
}

//https://www.chessprogramming.org/Match_Statistics
func computeStat(wins, losses, draws int) {
	var games = wins + losses + draws
	var winningFraction = (float64(wins) + 0.5*float64(draws)) / float64(games)
	var eloDifference = -math.Log(1/winningFraction-1) * 400 / math.Ln10
	var los = 0.5 + 0.5*math.Erf(float64(wins-losses)/math.Sqrt(2*float64(wins+losses)))
	fmt.Printf("Winning fraction: %.1f%%\n", winningFraction*100)
	fmt.Printf("Elo difference: %.f\n", eloDifference)
	fmt.Printf("LOS: %.1f%%\n", los*100)
}

//TODO Adjudicate the game as a loss if an engine's score is at least score centipawns below zero for at least count consecutive moves.
func playGame(engine1, engine2 UciEngine, initialPosition common.Position) int {
	engine1.Clear()
	engine2.Clear()
	//var chessClock = MoveTimeChessClock{moveSeconds: 1}
	var chessClock = NewClassicChessClock(2*60, 0, 40)
	var positions = []common.Position{initialPosition}
	for {
		var gameResult = computeGameResult(positions)
		if gameResult != gameResultNone {
			return gameResult
		}
		var searchParams = common.SearchParams{
			Positions: positions,
			Limits:    chessClock.Limits(),
		}
		var side = positions[len(positions)-1].WhiteMove
		var uciEngine UciEngine
		if side {
			uciEngine = engine1
		} else {
			uciEngine = engine2
		}
		chessClock.Start()
		var searchResult = uciEngine.Search(context.Background(), searchParams)
		chessClock.Stop()
		PrintSearchInfo(searchResult)
		fmt.Println(chessClock)
		var move = searchResult.MainLine[0]
		var newPos = common.Position{}
		var ok = positions[len(positions)-1].MakeMove(move, &newPos)
		if !ok {
			panic("engine illegal move")
		}
		positions = append(positions, newPos)
		fmt.Println(&newPos)
		PrintPosition(&newPos)
	}
}

func computeGameResult(positions []common.Position) int {
	var position = &positions[len(positions)-1]
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

func isRepetition(positions []common.Position) bool {
	var current = &positions[len(positions)-1]
	var repeats = 0

	for i := len(positions) - 2; i >= 0; i-- {
		if current.IsRepetition(&positions[i]) {
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

const Second = 1000

type ChessClock interface {
	Start()
	Stop()
	Limits() common.LimitsType
}

type MoveTimeChessClock struct {
	moveSeconds int
}

func (c *MoveTimeChessClock) Start() {}

func (c *MoveTimeChessClock) Stop() {}

func (c *MoveTimeChessClock) Limits() common.LimitsType {
	return common.LimitsType{MoveTime: c.moveSeconds * Second}
}

type ClassicChessClock struct {
	gameSeconds int
	movesToGo   int
	limits      common.LimitsType
	side        bool
	start       time.Time
}

func NewClassicChessClock(gameSeconds, incrementSeconds, movesToGo int) *ClassicChessClock {
	return &ClassicChessClock{
		gameSeconds: gameSeconds,
		movesToGo:   movesToGo,
		limits: common.LimitsType{
			WhiteTime:      gameSeconds * Second,
			WhiteIncrement: incrementSeconds * Second,
			BlackTime:      gameSeconds * Second,
			BlackIncrement: incrementSeconds * Second,
			MovesToGo:      movesToGo,
		},
		side: true,
	}
}

func (c *ClassicChessClock) String() string {
	return fmt.Sprintf("White: %v Black: %v",
		c.limits.WhiteTime, c.limits.BlackTime)
}

func (c *ClassicChessClock) Start() {
	c.start = time.Now()
}

func (c *ClassicChessClock) Stop() {
	var elapsed = int(time.Since(c.start) / time.Millisecond)
	if c.side {
		c.limits.WhiteTime -= elapsed
		if c.limits.WhiteTime < 0 {
			panic(fmt.Errorf("white dropped the flag"))
		}
		c.limits.WhiteTime += c.limits.WhiteIncrement
	} else {
		c.limits.BlackTime -= elapsed
		if c.limits.BlackTime < 0 {
			panic(fmt.Errorf("black dropped the flag"))
		}
		c.limits.BlackTime += c.limits.BlackIncrement
		if c.movesToGo != 0 {
			c.limits.MovesToGo--
			if c.limits.MovesToGo == 0 {
				c.limits.WhiteTime = c.gameSeconds * Second
				c.limits.BlackTime = c.gameSeconds * Second
				c.limits.MovesToGo = c.movesToGo
			}
		}
	}
	c.side = !c.side
}

func (c *ClassicChessClock) Limits() common.LimitsType {
	return c.limits
}
