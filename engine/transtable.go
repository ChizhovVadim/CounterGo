package engine

import (
	"sync/atomic"

	. "github.com/ChizhovVadim/CounterGo/common"
)

// good test: position fen 8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - - 0 1
type TransTable interface {
	Megabytes() int
	PrepareNewSearch()
	Clear()
	Read(p *Position) (depth, score, entryType int, move Move, ok bool)
	Update(p *Position, depth, score, entryType int, move Move)
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
	boundLower = 1 << iota
	boundUpper
)

func roundPowerOfTwo(size int) int {
	var x = 1
	for (x << 1) <= size {
		x <<= 1
	}
	return x
}

//----------------------------------------------------------------------------

type deepReplaceTransTable struct {
	megabytes  int
	entries    []transEntry
	generation uint8
	mask       uint32
}

func NewDeepReplaceTransTable(megabytes int) *deepReplaceTransTable {
	var size = roundPowerOfTwo(1024 * 1024 * megabytes / 16)
	return &deepReplaceTransTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size - 1),
	}
}

func (tt *deepReplaceTransTable) Megabytes() int {
	return tt.megabytes
}

func (tt *deepReplaceTransTable) PrepareNewSearch() {
	tt.generation = (tt.generation + 1) & 63
}

func (tt *deepReplaceTransTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = transEntry{}
	}
}

func (tt *deepReplaceTransTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
	var entry = &tt.entries[uint32(p.Key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		if entry.key32 == uint32(p.Key>>32) {
			entry.bound_gen = (entry.bound_gen & 3) + (tt.generation << 2)
			score = int(entry.score)
			move = entry.move
			depth = int(entry.depth)
			entryType = int(entry.bound_gen & 3)
			ok = true
		}
		atomic.StoreInt32(&entry.gate, 0)
	}
	return
}

func (tt *deepReplaceTransTable) Update(p *Position, depth, score, entryType int, move Move) {
	var entry = &tt.entries[uint32(p.Key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		if entry.bound_gen>>2 != tt.generation ||
			depth >= int(entry.depth) ||
			entry.key32 == uint32(p.Key>>32) {
			entry.key32 = uint32(p.Key >> 32)
			entry.move = move
			entry.score = int16(score)
			entry.depth = int8(depth)
			entry.bound_gen = uint8(entryType) + (tt.generation << 2)
		}
		atomic.StoreInt32(&entry.gate, 0)
	}
}

//----------------------------------------------------------------------------

type alwaysReplaceTransTable struct {
	megabytes int
	entries   []transEntry
	mask      uint32
}

func NewAlwaysReplaceTransTable(megabytes int) *alwaysReplaceTransTable {
	var size = roundPowerOfTwo(1024 * 1024 * megabytes / 16)
	return &alwaysReplaceTransTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size - 1),
	}
}

func (tt *alwaysReplaceTransTable) Megabytes() int {
	return tt.megabytes
}

func (tt *alwaysReplaceTransTable) PrepareNewSearch() {
}

func (tt *alwaysReplaceTransTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = transEntry{}
	}
}

func (tt *alwaysReplaceTransTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
	var entry = &tt.entries[uint32(p.Key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		if entry.key32 == uint32(p.Key>>32) {
			score = int(entry.score)
			move = entry.move
			depth = int(entry.depth)
			entryType = int(entry.bound_gen & 3)
			ok = true
		}
		atomic.StoreInt32(&entry.gate, 0)
	}
	return
}

func (tt *alwaysReplaceTransTable) Update(p *Position, depth, score, entryType int, move Move) {
	var entry = &tt.entries[uint32(p.Key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		entry.key32 = uint32(p.Key >> 32)
		entry.move = move
		entry.score = int16(score)
		entry.depth = int8(depth)
		entry.bound_gen = uint8(entryType)
		atomic.StoreInt32(&entry.gate, 0)
	}
}

//----------------------------------------------------------------------------

const clusterSize = 4

type tierTransTable struct {
	megabytes  int
	entries    []transEntry
	generation uint8
	mask       uint32
}

func NewTierTransTable(megabytes int) *tierTransTable {
	var size = roundPowerOfTwo(1024 * 1024 * megabytes / 16)
	return &tierTransTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size - 4),
	}
}

func (tt *tierTransTable) Megabytes() int {
	return tt.megabytes
}

func (tt *tierTransTable) PrepareNewSearch() {
	tt.generation = (tt.generation + 1) & 63
}

func (tt *tierTransTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = transEntry{}
	}
}

func (tt *tierTransTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
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
				entryType = int(entry.bound_gen & 3)
				ok = true
				break
			}
		}
		atomic.StoreInt32(gate, 0)
	}
	return
}

func (tt *tierTransTable) Update(p *Position, depth, score, entryType int, move Move) {
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
			var score = transEntryScore(entry.depth, entry.bound_gen>>2, tt.generation)
			if score > bestScore {
				bestScore = score
				bestEntry = entry
			}
		}
		bestEntry.key32 = uint32(p.Key >> 32)
		bestEntry.move = move
		bestEntry.score = int16(score)
		bestEntry.depth = int8(depth)
		bestEntry.bound_gen = uint8(entryType) + (tt.generation << 2)
		atomic.StoreInt32(gate, 0)
	}
}

func transEntryScore(depth int8, gen, curGen uint8) int {
	var score = -int(depth)
	if gen != curGen {
		score += 100
	}
	return score
}
