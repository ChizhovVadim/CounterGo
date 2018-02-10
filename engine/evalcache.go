package engine

import (
	"sync/atomic"

	"github.com/ChizhovVadim/CounterGo/common"
)

const (
	evalCacheSize = 1 << 16
	evalCacheMask = evalCacheSize - 1
)

type evalEntry struct {
	gate  int32
	key32 uint32
	eval  int
}

func evalCacheDecorator(evaluate evaluate) evaluate {
	var entries = make([]evalEntry, evalCacheSize)
	return func(p *common.Position) int {
		var entry = &entries[uint32(p.Key)&evalCacheMask]
		var eval int
		if atomic.CompareAndSwapInt32(&entry.gate, 0, 1) {
			if entry.key32 == uint32(p.Key>>32) {
				eval = entry.eval
			} else {
				eval = evaluate(p)
				entry.key32 = uint32(p.Key >> 32)
				entry.eval = eval
			}
			atomic.StoreInt32(&entry.gate, 0)
		} else {
			eval = evaluate(p)
		}
		return eval
	}
}
