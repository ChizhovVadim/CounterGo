package engine

import (
	"math/rand"
	"sync/atomic"
)

type TranspositionTable struct {
	items []ttEntry
	gates []int32
	age   uint8
}

type ttEntry struct {
	key   uint64
	move  Move
	score int16
	depth int8
	data  uint8 //bits 0-1: entry type, bits 2-7: age
}

const (
	Lower = 1 << iota
	Upper
)

func NewTranspositionTable(megabytes int) *TranspositionTable {
	return &TranspositionTable{
		items: make([]ttEntry, 1024*1024*megabytes/16),
		gates: make([]int32, 1024),
	}
}

func (tt *TranspositionTable) PrepareNewSearch() {
	tt.age = (tt.age + 1) & 63
}

func (tt *TranspositionTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
	var key = p.Key
	var index = int(key & uint64(len(tt.items)-1))
	var item = &tt.items[index]
	var gate = &tt.gates[index&(len(tt.gates)-1)]
	if atomic.CompareAndSwapInt32(gate, 0, 1) {
		if item.key == key {
			item.data = (item.data & 3) | (tt.age << 2)
			score = int(item.score)
			move = item.move
			depth = int(item.depth)
			entryType = int(item.data & 3)
			ok = true
		}
		atomic.StoreInt32(gate, 0)
	}
	return
}

func (tt *TranspositionTable) Update(p *Position, depth, score, entryType int, move Move) {
	var key = p.Key
	var index = int(key & uint64(len(tt.items)-1))
	var item = &tt.items[index]
	var gate = &tt.gates[index&(len(tt.gates)-1)]
	if atomic.CompareAndSwapInt32(gate, 0, 1) {
		if depth >= int(item.depth) || tt.age != (item.data>>2) {
			*item = ttEntry{
				key:   key,
				move:  move,
				score: int16(score),
				depth: int8(depth),
				data:  uint8(entryType) | (tt.age << 2),
			}
		}
		atomic.StoreInt32(gate, 0)
	}
}

var (
	sideKey        uint64
	enpassantKey   [8]uint64
	castlingKey    [16]uint64
	pieceSquareKey [7 * 2 * 64]uint64
)

func PieceSquareKey(piece int, side bool, square int) uint64 {
	return pieceSquareKey[MakePiece(piece, side)*64+square]
}

func (p *Position) ComputeKey() uint64 {
	var result = uint64(0)
	if p.WhiteMove {
		result ^= sideKey
	}
	result ^= castlingKey[p.CastleRights]
	if p.EpSquare != SquareNone {
		result ^= enpassantKey[File(p.EpSquare)]
	}
	for i := 0; i < 64; i++ {
		var piece = p.WhatPiece(i)
		if piece != Empty {
			var side = (p.White & squareMask[i]) != 0
			result ^= PieceSquareKey(piece, side, i)
		}
	}
	return result
}

func init() {
	var r = rand.New(rand.NewSource(0))
	sideKey = r.Uint64()
	for i := 0; i < len(enpassantKey); i++ {
		enpassantKey[i] = r.Uint64()
	}
	for i := 0; i < len(pieceSquareKey); i++ {
		pieceSquareKey[i] = r.Uint64()
	}

	var castle = make([]uint64, 4)
	for i := 0; i < len(castle); i++ {
		castle[i] = r.Uint64()
	}

	for i := 0; i < len(castlingKey); i++ {
		for j := 0; j < 4; j++ {
			if (i & (1 << uint(j))) != 0 {
				castlingKey[i] ^= castle[j]
			}
		}
	}
}
