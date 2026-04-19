package uci

import (
	"context"
	"fmt"
	"strings"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

type IEngine interface {
	Prepare()
	Clear()
	Search(ctx context.Context, searchParams common.SearchParams) common.SearchInfo
}

type UciMessage struct{}
type NewGameMessage struct{}
type SetOptionMessage struct{ Name, Value string }
type IsReadyMessage struct{}

type GoMessage struct {
	Ctx    context.Context
	Game   []common.Position
	Limits common.LimitsType
}

type EngineAgent struct {
	name    string
	author  string
	version string
	eng     IEngine
	options []Option
}

func NewEngineAgent(
	name string,
	author string,
	version string,
	eng IEngine,
	options []Option,
) *EngineAgent {
	return &EngineAgent{
		name:    name,
		author:  author,
		version: version,
		options: options,
		eng:     eng,
	}
}

func (agent *EngineAgent) Run(
	cmds <-chan any,
) error {
	for msg := range cmds {
		switch msg := msg.(type) {
		case UciMessage:
			fmt.Printf("id name %s %s\n", agent.name, agent.version)
			fmt.Printf("id author %s\n", agent.author)
			for _, option := range agent.options {
				fmt.Println(option.UciString())
			}
			fmt.Println("uciok")
		case NewGameMessage:
			agent.eng.Clear()
		case IsReadyMessage:
			agent.eng.Prepare()
			fmt.Println("readyok")
		case SetOptionMessage:
			for _, option := range agent.options {
				if strings.EqualFold(option.UciName(), msg.Name) {
					option.Set(msg.Value)
					break
				}
			}
		case GoMessage:
			var searchResult = agent.eng.Search(msg.Ctx, common.SearchParams{
				Positions: msg.Game,
				Limits:    msg.Limits,
				Progress: func(si common.SearchInfo) {
					fmt.Println(searchInfoToUci(si))
				},
			})
			fmt.Println(searchInfoToUci(searchResult))
			if len(searchResult.MainLine) != 0 {
				fmt.Printf("bestmove %v\n", searchResult.MainLine[0])
			}
		}
	}
	return nil
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
