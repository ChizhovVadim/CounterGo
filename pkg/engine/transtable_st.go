package engine

import (
	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const clusterSize = 3
const transEntrySimpleSize = 16

// 16 bytes
type transEntrySimple struct {
	key   uint32
	move  common.Move
	date  uint16
	score int16
	depth int8
	bound uint8
}

type transTableClusterSignleThread struct {
	megabytes int
	entries   []transEntrySimple
	date      uint16
	divider   uint64
}

func newTransTableClusterSignleThread(megabytes int) *transTableClusterSignleThread {
	var size = 1024 * 1024 * megabytes / transEntrySimpleSize
	return &transTableClusterSignleThread{
		megabytes: megabytes,
		entries:   make([]transEntrySimple, size),
		divider:   uint64(size - clusterSize + 1),
	}
}

func (tt *transTableClusterSignleThread) Size() int {
	return tt.megabytes
}

func (tt *transTableClusterSignleThread) IncDate() {
	tt.date += 1
}

func (tt *transTableClusterSignleThread) Clear() {
	tt.date = 0
	for i := range tt.entries {
		tt.entries[i] = transEntrySimple{}
	}
}

func (tt *transTableClusterSignleThread) Read(key uint64) (depth, score, bound int, move common.Move, ok bool) {
	var index = int(key % tt.divider)
	for i := 0; i < clusterSize; i++ {
		var entry = &tt.entries[index+i]
		if entry.key == uint32(key>>32) {
			entry.date = tt.date
			score = int(entry.score)
			move = entry.move
			depth = int(entry.depth)
			bound = int(entry.bound)
			ok = true
			return
		}
	}
	return
}

func entryCost(depth int8, cutDate bool) int {
	var result = int(depth)
	if cutDate {
		result += 256
	}
	return result
}

func (tt *transTableClusterSignleThread) Update(key uint64, depth, score, bound int, move common.Move) {
	var index = int(key % tt.divider)
	var entry *transEntrySimple
	for i := 0; i < clusterSize; i++ {
		var tempEntry = &tt.entries[index+i]
		if tempEntry.key == uint32(key>>32) {
			entry = tempEntry
			break
		}
		if i == 0 ||
			entryCost(tempEntry.depth, tempEntry.date == tt.date) <
				entryCost(entry.depth, entry.date == tt.date) {
			entry = tempEntry
		}
	}
	var replace bool
	if entry.key == uint32(key>>32) {
		replace = depth >= int(entry.depth)-3 || bound == boundExact
	} else {
		replace = true
	}
	if replace {
		entry.key = uint32(key >> 32)
		entry.score = int16(score)
		entry.depth = int8(depth)
		entry.bound = uint8(bound)
		entry.move = move
		entry.date = tt.date
	}
}

//-----------------------------

const transEntrySecondMoveSize = 32

// 32 bytes
type transEntrySecondMove struct {
	key   uint32
	key2  uint32
	move  common.Move
	move2 common.Move
	date  uint16
	score int16
	depth int8
	bound uint8
}

type transTableSecondMoveSignleThread struct {
	megabytes int
	entries   []transEntrySecondMove
	date      uint16
	divider   uint64
}

func newtransTableSecondMoveSignleThread(megabytes int) *transTableSecondMoveSignleThread {
	var size = 1024 * 1024 * megabytes / transEntrySecondMoveSize
	return &transTableSecondMoveSignleThread{
		megabytes: megabytes,
		entries:   make([]transEntrySecondMove, size),
		divider:   uint64(size),
	}
}

func (tt *transTableSecondMoveSignleThread) IsThreadSafe() bool {
	return false
}

func (tt *transTableSecondMoveSignleThread) Size() int {
	return tt.megabytes
}

func (tt *transTableSecondMoveSignleThread) IncDate() {
	tt.date += 1
}

func (tt *transTableSecondMoveSignleThread) Clear() {
	tt.date = 0
	for i := range tt.entries {
		tt.entries[i] = transEntrySecondMove{}
	}
}

func (tt *transTableSecondMoveSignleThread) Read(key uint64) (depth, score, bound int, move common.Move, ok bool) {
	var index = int(key % tt.divider)
	var entry = &tt.entries[index]
	if entry.key == uint32(key>>32) {
		entry.date = tt.date
		score = int(entry.score)
		move = entry.move
		depth = int(entry.depth)
		bound = int(entry.bound)
		ok = true
	}
	if move == common.MoveEmpty && entry.key2 == uint32(key>>32) {
		move = entry.move2
	}
	return
}

func (tt *transTableSecondMoveSignleThread) Update(key uint64, depth, score, bound int, move common.Move) {
	var index = int(key % tt.divider)
	var entry = &tt.entries[index]
	if move != common.MoveEmpty && bound&boundLower != 0 {
		entry.key2 = uint32(key >> 32)
		entry.move2 = move
	}
	var replace bool
	if entry.key == uint32(key>>32) {
		replace = depth >= int(entry.depth)-3 || bound == boundExact
	} else {
		replace = entry.date != tt.date ||
			depth >= int(entry.depth)
	}
	if replace {
		entry.key = uint32(key >> 32)
		entry.score = int16(score)
		entry.depth = int8(depth)
		entry.bound = uint8(bound)
		entry.move = move
		entry.date = tt.date
	}
}
