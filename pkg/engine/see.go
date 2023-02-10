package engine

import (
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

var pieceValuesSEE = [PIECE_NB]int{Pawn: 1, Knight: 4, Bishop: 4, Rook: 6, Queen: 12, King: 120}

func seeGEZero(p *Position, move Move) bool {
	return SeeGE(p, move, 0)
}

// based on Ethereal
func SeeGE(pos *Position, move Move, threshold int) bool {
	var from = move.From()
	var to = move.To()
	var movingPiece = move.MovingPiece()
	var capturedPiece = move.CapturedPiece()
	var promotionPiece = move.Promotion()

	var nextVictim = movingPiece
	if promotionPiece != Empty {
		nextVictim = promotionPiece
	}

	var balance = pieceValuesSEE[capturedPiece]
	if promotionPiece != Empty {
		balance += pieceValuesSEE[promotionPiece] - pieceValuesSEE[Pawn]
	}
	balance -= threshold

	if balance < 0 {
		return false
	}

	balance -= pieceValuesSEE[nextVictim]
	if balance >= 0 {
		return true
	}

	var occupied = pos.AllPieces()&^SquareMask[from] | SquareMask[to]
	if movingPiece == Pawn && to == pos.EpSquare {
		var capSq int
		if pos.WhiteMove {
			capSq = to - 8
		} else {
			capSq = to + 8
		}
		occupied &^= SquareMask[capSq]
	}

	var attackers = computeAttackers(pos, to, occupied) & occupied

	var bishops = pos.Bishops | pos.Queens
	var rooks = pos.Rooks | pos.Queens

	var side = sideToMove(pos) ^ 1

	for {
		var myAttackers = attackers & pos.Colours(side)
		if myAttackers == 0 {
			break
		}

		var attackerType, attackerFrom = getLeastValuableAttacker(pos, myAttackers)

		occupied &^= SquareMask[attackerFrom]

		if attackerType == Pawn || attackerType == Bishop || attackerType == Queen {
			attackers |= BishopAttacks(to, occupied) & bishops
		}
		if attackerType == Rook || attackerType == Queen {
			attackers |= RookAttacks(to, occupied) & rooks
		}

		attackers &= occupied

		side = side ^ 1

		balance = -balance - 1 - pieceValuesSEE[attackerType]
		if balance >= 0 {
			if attackerType == King &&
				(attackers&pos.Colours(side)) != 0 {
				side = side ^ 1
			}
			break
		}
	}

	return side != sideToMove(pos)
}

func sideToMove(p *Position) int {
	if p.WhiteMove {
		return SideWhite
	} else {
		return SideBlack
	}
}

func computeAttackers(pos *Position, sq int, occ uint64) uint64 {
	return (PawnAttacks(sq, true) & pos.Pawns & pos.Black) |
		(PawnAttacks(sq, false) & pos.Pawns & pos.White) |
		(KnightAttacks[sq] & pos.Knights) |
		(KingAttacks[sq] & pos.Kings) |
		(BishopAttacks(sq, occ) & (pos.Bishops | pos.Queens)) |
		(RookAttacks(sq, occ) & (pos.Rooks | pos.Queens))
}

func getLeastValuableAttacker(p *Position, attackers uint64) (attacker, from int) {
	if p.Pawns&attackers != 0 {
		return Pawn, FirstOne(p.Pawns & attackers)
	}
	if p.Knights&attackers != 0 {
		return Knight, FirstOne(p.Knights & attackers)
	}
	if p.Bishops&attackers != 0 {
		return Bishop, FirstOne(p.Bishops & attackers)
	}
	if p.Rooks&attackers != 0 {
		return Rook, FirstOne(p.Rooks & attackers)
	}
	if p.Queens&attackers != 0 {
		return Queen, FirstOne(p.Queens & attackers)
	}
	if p.Kings&attackers != 0 {
		return King, FirstOne(p.Kings & attackers)
	}
	return Empty, SquareNone
}
