package arena

import (
	"context"
	_ "embed"
	"strings"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

//go:embed openings.txt
var openingsTxt string

func loadOpenings(
	ctx context.Context,
	gameInfos chan<- gameInfo,
) error {

	var openings = getOpenings()

	for i, opening := range openings {
		var fen, err = parseOpening(opening)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameInfos <- gameInfo{opening: fen, engineAIsWhite: true, gameNumber: 1 + 2*i}:
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case gameInfos <- gameInfo{opening: fen, engineAIsWhite: false, gameNumber: 1 + 2*i + 1}:
		}
	}

	return nil
}

func parseOpening(opening string) (string, error) {
	var g, err = pgn.ParseGame(pgn.GameRaw{
		Tags: []pgn.Tag{
			{Key: "Result", Value: "*"},
		},
		BodyRaw: opening,
	})
	if err != nil {
		return "", err
	}
	pos, err := common.NewPositionFromFEN(common.InitialPositionFen)
	if err != nil {
		return "", err
	}
	for i := range g.Items {
		var move = g.Items[i].Move
		var child common.Position
		pos.MakeMove(move, &child)
		pos = child
	}
	return pos.String(), nil
}

func getOpenings() []string {
	var result []string
	var lines = strings.Split(openingsTxt, "\n")
	for _, line := range lines {
		if !(line == "" || strings.HasPrefix(line, "//")) {
			result = append(result, line)
		}
	}
	return result
}
