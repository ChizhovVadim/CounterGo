package engine

import (
	"context"
	"sync"
)

func (e *Engine) lazySmp(
	ctx context.Context,
	sharedContext *SharedContext,
) mainLine {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	sharedContext.done = ctx.Done()

	var wg = &sync.WaitGroup{}
	for threadIdx := 1; threadIdx < e.Config.Threads; threadIdx += 1 {
		wg.Add(1)
		go func(t *thread) {
			defer wg.Done()
			t.iterativeDeepening(sharedContext)
		}(&e.threads[threadIdx])
	}

	// Если есть производительные ядра и экономичные, то нет гарантии, что главный поток будет исполняться на производительном ядре.
	var res = e.threads[0].iterativeDeepening(sharedContext)
	cancel()
	wg.Wait()
	return res
}
