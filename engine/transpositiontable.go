package engine

import (
	"math/rand"
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

const ClusterSize = 4

func NewTransTable(megabytes int) *transTable {
	var size = 1024 * 1024 * megabytes / 16
	return &transTable{
		megabytes: megabytes,
		entries:   make([]transEntry, size+ClusterSize-1),
		mask:      uint32(size - 1),
	}
}

func (tt *transTable) PrepareNewSearch() {
	tt.generation = (tt.generation + 1) & 63
}

func (tt *transTable) Read(p *Position) (depth, score, entryType int, move Move, ok bool) {
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
				entryType = int(entry.bound_gen & 3)
				ok = true
			}
			atomic.StoreInt32(&entry.gate, 0)
			break
		}
	}
	return
}

//position fen 8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - - 0 1
func (tt *transTable) Update(p *Position, depth, score, entryType int, move Move) {
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
		bestEntry.bound_gen = uint8(entryType) + (tt.generation << 2)
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
