package uci

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/common"
)

// Play a move
func Play(uci *Protocol, fen string, time string) (string, error) {
	var initPosition, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return "", err
	}

	uci.positions = []common.Position{initPosition}
	err = uci.isReadyCommand([]string{})
	if err != nil {
		return "", err
	}

	err = uci.goCommand([]string{"movetime", time})
	if err != nil {
		return "", err
	}
	for {
		searchInfo, ok := <-uci.engineOutput
		if ok {
			fmt.Println(searchInfoToUci(searchInfo))
			if len(searchInfo.MainLine) != 0 {
				uci.bestMove = searchInfo.MainLine[0]
			}
		} else {
			uci.thinking = false
			uci.engineOutput = nil
			return uci.bestMove.String(), nil
		}
	}
}
