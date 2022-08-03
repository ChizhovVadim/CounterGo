package quiet

import (
	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

const maxHeight = 128

type Evaluator interface {
	Evaluate(p *common.Position) int
}

type QuietService struct {
	evaluator   Evaluator
	quietMargin int
	stack       [maxHeight]struct {
		positon common.Position
		buffer  [common.MaxMoves]common.OrderedMove
	}
}

func NewQuietService(evaluator Evaluator, quietMargin int) *QuietService {
	return &QuietService{
		evaluator:   evaluator,
		quietMargin: quietMargin,
	}
}

//TODO test
func (qs *QuietService) IsQuiet(p *common.Position) bool {
	const height = 0
	qs.stack[height].positon = *p
	if isDraw(p) {
		return true
	}
	var eval = qs.evaluator.Evaluate(p)
	var alpha = eval + qs.quietMargin
	return qs.qs(alpha, alpha+1, height) <= alpha
}

func (qs *QuietService) qs(alpha, beta, height int) int {
	var pos = &qs.stack[height].positon
	if isDraw(pos) {
		return 0
	}
	var staticEval = qs.evaluator.Evaluate(pos)
	if height >= maxHeight {
		return staticEval
	}
	if staticEval > alpha {
		alpha = staticEval
		if alpha >= beta {
			return alpha
		}
	}
	var ml = pos.GenerateCaptures(qs.stack[height].buffer[:])
	evalMoves(ml)
	var child = &qs.stack[height+1].positon
	for i := range ml {
		var move = nextMove(ml, i)
		if !engine.SeeGE(pos, move, 0) {
			continue
		}
		if !pos.MakeMove(move, child) {
			continue
		}
		var score = -qs.qs(-beta, -alpha, height+1)
		if score > alpha {
			alpha = score
			if alpha >= beta {
				return alpha
			}
		}
	}
	return alpha
}

func isDraw(p *common.Position) bool {
	if (p.Pawns|p.Rooks|p.Queens) == 0 &&
		!common.MoreThanOne(p.Knights|p.Bishops) {
		return true
	}
	return false
}

var sortPieceValues = [common.PIECE_NB]int{
	common.Pawn: 1, common.Knight: 2, common.Bishop: 3, common.Rook: 4, common.Queen: 5, common.King: 6}

func mvvlva(move common.Move) int {
	return 8*(sortPieceValues[move.CapturedPiece()]+
		sortPieceValues[move.Promotion()]) -
		sortPieceValues[move.MovingPiece()]
}

func evalMoves(ml []common.OrderedMove) {
	for i := range ml {
		var move = ml[i].Move
		var score = mvvlva(move)
		ml[i].Key = int32(score)
	}
}

func nextMove(ml []common.OrderedMove, index int) common.Move {
	var bestIndex = index
	for i := bestIndex + 1; i < len(ml); i++ {
		if ml[i].Key > ml[bestIndex].Key {
			bestIndex = i
		}
	}
	if bestIndex != index {
		ml[index], ml[bestIndex] = ml[bestIndex], ml[index]
	}
	return ml[index].Move
}
