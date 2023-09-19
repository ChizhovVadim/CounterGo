package arena

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func playGame(
	ctx context.Context,
	engineA, engineB IEngine,
	tc TimeControl,
	info gameInfo,
) (gameResult, error) {

	log.Printf("Started game %v\n", info.gameNumber)
	//defer log.Printf("Finished game %v\n", info.gameNumber)

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
		var limits common.LimitsType
		if tc.FixedNodes != 0 {
			limits.Nodes = tc.FixedNodes
		} else if tc.FixedTime != 0 {
			limits.MoveTime = int(tc.FixedTime / time.Millisecond)
		} else {
			panic("bad time control")
		}
		var searchResult = eng.Search(context.Background(), common.SearchParams{
			Positions: positions,
			Limits:    limits,
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

func isLowMaterial(p *common.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!common.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}

	return false
}

func containsMove(ml []common.OrderedMove, move common.Move) bool {
	for i := range ml {
		if ml[i].Move == move {
			return true
		}
	}
	return false
}
