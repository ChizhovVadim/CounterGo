package engine

import (
	"sync/atomic"

	"github.com/ChizhovVadim/CounterGo/common"
)

func evalCacheDecorator(evaluate evaluate) evaluate {
	const (
		Size     = 1 << 16
		SizeMask = Size - 1
		EvalMask = uint64(0xFFFF)
		KeyMask  = ^EvalMask
		EvalZero = 32768
	)
	var entries = make([]uint64, Size)
	return func(p *common.Position) int {
		var entry = &entries[uint32(p.Key)&SizeMask]
		var data = atomic.LoadUint64(entry)
		if data&KeyMask == p.Key&KeyMask {
			return int(data&EvalMask) - EvalZero
		}
		var eval = evaluate(p)
		atomic.StoreUint64(entry, (p.Key&KeyMask)|uint64(eval+EvalZero))
		return eval
	}
}
