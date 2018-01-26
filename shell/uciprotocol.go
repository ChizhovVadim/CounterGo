package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ChizhovVadim/CounterGo/engine"
)

type UciEngine interface {
	GetInfo() (name, version, author string)
	GetOptions() []engine.UciOption
	Prepare()
	Search(searchParams engine.SearchParams) engine.SearchInfo
}

type UciProtocol struct {
	messages  chan interface{}
	engine    UciEngine
	positions []*engine.Position
	cancel    context.CancelFunc
	fields    []string
}

func NewUciProtocol(uciEngine UciEngine) *UciProtocol {
	var initPosition = engine.NewPositionFromFEN(engine.InitialPositionFen)
	return &UciProtocol{
		messages:  make(chan interface{}),
		engine:    uciEngine,
		positions: []*engine.Position{initPosition},
	}
}

func (uci *UciProtocol) Run() {
	var name, version, _ = uci.engine.GetInfo()
	fmt.Printf("%v %v\n", name, version)
	go func() {
		var commands = map[string]func(){
			// UCI commands
			"uci":        uci.uciCommand,
			"setoption":  uci.setOptionCommand,
			"isready":    uci.isReadyCommand,
			"position":   uci.positionCommand,
			"go":         uci.goCommand,
			"ucinewgame": uci.uciNewGameCommand,
			"ponderhit":  uci.ponderhitCommand,
			"stop":       uci.stopCommand,

			// My commands
			"epd":   uci.epdCommand,
			"arena": uci.arenaCommand,
		}
		for msg := range uci.messages {
			switch msg := msg.(type) {
			case string:
				var fields = strings.Fields(msg)
				if len(fields) == 0 {
					continue
				}
				var commandName = fields[0]
				var cmd, ok = commands[commandName]
				if ok {
					uci.fields = fields[1:]
					cmd()
				} else {
					debugUci("Command not found.")
				}
			case engine.SearchInfo:
				PrintSearchInfo(msg)
			case engine.Move:
				fmt.Printf("bestmove %v\n", msg)
			}
		}
	}()
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" {
			break
		}
		uci.messages <- commandLine
	}
}

func debugUci(s string) {
	fmt.Println("info string " + s)
}

func PrintSearchInfo(si engine.SearchInfo) {
	var scoreToUci string
	if mate, isMate := engine.ScoreToMate(si.Score); isMate {
		scoreToUci = fmt.Sprintf("mate %v", mate)
	} else {
		scoreToUci = fmt.Sprintf("cp %v", si.Score)
	}
	var nps = si.Nodes * 1000 / (si.Time + 1)
	var sb bytes.Buffer
	for i, move := range si.MainLine {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(move.String())
	}
	fmt.Printf("info score %v depth %v nodes %v time %v nps %v pv %v\n",
		scoreToUci, si.Depth, si.Nodes, si.Time, nps, sb.String())
}

func (uci *UciProtocol) uciCommand() {
	var name, version, author = uci.engine.GetInfo()
	fmt.Printf("id name %s %s\n", name, version)
	fmt.Printf("id author %s\n", author)
	for _, option := range uci.engine.GetOptions() {
		switch option := option.(type) {
		case *engine.BoolUciOption:
			fmt.Printf("option name %v type %v default %v\n",
				option.Name(), "check", option.Value)
		case *engine.IntUciOption:
			fmt.Printf("option name %v type %v default %v min %v max %v\n",
				option.Name(), "spin", option.Value, option.Min, option.Max)
		}
	}
	fmt.Println("uciok")
}

func (uci *UciProtocol) setOptionCommand() {
	if len(uci.fields) < 4 {
		return
	}
	var name, value = uci.fields[1], uci.fields[3]
	for _, option := range uci.engine.GetOptions() {
		if strings.EqualFold(option.Name(), name) {
			switch option := option.(type) {
			case *engine.BoolUciOption:
				if v, err := strconv.ParseBool(value); err == nil {
					option.Value = v
				}
			case *engine.IntUciOption:
				if v, err := strconv.Atoi(value); err == nil &&
					option.Min <= v && v <= option.Max {
					option.Value = v
				}
			}
			return
		}
	}
}

func (uci *UciProtocol) isReadyCommand() {
	uci.engine.Prepare()
	fmt.Println("readyok")
}

func (uci *UciProtocol) positionCommand() {
	var args = uci.fields
	var token = args[0]
	var fen string
	var movesIndex = findIndexString(args, "moves")
	if token == "startpos" {
		fen = engine.InitialPositionFen
	} else if token == "fen" {
		if movesIndex == -1 {
			fen = strings.Join(args[1:], " ")
		} else {
			fen = strings.Join(args[1:movesIndex], " ")
		}
	} else {
		debugUci("Wrong position command")
		return
	}
	var p = engine.NewPositionFromFEN(fen)
	if p == nil {
		debugUci("Wrong fen")
		return
	}
	var positions = []*engine.Position{p}
	if movesIndex >= 0 && movesIndex+1 < len(args) {
		for _, smove := range args[movesIndex+1:] {
			var move = engine.ParseMove(smove)
			var newPos = positions[len(positions)-1].MakeMoveIfLegal(move)
			if newPos == nil {
				debugUci("Wrong move")
				return
			} else {
				positions = append(positions, newPos)
			}
		}
	}
	uci.positions = positions
}

func findIndexString(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (uci *UciProtocol) goCommand() {
	var limits = parseLimits(uci.fields)
	var ctx, cancel = context.WithCancel(context.Background())
	var searchParams = engine.SearchParams{
		Positions: uci.positions,
		Limits:    limits,
		Ctx:       ctx,
		Progress: func(si engine.SearchInfo) {
			if si.Time >= 500 || si.Depth >= 5 {
				uci.messages <- si
			}
		},
	}
	uci.cancel = cancel
	go func() {
		var searchResult = uci.engine.Search(searchParams)
		uci.messages <- searchResult
		if len(searchResult.MainLine) > 0 {
			uci.messages <- searchResult.MainLine[0]
		}
	}()
}

func parseLimits(args []string) (result engine.LimitsType) {
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

func (uci *UciProtocol) uciNewGameCommand() {
}

func (uci *UciProtocol) ponderhitCommand() {
}

func (uci *UciProtocol) stopCommand() {
	if uci.cancel != nil {
		uci.cancel()
	}
}

func (uci *UciProtocol) epdCommand() {
	var filePath = "tests.epd"
	if len(uci.fields) > 0 {
		filePath = uci.fields[0]
	}
	RunEpdTest(filePath, uci.engine)
}

func (uci *UciProtocol) arenaCommand() {
	RunTournament()
}
