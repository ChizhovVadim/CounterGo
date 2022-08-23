package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"golang.org/x/sync/errgroup"
)

const (
	gameResultDraw = iota
	gameResultWhiteWins
	gameResultBlackWins
)

type IEngine interface {
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

type arena struct {
	threads  int
	openings []string
}

type gameInfo struct {
	opening        string
	engineAIsWhite bool
}

type gameResult struct {
	gameInfo
	positions []common.Position
	comment   string
	result    int
}

func (a *arena) Run(ctx context.Context) error {
	log.Println("arena started")
	defer log.Println("arena finished")

	g, ctx := errgroup.WithContext(ctx)

	var gameInfos = make(chan gameInfo, 16)
	var gameResults = make(chan gameResult, 16)

	g.Go(func() error {
		defer close(gameInfos)
		return a.fillGameInfos(ctx, gameInfos)
	})

	g.Go(func() error {
		return a.saveGames(ctx, gameResults)
	})

	var wg = &sync.WaitGroup{}

	for i := 0; i < a.threads; i++ {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return a.playGames(ctx, gameInfos, gameResults)
		})
	}

	g.Go(func() error {
		wg.Wait()
		close(gameResults)
		return nil
	})

	return g.Wait()
}

func (a *arena) fillGameInfos(ctx context.Context,
	gameInfos chan<- gameInfo) error {
	for _, opening := range a.openings {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameInfos <- gameInfo{opening: opening, engineAIsWhite: true}:
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameInfos <- gameInfo{opening: opening, engineAIsWhite: false}:
		}
	}
	return nil
}

func (a *arena) playGames(
	ctx context.Context,
	gameInfos <-chan gameInfo,
	gameResults chan<- gameResult,
) error {
	var engineA = newEngineA()
	var engineB = newEngineB()
	for gameInfo := range gameInfos {
		var res, err = a.playGame(engineA, engineB, gameInfo)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameResults <- res:
		}
	}
	return nil
}

func (a *arena) saveGames(
	ctx context.Context,
	gameResults <-chan gameResult,
) error {
	var totalGames = 2 * len(a.openings)
	var games = 0
	var wins, losses, draws int
	for gameResult := range gameResults {
		games++
		log.Printf("Finished game %v of %v: %v {%v}\n",
			games, totalGames, gameResultString(gameResult.result), gameResult.comment)
		if gameResult.result == gameResultDraw {
			draws++
		} else if gameResult.result == gameResultWhiteWins && gameResult.engineAIsWhite ||
			gameResult.result == gameResultBlackWins && !gameResult.engineAIsWhite {
			wins++
		} else {
			losses++
		}
		var stat = computeStat(wins, losses, draws)
		log.Printf("Score: %v - %v - %v  [%.3f] %v\n",
			wins, losses, draws, stat.winningFraction, games)
		log.Printf("Elo difference: %.1f, LOS: %.1f %%\n",
			stat.eloDifference, stat.los*100)
	}
	return nil
}

// TODO timecontrol, draw-resign-endgame interrupt
func (a *arena) playGame(engineA, engineB IEngine, info gameInfo) (gameResult, error) {
	engineA.Clear()
	engineB.Clear()
	var startingPos, err = common.NewPositionFromFEN(info.opening)
	if err != nil {
		return gameResult{}, err
	}
	var positions []common.Position
	positions = append(positions, startingPos)
	var keys = make(map[uint64]int)
	var buf [common.MaxMoves]common.OrderedMove
	var child common.Position
	for {
		var curPosition = &positions[len(positions)-1]
		var ml = curPosition.GenerateMoves(buf[:])
		var hasLegalMove = false
		for i := range ml {
			if !curPosition.MakeMove(ml[i].Move, &child) {
				continue
			}
			hasLegalMove = true
			break
		}
		if !hasLegalMove {
			if curPosition.IsCheck() {
				var points int
				if curPosition.WhiteMove {
					points = gameResultBlackWins
				} else {
					points = gameResultWhiteWins
				}
				return gameResult{gameInfo: info, positions: positions, comment: "checkmate", result: points}, nil
			} else {
				return gameResult{gameInfo: info, positions: positions, comment: "stalemate", result: gameResultDraw}, nil
			}
		}
		if curPosition.Rule50 >= 100 {
			return gameResult{gameInfo: info, positions: positions, comment: "50 moves", result: gameResultDraw}, nil
		}
		if isLowMaterial(curPosition) {
			return gameResult{gameInfo: info, positions: positions, comment: "low material", result: gameResultDraw}, nil
		}
		keys[curPosition.Key] += 1
		if keys[curPosition.Key] == 3 {
			return gameResult{gameInfo: info, positions: positions, comment: "3 fold repetition", result: gameResultDraw}, nil
		}
		var eng IEngine
		if curPosition.WhiteMove == info.engineAIsWhite {
			eng = engineA
		} else {
			eng = engineB
		}
		var searchResult = eng.Search(context.Background(), common.SearchParams{
			Positions: positions,
			Limits:    common.LimitsType{Nodes: 500000 /*MoveTime: 300*/},
		})
		var bestMove = searchResult.MainLine[0]
		if !containsMove(ml, bestMove) {
			return gameResult{}, fmt.Errorf("bad move")
		}
		if !curPosition.MakeMove(bestMove, &child) {
			return gameResult{}, fmt.Errorf("bad move")
		}
		positions = append(positions, child)
	}
}

func containsMove(ml []common.OrderedMove, move common.Move) bool {
	for i := range ml {
		if ml[i].Move == move {
			return true
		}
	}
	return false
}

type GameStatistics struct {
	winningFraction float64
	eloDifference   float64
	los             float64
}

//https://chessprogramming.wikispaces.com/Match%20Statistics
func computeStat(wins, losses, draws int) GameStatistics {
	var games = wins + losses + draws
	var winning_fraction = (float64(wins) + 0.5*float64(draws)) / float64(games)
	var elo_difference = -math.Log(1/winning_fraction-1) * 400 / math.Ln10
	var los = 0.5 + 0.5*math.Erf(float64(wins-losses)/math.Sqrt(2*float64(wins+losses)))
	return GameStatistics{
		winningFraction: winning_fraction,
		eloDifference:   elo_difference,
		los:             los,
	}
}

func isLowMaterial(p *common.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!common.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	return false
}

func gameResultString(v int) string {
	if v == gameResultWhiteWins {
		return "1-0"
	}
	if v == gameResultBlackWins {
		return "0-1"
	}
	if v == gameResultDraw {
		return "1/2-1/2"
	}
	return ""
}
