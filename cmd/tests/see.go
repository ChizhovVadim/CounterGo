package main

import (
	"fmt"

	"github.com/ChizhovVadim/CounterGo/pkg/common"
	"github.com/ChizhovVadim/CounterGo/pkg/engine"
)

func runTestSee(datasetPath string) error {

	var errorCount int

	return walkDataset(datasetPath, func(pos *common.Position, gameResult float64) error {
		var ml = pos.GenerateLegalMoves()
		for _, move := range ml {
			var see = computeSeeSlow(pos, move)

			if !(engine.SeeGE(pos, move, see) &&
				!engine.SeeGE(pos, move, see+1)) {

				fmt.Println("SEE failed",
					"Fen:", pos.String(),
					"Move:", move.String(),
					"SEE:", see)

				errorCount++
				if errorCount >= 15 {
					return fmt.Errorf("errors limit %v", errorCount)
				}
			}
		}

		return nil
	})
}

var pieceValuesSEE = [common.PIECE_NB]int{
	common.Pawn: 1, common.Knight: 4, common.Bishop: 4, common.Rook: 6, common.Queen: 12}

func computeSeeSlow(pos *common.Position, move common.Move) int {
	var from = move.From()
	var to = move.To()
	var movingPiece = move.MovingPiece()
	var capturedPiece = move.CapturedPiece()
	var promotionPiece = move.Promotion()

	var balance = pieceValuesSEE[capturedPiece]

	var nextVictim = movingPiece
	if promotionPiece != common.Empty {
		balance += pieceValuesSEE[promotionPiece] - pieceValuesSEE[common.Pawn]
		nextVictim = promotionPiece
	}

	var occ = pos.AllPieces()&^common.SquareMask[from] | common.SquareMask[to]
	if movingPiece == common.Pawn && to == pos.EpSquare {
		var capSq int
		if pos.WhiteMove {
			capSq = to - 8
		} else {
			capSq = to + 8
		}
		occ &^= common.SquareMask[capSq]
	}

	return balance - computeSeeInternal(pos, !pos.WhiteMove, to, occ, nextVictim)
}

func computeSeeInternal(pos *common.Position, side bool, sq int, occupied uint64, pieceType int) int {
	var myAttackers = computeAttackers(pos, sq, occupied) & occupied & pos.PiecesByColor(side)
	if myAttackers == 0 {
		return 0
	}
	if pieceType == common.King {
		return 200
	}
	var attackerType, attackerFrom = getLeastValuableAttacker(pos, myAttackers)
	occupied &^= common.SquareMask[attackerFrom]
	return common.Max(0, pieceValuesSEE[pieceType]-computeSeeInternal(pos, !side, sq, occupied, attackerType))
}

func computeAttackers(pos *common.Position, sq int, occ uint64) uint64 {
	return (common.PawnAttacks(sq, true) & pos.Pawns & pos.Black) |
		(common.PawnAttacks(sq, false) & pos.Pawns & pos.White) |
		(common.KnightAttacks[sq] & pos.Knights) |
		(common.KingAttacks[sq] & pos.Kings) |
		(common.BishopAttacks(sq, occ) & (pos.Bishops | pos.Queens)) |
		(common.RookAttacks(sq, occ) & (pos.Rooks | pos.Queens))
}

func getLeastValuableAttacker(p *common.Position, attackers uint64) (attacker, from int) {
	if p.Pawns&attackers != 0 {
		return common.Pawn, common.FirstOne(p.Pawns & attackers)
	}
	if p.Knights&attackers != 0 {
		return common.Knight, common.FirstOne(p.Knights & attackers)
	}
	if p.Bishops&attackers != 0 {
		return common.Bishop, common.FirstOne(p.Bishops & attackers)
	}
	if p.Rooks&attackers != 0 {
		return common.Rook, common.FirstOne(p.Rooks & attackers)
	}
	if p.Queens&attackers != 0 {
		return common.Queen, common.FirstOne(p.Queens & attackers)
	}
	if p.Kings&attackers != 0 {
		return common.King, common.FirstOne(p.Kings & attackers)
	}
	return common.Empty, common.SquareNone
}
