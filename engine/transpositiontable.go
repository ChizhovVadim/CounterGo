package engine

import (
	"sync/atomic"
)

type transTable struct {
	megabytes  int
	entries    []transEntry
	generation uint8
	mask       uint32
}

type transEntry struct {
	gate      int32
	key32     uint32
	move      Move
	score     int16
	depth     int8
	bound_gen uint8
}

const (
	Lower = 1 << iota
	Upper
)

const ClusterSize = 1 << 2

func NewTransTable(megabytes int) *transTable {
	var size = 1024 * 1024 * megabytes / 16
	size = roundPowerOfTwo(size)
	return &transTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size-1) &^ (ClusterSize - 1),
	}
}

func roundPowerOfTwo(size int) int {
	var x = 1
	for (x << 1) <= size {
		x <<= 1
	}
	return x
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
	var index = int(uint32(p.Key) & tt.mask)
	for i := 0; i < ClusterSize; i++ {
		var entry = &tt.entries[index+i]
		if entry.key32 == uint32(p.Key>>32) &&
			atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
			if entry.key32 == uint32(p.Key>>32) {
				entry.bound_gen = (entry.bound_gen & 3) + (tt.generation << 2)
				score = int(entry.score)
				move = entry.move
				depth = int(entry.depth)
				bound = int(entry.bound_gen & 3)
				ok = true
			}
			atomic.StoreInt32(&entry.gate, 0)
			break
		}
	}
	return
}

//position fen 8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - - 0 1
func (tt *transTable) Update(p *Position, depth, score, bound int, move Move) {
	var index = int(uint32(p.Key) & tt.mask)
	var bestEntry *transEntry
	var bestScore = -32767
	for i := 0; i < ClusterSize; i++ {
		var entry = &tt.entries[index+i]
		if entry.key32 == uint32(p.Key>>32) {
			bestEntry = entry
			break
		}
		var score = Score(entry.depth, entry.bound_gen>>2, tt.generation)
		if score > bestScore {
			bestScore = score
			bestEntry = entry
		}
	}
	if atomic.CompareAndSwapInt32(&bestEntry.gate, 0, 1) {
		bestEntry.key32 = uint32(p.Key >> 32)
		bestEntry.move = move
		bestEntry.score = int16(score)
		bestEntry.depth = int8(depth)
		bestEntry.bound_gen = uint8(bound) + (tt.generation << 2)
		atomic.StoreInt32(&bestEntry.gate, 0)
	}
}

func Score(depth int8, gen, curGen uint8) int {
	var score = -int(depth)
	if gen != curGen {
		score += 100
	}
	return score
}
