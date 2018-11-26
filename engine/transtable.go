package engine

import (
	"sync/atomic"

	. "github.com/ChizhovVadim/CounterGo/common"
)

type transEntry struct {
	gate      int32
	key32     uint32
	move      Move
	score     int16
	depth     int8
	bound_gen uint8
}

const (
	boundLower = 1 << iota
	boundUpper
)

const boundExact = boundLower | boundUpper

func roundPowerOfTwo(size int) int {
	var x = 1
	for (x << 1) <= size {
		x <<= 1
	}
	return x
}

const clusterSize = 4

type transTable struct {
	megabytes  int
	entries    []transEntry
	generation uint8
	mask       uint32
}

// good test: position fen 8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - - 0 1
// good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
func NewTransTable(megabytes int) *transTable {
	var size = roundPowerOfTwo(1024 * 1024 * megabytes / 16)
	return &transTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size - 4),
	}
}

func (tt *transTable) Megabytes() int {
	return tt.megabytes
}

func (tt *transTable) PrepareNewSearch() {
	tt.generation = (tt.generation + 1) & 63
}

func (tt *transTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = transEntry{}
	}
}

func (tt *transTable) Read(p *Position) (depth, score, bound int, move Move, ok bool) {
	var index = uint32(p.Key) & tt.mask
	var entries = tt.entries[index : index+clusterSize]
	var gate = &entries[0].gate
	if atomic.CompareAndSwapInt32(gate, 0, 1) {
		for i := range entries {
			var entry = &entries[i]
			if entry.key32 == uint32(p.Key>>32) {
				entry.bound_gen = (entry.bound_gen & 3) + (tt.generation << 2)
				score = int(entry.score)
				move = entry.move
				depth = int(entry.depth)
				bound = int(entry.bound_gen & 3)
				ok = true
				break
			}
		}
		atomic.StoreInt32(gate, 0)
	}
	return
}

func (tt *transTable) Update(p *Position, depth, score, bound int, move Move) {
	var index = uint32(p.Key) & tt.mask
	var entries = tt.entries[index : index+clusterSize]
	var gate = &entries[0].gate
	if atomic.CompareAndSwapInt32(gate, 0, 1) {
		var bestEntry *transEntry
		var bestScore = -32767
		for i := range entries {
			var entry = &entries[i]
			if entry.key32 == uint32(p.Key>>32) {
				bestEntry = entry
				break
			}
			var score = transEntryScore(int(entry.depth), (entry.bound_gen>>2) == tt.generation)
			if score > bestScore {
				bestScore = score
				bestEntry = entry
			}
		}
		if bestEntry.key32 == uint32(p.Key>>32) {
			if move != MoveEmpty {
				bestEntry.move = move
			}
			//if bound != 0 {
			if bound == boundExact || depth >= int(bestEntry.depth)-3 {
				bestEntry.score = int16(score)
				bestEntry.depth = int8(depth)
				bestEntry.bound_gen = uint8(bound) + (tt.generation << 2)
			}
		} else {
			bestEntry.key32 = uint32(p.Key >> 32)
			bestEntry.move = move
			bestEntry.score = int16(score)
			bestEntry.depth = int8(depth)
			bestEntry.bound_gen = uint8(bound) + (tt.generation << 2)
		}
		atomic.StoreInt32(gate, 0)
	}
}

func transEntryScore(depth int, newGen bool) int {
	var score = -depth
	if !newGen {
		score += 100
	}
	return score
}
