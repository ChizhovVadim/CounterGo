package uci

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/ChizhovVadim/CounterGo/common"
)

type Engine interface {
	Prepare()
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

type Protocol struct {
	name      string
	author    string
	version   string
	options   []Option
	engine    Engine
	positions []common.Position
	thinking  int32
	cancel    context.CancelFunc
}

func New(name, author, version string, engine Engine, options []Option) *Protocol {
	var initPosition, err = common.NewPositionFromFEN(common.InitialPositionFen)
	if err != nil {
		panic(err)
	}
	return &Protocol{
		name:      name,
		author:    author,
		version:   version,
		engine:    engine,
		options:   options,
		positions: []common.Position{initPosition},
	}
}

func (uci *Protocol) Handle(ctx context.Context, commandLine string) error {
	var fields = strings.Fields(commandLine)
	if len(fields) == 0 {
		return nil
	}
	var commandName = fields[0]
	fields = fields[1:]

	if atomic.LoadInt32(&uci.thinking) == 1 {
		if commandName == "stop" {
			uci.cancel()
			return nil
		}
		return errors.New("search still run")
	}

	var h func(fields []string) error

	switch commandName {
	case "uci":
		h = uci.uciCommand
	case "setoption":
		h = uci.setOptionCommand
	case "isready":
		h = uci.isReadyCommand
	case "position":
		h = uci.positionCommand
	case "go":
		h = uci.goCommand
	case "ucinewgame":
		h = uci.uciNewGameCommand
	case "ponderhit":
		h = uci.ponderhitCommand
	}

	if h == nil {
		return errors.New("command not found")
	}

	return h(fields)
}

func (uci *Protocol) uciCommand(fields []string) error {
	fmt.Printf("id name %s %s\n", uci.name, uci.version)
	fmt.Printf("id author %s\n", uci.author)
	for _, option := range uci.options {
		fmt.Println(option.UciString())
	}
	fmt.Println("uciok")
	return nil
}

func (uci *Protocol) setOptionCommand(fields []string) error {
	if len(fields) < 4 {
		return errors.New("invalid setoption arguments")
	}
	var name, value = fields[1], fields[3]
	for _, option := range uci.options {
		if strings.EqualFold(option.UciName(), name) {
			return option.Set(value)
		}
	}
	return errors.New("unhandled option")
}

func (uci *Protocol) isReadyCommand(fields []string) error {
	uci.engine.Prepare()
	fmt.Println("readyok")
	return nil
}

func (uci *Protocol) positionCommand(fields []string) error {
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
		return errors.New("unknown position command")
	}
	var p, err = common.NewPositionFromFEN(fen)
	if err != nil {
		return err
	}
	var positions = []common.Position{p}
	if movesIndex >= 0 && movesIndex+1 < len(args) {
		for _, smove := range args[movesIndex+1:] {
			var newPos, ok = positions[len(positions)-1].MakeMoveLAN(smove)
			if !ok {
				return errors.New("parse move failed")
			}
			positions = append(positions, newPos)
		}
	}
	uci.positions = positions
	return nil
}

func (uci *Protocol) goCommand(fields []string) error {
	var limits = parseLimits(fields)
	var ctx, cancel = context.WithCancel(context.TODO())
	uci.cancel = cancel
	atomic.StoreInt32(&uci.thinking, 1)
	go func() {
		var searchResult = uci.engine.Search(ctx, common.SearchParams{
			Positions: uci.positions,
			Limits:    limits,
			Progress: func(si common.SearchInfo) {
				if si.Time >= 500 || si.Depth >= 5 {
					fmt.Println(searchInfoToUci(si))
				}
			},
		})
		fmt.Println(searchInfoToUci(searchResult))
		atomic.StoreInt32(&uci.thinking, 0)
		if len(searchResult.MainLine) != 0 {
			fmt.Printf("bestmove %v\n", searchResult.MainLine[0])
		}
	}()
	return nil
}

func (uci *Protocol) uciNewGameCommand(fields []string) error {
	uci.engine.Clear()
	return nil
}

func (uci *Protocol) ponderhitCommand(fields []string) error {
	return errors.New("not implemented")
}

func searchInfoToUci(si common.SearchInfo) string {
	var sb = &strings.Builder{}
	fmt.Fprintf(sb, "info depth %v", si.Depth)
	if si.Score.Mate != 0 {
		fmt.Fprintf(sb, " score mate %v", si.Score.Mate)
	} else {
		fmt.Fprintf(sb, " score cp %v", si.Score.Centipawns)
	}
	var timeMs = si.Time.Milliseconds()
	var nps = si.Nodes * 1000 / (timeMs + 1)
	fmt.Fprintf(sb, " nodes %v time %v nps %v", si.Nodes, timeMs, nps)
	if len(si.MainLine) != 0 {
		fmt.Fprintf(sb, " pv")
		for _, move := range si.MainLine {
			sb.WriteString(" ")
			sb.WriteString(move.String())
		}
	}
	return sb.String()
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

func findIndexString(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}
