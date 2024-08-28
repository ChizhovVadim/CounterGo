package tactic

import "github.com/ChizhovVadim/CounterGo/pkg/common"

type EpdItem struct {
	content   string
	position  common.Position
	bestMoves []common.Move
}
