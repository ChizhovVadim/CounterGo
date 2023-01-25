package main

import (
	"context"
	"log"
	"math"
)

func showResults(
	ctx context.Context,
	gameResults <-chan gameResult,
) error {
	//var totalGames = 2 * len(a.openings)
	var games = 0
	var wins, losses, draws int
	for gameResult := range gameResults {
		games++
		//log.Printf("Finished game %v of %v: %v {%v}\n",
		//	games, totalGames, gameResultString(gameResult.result), gameResult.comment)
		log.Printf("Finished game %v: %v {%v}\n",
			gameResult.gameInfo.gameNumber,
			gameResultString(gameResult.result),
			gameResult.comment)
		if gameResult.result == gameResultDraw {
			draws++
		} else if gameResult.result == gameResultWhiteWins && gameResult.gameInfo.engineAIsWhite ||
			gameResult.result == gameResultBlackWins && !gameResult.gameInfo.engineAIsWhite {
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
