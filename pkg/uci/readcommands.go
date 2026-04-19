package uci

import (
	"bufio"
	"context"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

func ReadCommands(
	r io.Reader,
	cmds chan<- any,
) error {
	var initPosition, _ = common.NewPositionFromFEN(common.InitialPositionFen)

	var (
		positions = []common.Position{initPosition}
		cancel    context.CancelFunc
	)

	var scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" {
			if cancel != nil {
				cancel()
			}
			return nil
		}
		var commandName, fields = getCommand(commandLine)
		switch commandName {
		case "stop":
			if cancel != nil {
				cancel()
			}
		case "uci":
			cmds <- UciMessage{}
		case "ucinewgame":
			cmds <- NewGameMessage{}
		case "setoption":
			if len(fields) < 4 {
				log.Println("invalid setoption arguments")
			} else {
				cmds <- SetOptionMessage{
					Name:  fields[1],
					Value: fields[3],
				}
			}
		case "isready":
			cmds <- IsReadyMessage{}
		case "position":
			var game = parsePositionCommand(fields)
			if game == nil {
				log.Println("parse position failed")
			} else {
				positions = game
			}
		case "go":
			var limits = parseLimits(fields)
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			cmds <- GoMessage{
				Ctx:    ctx,
				Game:   positions,
				Limits: limits,
			}
		}
	}
	return scanner.Err()
}

func getCommand(commandLine string) (string, []string) {
	if commandLine == "" {
		return "", nil
	}
	var fields = strings.Fields(commandLine)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}

func parsePositionCommand(fields []string) []common.Position {
	var args = fields
	var token = args[0]
	var fen string
	var movesIndex = findIndexString(args, "moves")
	if token == "startpos" {
		fen = common.InitialPositionFen
	} else if token == "fen" {
		if movesIndex == -1 {
			fen = strings.Join(args[1:], " ")
		} else {
			fen = strings.Join(args[1:movesIndex], " ")
		}
	} else {
		return nil
	}
	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return nil
	}
	var positions = []common.Position{p}
	if movesIndex >= 0 && movesIndex+1 < len(args) {
		for _, smove := range args[movesIndex+1:] {
			var newPos, ok = positions[len(positions)-1].MakeMoveLAN(smove)
			if !ok {
				return nil
			}
			positions = append(positions, newPos)
		}
	}
	return positions
}

func findIndexString(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func parseLimits(args []string) (result common.LimitsType) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "ponder":
			result.Ponder = true
		case "wtime":
			result.WhiteTime, _ = strconv.Atoi(args[i+1])
			i++
		case "btime":
			result.BlackTime, _ = strconv.Atoi(args[i+1])
			i++
		case "winc":
			result.WhiteIncrement, _ = strconv.Atoi(args[i+1])
			i++
		case "binc":
			result.BlackIncrement, _ = strconv.Atoi(args[i+1])
			i++
		case "movestogo":
			result.MovesToGo, _ = strconv.Atoi(args[i+1])
			i++
		case "depth":
			result.Depth, _ = strconv.Atoi(args[i+1])
			i++
		case "nodes":
			result.Nodes, _ = strconv.Atoi(args[i+1])
			i++
		case "mate":
			result.Mate, _ = strconv.Atoi(args[i+1])
			i++
		case "movetime":
			result.MoveTime, _ = strconv.Atoi(args[i+1])
			i++
		case "infinite":
			result.Infinite = true
		}
	}
	return
}
