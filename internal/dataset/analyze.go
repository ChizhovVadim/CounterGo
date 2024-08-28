package dataset

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type GameInfo struct {
	GameResult float64
	Positions  []PositionInfo
}

type PositionInfo struct {
	ScoreMate       int
	ScoreCentipawns int
	Position        common.Position
}

func AnalyzeGame(gameRaw pgn.GameRaw) (GameInfo, error) {
	var game, err = pgn.ParseGame(gameRaw)
	if err != nil {
		return GameInfo{}, err
	}
	var gameResult, gameResOk = calcGameResult(game.Result)
	if !gameResOk {
		return GameInfo{}, fmt.Errorf("bad game result %v", game.Result)
	}
	var startFen = game.Fen
	if startFen == "" {
		startFen = common.InitialPositionFen
	}
	pos, err := common.NewPositionFromFEN(startFen)
	if err != nil {
		return GameInfo{}, err
	}

	var repeatPositions = make(map[uint64]struct{})
	var posInfos []PositionInfo

	for i := range game.Items {
		var move = game.Items[i].Move
		var comment = game.Items[i].Comment

		_, found := repeatPositions[pos.Key]
		//filter repeats and noisy positions
		if !(found ||
			ignorePos(&pos, move, comment)) {

			posInfos = append(posInfos, PositionInfo{
				ScoreMate:       comment.Score.Mate,
				ScoreCentipawns: comment.Score.Centipawns,
				Position:        pos,
			})
		}

		repeatPositions[pos.Key] = struct{}{}

		//Make move
		var child common.Position
		if !pos.MakeMove(move, &child) {
			break
		}
		pos = child
	}

	return GameInfo{
		GameResult: gameResult,
		Positions:  posInfos,
	}, nil
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

func ignorePos(
	pos *common.Position,
	searchBestMove common.Move,
	searchComment pgn.Comment,
) bool {
	return searchComment.Depth == 0 ||
		searchComment.Score.Mate != 0 ||
		pos.IsCheck() ||
		searchBestMove.CapturedPiece() != common.Empty
}
