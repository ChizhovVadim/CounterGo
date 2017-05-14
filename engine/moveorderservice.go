package engine

type MoveOrderService struct {
	histSuccess, histTry []int
}

func NewMoveOrderService() *MoveOrderService {
	var result = &MoveOrderService{
		histSuccess: make([]int, 7*2*64),
		histTry:     make([]int, 7*2*64),
	}
	return result
}

func (this *MoveOrderService) Clear() {
	for i := 0; i < len(this.histSuccess); i++ {
		this.histSuccess[i] = 0
		this.histTry[i] = 0
	}
}

func HistoryIndex(side bool, move Move) int {
	return MakePiece(move.MovingPiece(), side)*64 + move.To()
}

func (this *MoveOrderService) UpdateHistory(ss *SearchStack, bestMove Move, depth int) {
	ss.KillerMove = bestMove
	var side = ss.Position.WhiteMove
	this.histSuccess[HistoryIndex(side, bestMove)] += depth
	for _, move := range ss.QuietsSearched {
		this.histTry[HistoryIndex(side, move)] += depth
	}
}

func (this *MoveOrderService) NoteMoves(ss *SearchStack, hashMove Move) {
	var killerMove = ss.KillerMove
	var side = ss.Position.WhiteMove
	var buffer = ss.MoveList.Items[:]
	var count = ss.MoveList.Count
	for i := 0; i < count; i++ {
		var move = buffer[i].Move
		var score int
		if move == hashMove {
			score = 30000
		} else {
			var captureScore = pieceValuesSEE[move.CapturedPiece()]
			if move.Promotion() != Empty {
				captureScore += pieceValuesSEE[move.Promotion()] - pieceValuesSEE[Pawn]
			}
			if captureScore != 0 {
				score = 29000 + captureScore*8 - move.MovingPiece()
			} else if move == killerMove {
				score = 28000
			} else {
				var index = HistoryIndex(side, move)
				if this.histTry[index] != 0 {
					score = 100 * this.histSuccess[index] / this.histTry[index]
				} else {
					score = 0
				}
			}
		}
		buffer[i].Score = score
	}
}
