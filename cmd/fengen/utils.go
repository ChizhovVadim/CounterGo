package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func pgnFiles(folderPath string) ([]string, error) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pgn" {
			result = append(result, filepath.Join(folderPath, file.Name()))
		}
	}
	return result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var errGameResultFail = errors.New("game result fail")
var errGameResultNone = errors.New("game result none")

func calcGameResult(game *pgn.Game) (float32, error) {
	var sGameResult, gameResultOk = game.TagValue("Result")
	if !gameResultOk {
		return 0, errGameResultFail
	}

	var result float32
	switch sGameResult {
	case pgn.GameResultNone:
		return 0, errGameResultNone
	case pgn.GameResultWhiteWin:
		result = 1
	case pgn.GameResultBlackWin:
		result = 0
	case pgn.GameResultDraw:
		result = 0.5
	default:
		return 0, errGameResultFail
	}

	return result, nil
}

func hasLegalMove(p *common.Position) bool {
	var buf [common.MaxMoves]common.OrderedMove
	var ml = p.GenerateMoves(buf[:])
	var child common.Position
	for i := range ml {
		if !p.MakeMove(ml[i].Move, &child) {
			continue
		}
		return true
	}
	return false
}
