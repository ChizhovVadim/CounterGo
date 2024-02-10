package dataset

import (
	"fmt"
	"math"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type DatasetItem2 struct {
	Pos          *common.Position
	Target       float64
	GameResult   float64
	SearchResult common.UciScore
}

func AnalyzeGame(
	sigmoidScale float64,
	searchRatio float64,
	gameRaw pgn.GameRaw,
	onPosition func(DatasetItem2) error,
) error {
	var game, err = pgn.ParseGame(gameRaw)
	if err != nil {
		return err
	}
	var gameResult, gameResOk = calcGameResult(game.Result)
	if !gameResOk {
		return fmt.Errorf("bad game result %v", game.Result)
	}

	var startFen = game.Fen
	if startFen == "" {
		startFen = common.InitialPositionFen
	}
	pos, err := common.NewPositionFromFEN(startFen)
	if err != nil {
		return err
	}

	var repeatPositions = make(map[uint64]struct{})

	for i := range game.Items {
		//filter quiet positions
		var move = game.Items[i].Move
		var comment = game.Items[i].Comment

		_, found := repeatPositions[pos.Key]
		if !(found ||
			pos.IsCheck() ||
			pos.Rule50 >= 50 ||
			isCaptureOrPromotion(move) ||
			comment.Depth < 8 ||
			isDraw(&pos)) {

			var searchResult common.UciScore
			if pos.WhiteMove {
				searchResult = comment.Score
			} else {
				searchResult = common.UciScore{
					Mate:       -comment.Score.Mate,
					Centipawns: -comment.Score.Centipawns,
				}
			}

			var targetBySearch float64
			if comment.Score.Mate != 0 {
				if comment.Score.Mate > 0 {
					targetBySearch = 1
				} else {
					targetBySearch = 0
				}
			} else {
				targetBySearch = sigmoid(float64(comment.Score.Centipawns), sigmoidScale)
			}
			if !pos.WhiteMove {
				targetBySearch = 1 - targetBySearch
			}

			var target = targetBySearch*searchRatio + gameResult*(1-searchRatio)
			var err = onPosition(DatasetItem2{
				Pos:          &pos,
				Target:       target,
				GameResult:   gameResult,
				SearchResult: searchResult,
			})
			if err != nil {
				return err
			}
		}

		repeatPositions[pos.Key] = struct{}{}

		//Make move
		var child common.Position
		if !pos.MakeMove(game.Items[i].Move, &child) {
			break
		}
		pos = child
	}

	return nil
}

func calcGameResult(sGameResult string) (float64, bool) {
	switch sGameResult {
	case pgn.GameResultWhiteWin:
		return 1, true
	case pgn.GameResultBlackWin:
		return 0, true
	case pgn.GameResultDraw:
		return 0.5, true
	default:
		return 0, false
	}
}

func isCaptureOrPromotion(move common.Move) bool {
	return move.CapturedPiece() != common.Empty ||
		move.Promotion() != common.Empty
}

func isDraw(p *common.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!common.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}
	return false
}

func sigmoid(x, sigmoidScale float64) float64 {
	return 1.0 / (1.0 + math.Exp(sigmoidScale*(-x)))
}
