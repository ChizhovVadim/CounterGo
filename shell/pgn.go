package shell

import (
	"bufio"
	"io"
	"strings"

	"github.com/ChizhovVadim/CounterGo/common"
)

type PgnGame struct {
	Moves []common.Move
}

func LoadPgn(r io.Reader) []PgnGame {
	var games []PgnGame
	var pos, child common.Position
	var scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		var line = scanner.Text()
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "[Event") {
			games = append(games, PgnGame{})
			pos, _ = common.NewPositionFromFEN(common.InitialPositionFen)
		}
		if strings.HasPrefix(line, "[") {
			continue
		}
		var tokens = strings.FieldsFunc(line, func(ch rune) bool {
			return ch == '.' || ch == ' '
		})
		for _, tk := range tokens {
			if tk == "1-0" || tk == "0-1" || tk == "1/2-1/2" || tk == "*" {
				break
			}
			if !canBeMove(tk) {
				continue
			}
			var mv = common.ParseMoveSAN(&pos, tk)
			if mv != common.MoveEmpty &&
				pos.MakeMove(mv, &child) {
				var game = &games[len(games)-1]
				game.Moves = append(game.Moves, mv)
				pos = child
			}
		}
	}
	return games
}

func canBeMove(s string) bool {
	return -1 == strings.IndexFunc(s, func(ch rune) bool {
		return !strings.ContainsRune("12345678abcdefghNBRQKOxnbrq=-+#!?", ch)
	})
}
