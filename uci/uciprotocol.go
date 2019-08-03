package uci

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/common"
)

type UciEngine interface {
	GetInfo() (name, version, author string)
	GetOptions() []common.UciOption
	Prepare()
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

type UciProtocol struct {
	engine    UciEngine
	positions []common.Position
	done      chan struct{}
	cancel    context.CancelFunc
	fields    []string
}

func Run(eng UciEngine) {
	var protocol = NewUciProtocol(eng)
	protocol.Run()
}

func NewUciProtocol(uciEngine UciEngine) *UciProtocol {
	var initPosition, _ = common.NewPositionFromFEN(common.InitialPositionFen)
	var uci = &UciProtocol{
		engine:    uciEngine,
		positions: []common.Position{initPosition},
		done:      make(chan struct{}),
	}
	close(uci.done)
	return uci
}

func (uci *UciProtocol) Run() {
	var name, version, _ = uci.engine.GetInfo()
	fmt.Printf("%v %v build with %v\n", name, version, runtime.Version())

	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" {
			break
		}
		var err = uci.handle(commandLine)
		if err != nil {
			debugUci(err.Error())
		}
	}
}

func (uci *UciProtocol) handle(msg string) error {
	var fields = strings.Fields(msg)
	if len(fields) == 0 {
		return nil
	}
	var commandName = fields[0]
	uci.fields = fields[1:]

	if commandName == "stop" {
		return uci.stopCommand()
	}

	select {
	case <-uci.done:
	default:
		return errors.New("search still run")
	}

	var h func() error
	switch commandName {
	// UCI commands
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
	case "stop":
		h = uci.stopCommand

	// My commands
	case "move":
		h = uci.moveCommand
	}
	if h == nil {
		return errors.New("command not found")
	}
	return h()
}

func debugUci(s string) {
	fmt.Println("info string " + s)
}

func PrintSearchInfo(si common.SearchInfo) {
	var scoreToUci string
	if si.Score.Mate != 0 {
		scoreToUci = fmt.Sprintf("mate %v", si.Score.Mate)
	} else {
		scoreToUci = fmt.Sprintf("cp %v", si.Score.Centipawns)
	}
	var nps = si.Nodes * 1000 / (si.Time + 1)
	var sb strings.Builder
	for i, move := range si.MainLine {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(move.String())
	}
	fmt.Printf("info depth %v score %v nodes %v time %v nps %v pv %v\n",
		si.Depth, scoreToUci, si.Nodes, si.Time, nps, sb.String())
}

func (uci *UciProtocol) uciCommand() error {
	var name, version, author = uci.engine.GetInfo()
	fmt.Printf("id name %s %s\n", name, version)
	fmt.Printf("id author %s\n", author)
	for _, option := range uci.engine.GetOptions() {
		switch option := option.(type) {
		case *common.BoolUciOption:
			fmt.Printf("option name %v type %v default %v\n",
				option.GetName(), "check", option.Value)
		case *common.IntUciOption:
			fmt.Printf("option name %v type %v default %v min %v max %v\n",
				option.GetName(), "spin", option.Value, option.Min, option.Max)
		}
	}
	fmt.Println("uciok")
	return nil
}

func (uci *UciProtocol) setOptionCommand() error {
	if len(uci.fields) < 4 {
		return errors.New("invalid setoption arguments")
	}
	var name, value = uci.fields[1], uci.fields[3]
	for _, option := range uci.engine.GetOptions() {
		if strings.EqualFold(option.GetName(), name) {
			switch option := option.(type) {
			case *common.BoolUciOption:
				v, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				option.Value = v
			case *common.IntUciOption:
				v, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				if v < option.Min || v > option.Max {
					return errors.New("argument out of range")
				}
				option.Value = v
			}
			return nil
		}
	}
	return errors.New("unhandled option")
}

func (uci *UciProtocol) isReadyCommand() error {
	uci.engine.Prepare()
	fmt.Println("readyok")
	return nil
}

func (uci *UciProtocol) positionCommand() error {
	var args = uci.fields
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

func findIndexString(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (uci *UciProtocol) goCommand() error {
	var limits = parseLimits(uci.fields)
	var ctx, cancel = context.WithCancel(context.Background())
	var searchParams = common.SearchParams{
		Positions: uci.positions,
		Limits:    limits,
		Progress: func(si common.SearchInfo) {
			if si.Time >= 500 || si.Depth >= 5 {
				PrintSearchInfo(si)
			}
		},
	}
	uci.done = make(chan struct{})
	uci.cancel = cancel
	go func() {
		var searchResult = uci.engine.Search(ctx, searchParams)
		PrintSearchInfo(searchResult)
		/*Probably even better:
		uci.gate.Lock()
		print bestmove
		uci.idle=true
		uci.gate.Unlock()
		*/
		close(uci.done)
		if len(searchResult.MainLine) == 0 {
			return
		}
		fmt.Printf("bestmove %v\n", searchResult.MainLine[0])
	}()
	return nil
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

func (uci *UciProtocol) uciNewGameCommand() error {
	uci.engine.Clear()
	return nil
}

func (uci *UciProtocol) ponderhitCommand() error {
	return errors.New("not implemented")
}

func (uci *UciProtocol) stopCommand() error {
	if uci.cancel != nil {
		uci.cancel()
	}
	return nil
}

func (uci *UciProtocol) moveCommand() error {
	if len(uci.fields) == 0 {
		return errors.New("invalid move arguments")
	}
	var newPos, ok = uci.positions[len(uci.positions)-1].MakeMoveLAN(uci.fields[0])
	if !ok {
		return errors.New("parse move failed")
	}
	uci.positions = append(uci.positions, newPos)

	var searchParams = common.SearchParams{
		Positions: uci.positions,
		Limits:    common.LimitsType{MoveTime: 3000},
		Progress: func(si common.SearchInfo) {
			if si.Time >= 500 || si.Depth >= 5 {
				PrintSearchInfo(si)
			}
		},
	}
	var searchResult = uci.engine.Search(context.Background(), searchParams)
	PrintSearchInfo(searchResult)
	var child common.Position
	newPos.MakeMove(searchResult.MainLine[0], &child)
	uci.positions = append(uci.positions, child)
	PrintPosition(&child)
	fmt.Println(&child)
	return nil
}
