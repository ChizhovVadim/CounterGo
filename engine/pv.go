package engine

import (
	. "github.com/ChizhovVadim/CounterGo/common"
)

const (
	pvTableSize = 1 << 14
	pvTableMask = pvTableSize - 1
)

type pvTable struct {
	items [pvTableSize]pvItem
}

type pvItem struct {
	key  uint64
	move Move
}

func NewPvTable() *pvTable {
	return &pvTable{}
}

func (pt *pvTable) Clear() {
	for i := range pt.items {
		pt.items[i] = pvItem{0, 0}
	}
}

func (pt *pvTable) Save(p *Position, m Move) {
	pt.items[p.Key&pvTableMask] = pvItem{p.Key, m}
}

func (pt *pvTable) Read(p *Position) []Move {
	var moves []Move
	var seen = make(map[uint64]bool)
	var position, child Position
	position = *p
	for !seen[position.Key] {
		seen[position.Key] = true
		var item = pt.items[position.Key&pvTableMask]
		if item.key != position.Key {
			break
		}
		var move = item.move
		if !position.MakeMove(move, &child) {
			break
		}
		position = child
		moves = append(moves, move)
	}
	return moves
}
