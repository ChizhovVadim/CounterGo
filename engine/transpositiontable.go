package engine

import (
	"fmt"
	"math/rand"
	"sync"
)

type TranspositionTable struct {
	items      []TTEntry
	readNumber int
	readHit    int
	age        uint8
	gates      []sync.Mutex
}

type TTEntry struct {
	Key   uint64
	Move  Move
	Score int16
	Depth int8
	Data  uint8 //bits 0-1: entry type, bits 2-7: age
}

const (
	Lower = 1 << iota
	Upper
)

func NewTranspositionTable(megabytes int) *TranspositionTable {
	return &TranspositionTable{
		items: make([]TTEntry, 1024*1024*megabytes/16),
		gates: make([]sync.Mutex, 128),
	}
}

func (tt *TranspositionTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
	var key = p.Key
	var index = int(key % uint64(len(tt.items)))
	var gate = &tt.gates[index%len(tt.gates)]
	var item = &tt.items[index]
	gate.Lock()
	tt.readNumber++
	if item.Key == key {
		tt.readHit++
		item.Data = (item.Data & 3) | (tt.age << 2)
		score = int(item.Score)
		move = item.Move
		depth = int(item.Depth)
		entryType = int(item.Data & 3)
		ok = true
	}
	gate.Unlock()
	return
}

func (tt *TranspositionTable) Update(p *Position, depth, score, entryType int, move Move) {
	var key = p.Key
	var index = int(key % uint64(len(tt.items)))
	var gate = &tt.gates[index%len(tt.gates)]
	var item = &tt.items[index]
	gate.Lock()
	if depth >= int(item.Depth) || tt.age != (item.Data>>2) {
		*item = TTEntry{
			Key:   key,
			Move:  move,
			Score: int16(score),
			Depth: int8(depth),
			Data:  uint8(entryType) | (tt.age << 2),
		}
	}
	gate.Unlock()
}

func (tt *TranspositionTable) ClearStatistics() {
	tt.readNumber = 0
	tt.readHit = 0
	tt.age = (tt.age + 1) & 63
}

func (tt *TranspositionTable) PrintStatistics() {
	var hit = float64(tt.readHit) / float64(tt.readNumber)
	fmt.Printf("info string Transposition table hit: %v\n", hit)
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
