package engine

import (
	"sync/atomic"

	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

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

//16 bytes
type transEntry struct {
	gate     int32
	key32    uint32
	moveDate uint32
	score    int16
	depth    int8
	bound    uint8
}

func (entry *transEntry) Move() Move {
	return Move(entry.moveDate & 0x1fffff)
}

func (entry *transEntry) Date() uint16 {
	return uint16(entry.moveDate >> 21)
}

func (entry *transEntry) SetMoveAndDate(move Move, date uint16) {
	entry.moveDate = uint32(move) + uint32(date)<<21
}

type transTable struct {
	megabytes int
	entries   []transEntry
	date      uint16
	mask      uint32
}

// good test: position fen 8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - - 0 1
// good test: position fen 8/pp6/2p5/P1P5/1P3k2/3K4/8/8 w - - 5 47
func newTransTable(megabytes int) *transTable {
	var size = roundPowerOfTwo(1024 * 1024 * megabytes / 16)
	return &transTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size),
		mask:      uint32(size - 1),
	}
}

func (tt *transTable) Size() int {
	return tt.megabytes
}

func (tt *transTable) IncDate() {
	tt.date = (tt.date + 1) & 0x7ff
}

func (tt *transTable) Clear() {
	tt.date = 0
	for i := range tt.entries {
		tt.entries[i] = transEntry{}
	}
}

func (tt *transTable) Read(key uint64) (depth, score, bound int, move Move, ok bool) {
	var entry = &tt.entries[uint32(key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		if entry.key32 == uint32(key>>32) {
			entry.SetMoveAndDate(entry.Move(), tt.date)
			score = int(entry.score)
			move = entry.Move()
			depth = int(entry.depth)
			bound = int(entry.bound)
			ok = true
		}
		atomic.StoreInt32(&entry.gate, 0)
	}
	return
}

func (tt *transTable) Update(key uint64, depth, score, bound int, move Move) {
	var entry = &tt.entries[uint32(key)&tt.mask]
	if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
		var replace bool
		if entry.key32 == uint32(key>>32) {
			replace = depth >= int(entry.depth)-3 || bound == boundExact
		} else {
			replace = entry.Date() != tt.date ||
				depth >= int(entry.depth)
		}
		if replace {
			entry.key32 = uint32(key >> 32)
			entry.score = int16(score)
			entry.depth = int8(depth)
			entry.bound = uint8(bound)
			entry.SetMoveAndDate(move, tt.date)
		}
		atomic.StoreInt32(&entry.gate, 0)
	}
}
