package main

import (
	"context"
	"log"
	"math/rand"

	"golang.org/x/sync/errgroup"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
)

const maxHeight = 128

type searchStack struct {
	rand  *rand.Rand
	stack [maxHeight]struct {
		positon common.Position
		buffer  [common.MaxMoves]common.OrderedMove
	}
}

func generateOpeningsRandomPipeline(
	ctx context.Context,
	outputFenFilePath string,
	ply int,
) error {
	log.Println("generateOpenings started")
	defer log.Println("generateOpenings finished")

	g, ctx := errgroup.WithContext(ctx)

	var positions = make(chan common.Position, 128)

	g.Go(func() error {
		defer close(positions)
		const height = 0
		var ss = &searchStack{}
		ss.rand = rand.New(rand.NewSource(int64(config.seed)))
		ss.stack[height].positon = startPosition
		for {
			var err = search(ctx, ss, ply, height, positions)
			if err != nil {
				return err
			}
		}
	})

	g.Go(func() error {
		return saveFens(ctx, outputFenFilePath, positions)
	})

	return g.Wait()
}

func search(ctx context.Context, searchStack *searchStack, depth, height int, positions chan<- common.Position) error {
	var position = &searchStack.stack[height].positon
	if depth <= 0 {
		var eval = evaluateMaterial(position)
		const EvalBound = 700
		if -EvalBound < eval && eval < EvalBound {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case positions <- *position:
			}
		}
		return nil
	}
	var ml = position.GenerateMoves(searchStack.stack[height].buffer[:])
	if len(ml) == 0 {
		return nil
	}
	var child = &searchStack.stack[height+1].positon
	for i := 0; i < 3; i++ {
		var move = ml[searchStack.rand.Intn(len(ml))].Move
		if !position.MakeMove(move, child) {
			continue
		}
		var err = search(ctx, searchStack, depth-1, height+1, positions)
		if err != nil {
			return err
		}
	}
	return nil
}

var startPosition, _ = common.NewPositionFromFEN(common.InitialPositionFen)

func evaluateMaterial(p *common.Position) int {
	var eval = 100*(common.PopCount(p.Pawns&p.White)-common.PopCount(p.Pawns&p.Black)) +
		400*(common.PopCount(p.Knights&p.White)-common.PopCount(p.Knights&p.Black)) +
		400*(common.PopCount(p.Bishops&p.White)-common.PopCount(p.Bishops&p.Black)) +
		600*(common.PopCount(p.Rooks&p.White)-common.PopCount(p.Rooks&p.Black)) +
		1200*(common.PopCount(p.Queens&p.White)-common.PopCount(p.Queens&p.Black))
	if !p.WhiteMove {
		eval = -eval
	}
	return eval
}
