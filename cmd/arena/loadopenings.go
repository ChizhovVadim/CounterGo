package main

import (
	"context"
	_ "embed"
	"strings"

	"github.com/ChizhovVadim/CounterGo/internal/pgn"
)

//go:embed openings.txt
var openingsTxt string

func loadOpenings(
	ctx context.Context,
	gameInfos chan<- gameInfo,
) error {

	var openings = getOpenings()

	for i, opening := range openings {
		var g, err = pgn.ParseGame(opening)
		if err != nil {
			return err
		}
		var fen = g.Items[len(g.Items)-1].Position.String()
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
