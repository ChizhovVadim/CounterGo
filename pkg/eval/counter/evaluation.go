package eval

import (
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

// 2023/09/18 15:39:00 Finished Epoch 100
// 2023/09/18 15:39:00 Current validation cost is: 0.053783
// 2023/09/18 15:39:00 Train finished
var w = []int{11041, 13738, 48074, 44091, 49739, 46967, 65573, 80534, 145291, 145222, 3807, 8095, 0, 0, 0, 0, 0, 0, 0, 0, -3807, 124, 484, 386, -64, 1372, -1206, 1368, -2530, -94, -854, -285, 178, -887, -1588, -1444, -1791, -101, -706, 991, 2494, -1616, 1955, -2216, -32, 915, 4394, -106, 2416, -474, 3448, -1002, 2858, 965, 3390, 1575, 3373, 350, 5655, 920, 7691, 6485, 443, 10082, 4275, 9341, 10126, 8519, 0, 0, 0, 0, 0, 0, 0, 0, -4874, -1688, -722, 1365, 143, 122, 2195, 1912, 485, -1323, -126, 3160, 1371, 884, 1626, 2157, -920, -197, 3103, 203, 1662, 2295, 3441, 3252, 1822, 0, 3204, 899, 3522, 3094, 3827, 3430, 3315, 263, 3265, 1458, 3532, 3421, 4783, 4351, 1118, 357, 3943, 694, 4507, 2802, 8428, 2791, 4010, -211, 857, 2130, 9176, 362, 11474, 1613, -12947, -4953, -3990, 1349, -7065, 3568, 2378, 2024, 2722, 1217, 6216, 676, -21, 3020, 1419, 2044, 3580, 59, 3837, 168, 4369, 573, 2015, 2016, 1908, 952, 4498, 802, 3018, 2534, 3206, 2549, 4307, -1072, 1722, 2912, 2445, 2920, 3073, 3549, 3135, 577, 3117, 2906, 4566, 2913, 3804, 3755, 2127, 2626, 5000, 3023, 3963, 3281, 8901, 1117, -227, 3570, -1074, 3746, 958, 2944, -499, 4044, -1972, 2681, -4516, 4087, -4077, 4240, -104, 3840, -1107, 2875, -1035, 4349, 443, 3284, 2360, 2370, -4374, 4909, -933, 3560, 616, 3406, 1575, 3714, -1651, 4675, -93, 5227, 415, 5127, 1486, 4766, -1394, 6139, -688, 6871, 1267, 6303, 1989, 5873, 1064, 6980, 3194, 6393, 4204, 6298, 5190, 5667, 2837, 6869, 7109, 5778, 5714, 6237, 9306, 4214, 4581, 6313, 1813, 7733, 6309, 6263, 7793, 5983, 5579, 6623, 8933, 6627, 7658, 7231, 8455, 6473, 2950, -362, 3008, 389, 2179, 87, 1175, 7075, 3937, -698, 3148, 2180, 4611, 1261, 3515, 4234, 1874, 7502, 2480, 8275, 2955, 8504, 2558, 10431, 2234, 8234, 1649, 12112, 2064, 11481, -42, 15731, 4030, 9866, 970, 16075, 2244, 15057, 687, 17608, 37, 15244, 4239, 13919, 1923, 16902, 5788, 15611, 3126, 13022, -1962, 17197, 2734, 16027, 2354, 17956, 5979, 9822, 8673, 9894, 9666, 9617, 8048, 12972, 3516, -5588, 3294, -4969, -468, -4078, -907, -5248, 2020, -2045, -1386, -576, -1633, -1050, -5835, 742, -2290, -2, -3381, 617, -3790, 932, -5050, 2080, -230, 681, -5229, 2380, -1444, 1961, -74, 2425, -2074, 2503, 356, 3257, 630, 2884, 1679, 3047, 3489, 1730, 3175, 3881, 3340, 4010, 4701, 3697, 784, 174, 1215, 5470, 4047, 2482, 7962, 1568, 990, -4422, 714, 4171, 1289, 3047, 1099, 1147, -2226, -10154, -507, -2552, 712, 973, 1789, 3234, 3076, 3916, 3956, 4963, 4608, 5026, 5098, 5042, 6071, 3238, -1159, -6985, 127, -2633, 1600, -781, 2338, 1551, 3257, 3207, 3742, 4382, 4223, 4630, 4289, 5170, 4423, 5591, 4715, 5383, 4721, 5890, 6937, 4566, 4214, 6541, 6371, 2582, -5530, -6217, -3218, -1382, -1026, 1055, -345, 3602, 1054, 4695, 1442, 5709, 1255, 7026, 1614, 7214, 2236, 7630, 2836, 8226, 3531, 8455, 3645, 8882, 3686, 9317, 4570, 9034, 5442, 8782, 0, 0, -4700, -3985, -510, 4833, 879, -1803, 1256, 1951, 1782, 3147, 1652, 8617, 2251, 9699, 2579, 11008, 2946, 12170, 3075, 12707, 3018, 14338, 3206, 14910, 3145, 15579, 3093, 16110, 3459, 16046, 3015, 16241, 3125, 16033, 4296, 14539, 5145, 14139, 6852, 11416, 8041, 10129, 9276, 8960, 9302, 8596, 8282, 6649, 6208, 7327, 7005, 5145, 5133, 4782, 0, 0, 1645, -1783, 1680, -543, 3040, -1414, 3074, -12, 4023, 194, 1098, 359, -2816, -6, 0, 0, 3232, -1595, 1975, -1448, 830, -393, -167, 153, 304, 836, 4787, 2580, -4009, -124, 0, 0, 1207, 20, -2719, 495, -2622, -678, -2301, -977, 3106, 7, 149, 3457, -3074, -898, 0, 0, 1760, 862, -1051, 793, 668, -1106, 409, -1533, -1015, -423, -5348, 2703, -2835, 330, 0, 0, -2151, 1566, -1324, 966, -2144, 1039, -944, -1221, -3046, 738, -2007, 312, -4341, 1143, 0, 0, 3146, -1138, -1262, 63, -849, -1429, 799, -1990, 1817, -49, -1763, 4374, -3306, -546, 0, 0, 2150, -787, 340, -333, -473, -1169, -2658, -949, -2380, 710, -1164, 3882, -3330, -615, 0, 0, 1191, -2410, 2633, 47, 2016, -353, 1245, 168, 807, 1823, -1997, 4027, -2526, 423, 0, 0, 3113, -3378, 3598, -1820, 3740, -1593, 4221, -2214, 3024, 1007, -1867, 2816, -5736, -482, 0, 0, 3392, -1851, 3535, -2498, 164, 27, -1017, 1001, 576, 3175, 3821, 5328, -3965, -369, 0, 0, 4300, 435, -2071, 1460, 227, -1477, -1324, -311, -4219, 2811, 5325, 2328, -3586, -366, 0, 0, -869, 3200, -1251, 2886, -1659, 697, -2181, -49, -2809, -1533, -2734, 2942, -6291, -5, 0, 0, 730, -953, 107, 930, -41, 245, -125, -1615, -6136, 619, -758, 1933, -3222, -402, 0, 0, 3971, -256, -1232, 1318, -2000, 548, -941, -656, -1939, 2313, 2854, 549, -5965, 445, 0, 0, 3537, -568, 2192, -424, -711, -747, -3599, 358, -1801, 3578, -5662, 6090, -4083, -370, 0, 0, 3225, -5746, 3159, -1772, 825, -974, 661, -1138, -3300, 1512, 2178, 2765, -4594, -237, 0, 0, 0, 0, 0, 0, 0, 0, 1238, -858, 1269, -135, 36, 111, 204, 1150, 2005, -862, 566, 133, 1041, 676, 2135, 809, -48, 94, 1980, 496, 930, 1533, 1938, 1710, -857, 2461, 2467, 3764, 1869, 4000, 2139, 5800, 3121, 6862, 8092, 9979, 9693, 10369, 9155, 9237, 14473, 15907, 17502, 19045, 15084, 17434, 12818, 16936, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1859, 591, 4160, -46, 2136, 1715, 2690, 2072, 1210, 436, 3567, -570, 1670, 1081, 2243, 1414, 1174, 934, 975, 2240, 3592, 1888, 3425, 2767, -2279, 6719, 4369, 7037, 5911, 7218, 10823, 6371, 9727, 12879, 15076, 13016, 17731, 10731, 15048, 16034, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1643, -1819, -3310, 1757, -221, 435, 1231, 834, 2257, 1376, -1121, 2381, 0, 0, 0, 0, -174, 366, -77, 551, 700, 1713, 1493, 2583, 1143, 4849, 8654, 4513, 0, 0, 0, 0, 243, 1176, -1430, 1597, -590, 1670, 671, 1708, 2261, 1402, 3138, -509, 0, 0, 0, 0, -396, 661, -505, 1472, -1430, 4166, -616, 6441, 1636, 10403, 6303, 15290, 0, 0, -5102, 3849, 1115, 2687, -199, 1565, -1226, 709, -1384, -685, -1970, -541, 1676, -1694, 1741, -2274, 935, 6701, -1499, 6714, -1178, 3778, -2346, 1467, -1481, -800, 204, -1842, 3637, -2846, 2983, -3002, 6295, 10422, 2485, 11194, 3827, 5023, 179, 2423, -866, 277, -1175, -673, 1778, -1632, 5021, -1782, 14717, 20059, 10972, 16848, 11329, 8514, 5959, 4614, 5670, 1930, -212, 1759, -1658, 2694, 7408, 862, 26267, 28592, 17975, 21399, 14171, 16083, 7368, 9024, 8128, 3481, 10177, 341, 5382, 1537, -495, 3412, 347, -776, 5474, 317, 1243, 916, 693, 1207, 1011, 412, -693, 1104, -2624, 1957, -5353, 1721, -1702, -1276, 1166, -34, 3025, -380, 154, 1325, -734, 3709, -1831, 4819, -815, 5367, -3248, 4385, -3232, -2956, 5254, -2039, 1753, -719, 708, 3184, 277, 6456, -1022, 8824, -860, 9192, -2105, 7537, -5546, -5398, 1679, -4324, 4399, -1183, 2451, 5357, 2092, 9951, 1036, 13316, 2105, 13712, 3361, 11872, -9094, -12983, -1020, -8700, 7243, -3547, 4518, 6019, 6593, 12240, 8735, 14177, 5294, 16292, 7479, 16069, -3, -1689, -905, -863, 4003, -88, 1880, 1270, -547, -1699, -444, 67, -651, -599, -951, -84, -468, -415, -1114, 171, 3098, 241, 484, 1942, 788, 2455, -6908, -5410, -2815, -3911, -4274, -4661, -6940, -4274, -4624, -5099, -6496, -1775, -6515, 2567, -1512, -4573, 2798, 580}

const (
	totalPhase        = 24
	scaleFactorNormal = 128
)

type EvaluationService struct {
	WeightList
	phase              int
	passed             uint64
	pieceCount         [COLOUR_NB][PIECE_NB]int
	attacked           [COLOUR_NB]uint64
	attackedBy2        [COLOUR_NB]uint64
	attackedBy         [COLOUR_NB][PIECE_NB]uint64
	pawnAttacksBy2     [COLOUR_NB]uint64
	kingSq             [COLOUR_NB]int
	kingAttackersCount [COLOUR_NB]int
	kingpawnTable      []kingPawnEntry
}

type kingPawnEntry struct {
	wpawns, bpawns uint64
	wking, bking   int
	score          Score
	passed         uint64
}

func NewEvaluationService() *EvaluationService {
	var e = &EvaluationService{
		kingpawnTable: make([]kingPawnEntry, 1<<16),
	}
	e.WeightList.init(featureSize)
	e.WeightList.InitWeights(w)
	return e
}

func (e *EvaluationService) Evaluate(p *Position) int {
	//init
	for pt := Pawn; pt <= King; pt++ {
		e.pieceCount[SideWhite][pt] = 0
		e.pieceCount[SideBlack][pt] = 0
		e.attackedBy[SideWhite][pt] = 0
		e.attackedBy[SideBlack][pt] = 0
	}

	e.passed = 0

	e.kingAttackersCount[SideWhite] = 0
	e.kingAttackersCount[SideBlack] = 0

	e.kingSq[SideWhite] = FirstOne(p.Kings & p.White)
	e.kingSq[SideBlack] = FirstOne(p.Kings & p.Black)

	e.pieceCount[SideWhite][Pawn] = PopCount(p.Pawns & p.White)
	e.pieceCount[SideBlack][Pawn] = PopCount(p.Pawns & p.Black)

	e.attackedBy[SideWhite][Pawn] = AllWhitePawnAttacks(p.Pawns & p.White)
	e.attackedBy[SideBlack][Pawn] = AllBlackPawnAttacks(p.Pawns & p.Black)

	e.pawnAttacksBy2[SideWhite] = UpLeft(p.Pawns&p.White) & UpRight(p.Pawns&p.White)
	e.pawnAttacksBy2[SideBlack] = DownLeft(p.Pawns&p.Black) & DownRight(p.Pawns&p.Black)

	e.attackedBy[SideWhite][King] = KingAttacks[e.kingSq[SideWhite]]
	e.attackedBy[SideBlack][King] = KingAttacks[e.kingSq[SideBlack]]

	e.attacked[SideWhite] = e.attackedBy[SideWhite][King]
	e.attacked[SideBlack] = e.attackedBy[SideBlack][King]

	e.attackedBy2[SideWhite] = e.attackedBy[SideWhite][Pawn] & e.attacked[SideWhite]
	e.attacked[SideWhite] |= e.attackedBy[SideWhite][Pawn]

	e.attackedBy2[SideBlack] = e.attackedBy[SideBlack][Pawn] & e.attacked[SideBlack]
	e.attacked[SideBlack] |= e.attackedBy[SideBlack][Pawn]

	//eval
	var score Score

	var pawnKingKey = murmurMix(p.Pawns&p.White,
		murmurMix(p.Pawns&p.Black,
			murmurMix(p.Kings&p.White,
				p.Kings&p.Black)))
	var pke = &e.kingpawnTable[pawnKingKey%uint64(len(e.kingpawnTable))]
	if e.tuning ||
		!(pke.wpawns == p.Pawns&p.White &&
			pke.bpawns == p.Pawns&p.Black &&
			pke.wking == e.kingSq[SideWhite] &&
			pke.bking == e.kingSq[SideBlack]) {

		var pawnKingScore = e.evalPawnsAndKings(p, SideWhite) + e.evalPawnsAndKings(p, SideBlack)
		score += pawnKingScore

		pke.wpawns = p.Pawns & p.White
		pke.bpawns = p.Pawns & p.Black
		pke.wking = e.kingSq[SideWhite]
		pke.bking = e.kingSq[SideBlack]
		pke.score = pawnKingScore
		pke.passed = e.passed
	} else {
		score += pke.score
		e.passed = pke.passed
	}

	score += e.evalFirstPass(p, SideWhite) + e.evalFirstPass(p, SideBlack)
	score += e.evalSecondPass(p, SideWhite) + e.evalSecondPass(p, SideBlack)

	score += e.Value(fMinorBehindPawn,
		PopCount((p.Knights|p.Bishops)&p.White&Down(p.Pawns))-
			PopCount((p.Knights|p.Bishops)&p.Black&Up(p.Pawns)))

	score += e.Value(fMinorProtected,
		PopCount((p.Knights|p.Bishops)&p.White&e.attackedBy[SideWhite][Pawn])-
			PopCount((p.Knights|p.Bishops)&p.Black&e.attackedBy[SideBlack][Pawn]))

	score += e.Value(fPawnValue, e.pieceCount[SideWhite][Pawn]-e.pieceCount[SideBlack][Pawn])
	score += e.Value(fKnightValue, e.pieceCount[SideWhite][Knight]-e.pieceCount[SideBlack][Knight])
	score += e.Value(fBishopValue, e.pieceCount[SideWhite][Bishop]-e.pieceCount[SideBlack][Bishop])
	score += e.Value(fRookValue, e.pieceCount[SideWhite][Rook]-e.pieceCount[SideBlack][Rook])
	score += e.Value(fQueenValue, e.pieceCount[SideWhite][Queen]-e.pieceCount[SideBlack][Queen])

	if p.WhiteMove {
		score += e.Value(fTempo, 1)
	} else {
		score += e.Value(fTempo, -1)
	}

	//mix score
	var phase = e.pieceCount[SideWhite][Knight] + e.pieceCount[SideBlack][Knight] +
		e.pieceCount[SideWhite][Bishop] + e.pieceCount[SideBlack][Bishop] +
		2*(e.pieceCount[SideWhite][Rook]+e.pieceCount[SideBlack][Rook]) +
		4*(e.pieceCount[SideWhite][Queen]+e.pieceCount[SideBlack][Queen])
	if phase > totalPhase {
		phase = totalPhase
	}
	e.phase = phase

	var result = (score.Mg()*phase + score.Eg()*(totalPhase-phase)) / (totalPhase * 100)

	var strongSide int
	if result > 0 {
		strongSide = SideWhite
	} else {
		strongSide = SideBlack
	}
	result = result * e.computeFactor(strongSide, p) / scaleFactorNormal

	if !p.WhiteMove {
		result = -result
	}

	return result
}

func (e *EvaluationService) computeFactor(strongSide int, p *Position) int {
	var result = scaleFactorNormal

	result = result * (200 - p.Rule50) / 200

	var strongSidePawns = e.pieceCount[strongSide][Pawn]
	var x = 8 - strongSidePawns
	result = result * (128 - x*x) / 128

	/*const (
		QueenSideBB = FileAMask | FileBMask | FileCMask | FileDMask
		KingSideBB  = FileEMask | FileFMask | FileGMask | FileHMask
	)
	if p.Colours(strongSide)&p.Pawns&QueenSideBB == 0 ||
		p.Colours(strongSide)&p.Pawns&KingSideBB == 0 {
		result = result * 85 / 100
	}*/

	if strongSidePawns == 0 {
		var strongMinors = e.pieceCount[strongSide][Knight] + e.pieceCount[strongSide][Bishop]
		var strongMajors = e.pieceCount[strongSide][Rook] + 2*e.pieceCount[strongSide][Queen]

		var weakSide = strongSide ^ 1
		var weakMinors = e.pieceCount[weakSide][Knight] + e.pieceCount[weakSide][Bishop]
		var weakMajors = e.pieceCount[weakSide][Rook] + 2*e.pieceCount[weakSide][Queen]

		var balance = 4*(strongMinors-weakMinors) + 6*(strongMajors-weakMajors)

		if strongMajors == 0 && strongMinors <= 1 {
			return scaleFactorNormal / 16
		} else if balance <= 4 {
			return scaleFactorNormal / 4
		}
	}

	if e.pieceCount[SideWhite][Bishop] == 1 &&
		e.pieceCount[SideBlack][Bishop] == 1 &&
		onlyOne(p.Bishops&darkSquares) {
		if p.Knights|p.Rooks|p.Queens == 0 {
			result = result * 1 / 2
		}
	}

	return result
}

func (e *EvaluationService) evalPawnsAndKings(p *Position, side int) Score {
	var s Score
	var x uint64
	var sq int

	var sign int
	if side == SideWhite {
		sign = 1
	} else {
		sign = -1
	}
	var US = side
	var THEM = side ^ 1
	var friendly = p.Colours(US)
	var enemy = p.Colours(THEM)

	for x = p.Pawns & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		s += e.addPst32(fPawnPST, US, sq, sign)
		if PawnAttacksNew(THEM, sq)&friendly&p.Pawns != 0 {
			s += e.addPst32(fPawnProtected, US, sq, sign)
		}
		if adjacentFilesMask[File(sq)]&rankMasks[Rank(sq)]&friendly&p.Pawns != 0 {
			s += e.addPst32(fPawnDuo, US, sq, sign)
		}
		if adjacentFilesMask[File(sq)]&friendly&p.Pawns == 0 {
			s += e.Value(fPawnIsolated, sign)
		}
		if FileMask[File(sq)]&^SquareMask[sq]&friendly&p.Pawns != 0 {
			s += e.Value(fPawnDoubled, sign)
		}
		if pawnPassedMask[US][sq]&enemy&p.Pawns == 0 {
			e.passed |= SquareMask[sq]
		}
	}
	{
		sq = FirstOne(p.Kings & friendly)
		s += e.addPst32(fKingPST, US, sq, sign)
		var file = limit(File(sq), FileB, FileG)
		var mask = friendly & p.Pawns & forwardRanksMasks[US][Rank(sq)]
		for f := file - 1; f <= file+1; f++ {
			var ours = FileMask[f] & mask
			var ourDist int
			if ours == 0 {
				ourDist = 7
			} else {
				ourDist = relativeRankOf(US, backmost(US, ours))

				/*ourDist = Rank(sq) - Rank(backmost(US, ours))
				if ourDist < 0 {
					ourDist = -ourDist
				}*/
			}
			var index = boolToInt(f == File(sq))*64 + f*8 + ourDist
			s += e.Value(fKingShelter+index, sign)
		}
	}
	return s
}

func (e *EvaluationService) evalFirstPass(p *Position, side int) Score {
	var s Score
	var x, attacks uint64
	var sq int

	var allPieces = p.AllPieces()
	var sign int
	if side == SideWhite {
		sign = 1
	} else {
		sign = -1
	}
	var US = side
	var THEM = side ^ 1
	var friendly = p.Colours(US)
	var enemy = p.Colours(THEM)

	var mobilityArea = ^(p.Pawns&friendly | e.attackedBy[THEM][Pawn])
	var kingArea = kingAreaMasks[THEM][e.kingSq[THEM]] &^ e.pawnAttacksBy2[THEM]
	for x = p.Knights & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		e.pieceCount[US][Knight]++
		s += e.addPst32(fKnightPST, US, sq, sign)
		attacks = KnightAttacks[sq]
		s += e.Value(fKnightMobility+PopCount(attacks&mobilityArea), sign)
		e.attackedBy2[US] |= e.attacked[US] & attacks
		e.attacked[US] |= attacks
		e.attackedBy[US][Knight] |= attacks
		if attacks&kingArea != 0 {
			e.kingAttackersCount[THEM]++
		}
		if outpostSquares[side]&SquareMask[sq] != 0 &&
			outpostSquareMasks[US][sq]&enemy&p.Pawns == 0 {
			s += e.Value(fKnightOutpost, sign)
		}
	}
	for x = p.Bishops & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		e.pieceCount[US][Bishop]++
		s += e.addPst32(fBishopPST, US, sq, sign)
		attacks = BishopAttacks(sq, allPieces)
		s += e.Value(fBishopMobility+PopCount(attacks&mobilityArea), sign)
		e.attackedBy2[US] |= e.attacked[US] & attacks
		e.attacked[US] |= attacks
		e.attackedBy[US][Bishop] |= attacks
		if attacks&kingArea != 0 {
			e.kingAttackersCount[THEM]++
		}
		if side == SideWhite {
			s += e.Value(fBishopRammedPawns, sign*PopCount(
				sameColorSquares(sq)&p.Pawns&p.White&Down(p.Pawns&p.Black)))
		} else {
			s += e.Value(fBishopRammedPawns, sign*PopCount(
				sameColorSquares(sq)&p.Pawns&p.Black&Up(p.Pawns&p.White)))
		}
	}
	for x = p.Rooks & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		e.pieceCount[US][Rook]++
		s += e.addPst32(fRookPST, US, sq, sign)
		attacks = RookAttacks(sq, allPieces&^(friendly&p.Rooks))
		s += e.Value(fRookMobility+PopCount(attacks&mobilityArea), sign)
		e.attackedBy2[US] |= e.attacked[US] & attacks
		e.attacked[US] |= attacks
		e.attackedBy[US][Rook] |= attacks
		if attacks&kingArea != 0 {
			e.kingAttackersCount[THEM]++
		}
		var mask = FileMask[File(sq)]
		if (mask & friendly & p.Pawns) == 0 {
			if (mask & p.Pawns) == 0 {
				s += e.Value(fRookOpen, sign)
			} else {
				s += e.Value(fRookSemiopen, sign)
			}
		}
	}
	for x = p.Queens & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		e.pieceCount[US][Queen]++
		s += e.addPst32(fQueenPST, US, sq, sign)
		attacks = QueenAttacks(sq, allPieces)
		s += e.Value(fQueenMobility+PopCount(attacks&mobilityArea), sign)
		e.attackedBy2[US] |= e.attacked[US] & attacks
		e.attacked[US] |= attacks
		e.attackedBy[US][Queen] |= attacks
		if attacks&kingArea != 0 {
			e.kingAttackersCount[THEM]++
		}
	}
	if e.pieceCount[US][Bishop] >= 2 {
		s += e.Value(fBishopPair, sign)
	}
	return s
}

func (e *EvaluationService) addPst32(index, side, sq, sign int) Score {
	return e.Value(index+relativeSq32(side, sq), sign)
}

func (e *EvaluationService) addPst12(index, side, sq, sign int) Score {
	return e.Value(index+file4(sq), sign) +
		e.Value(index+4+relativeRankOf(side, sq), sign)
}

var kingAttackWeight = [...]int{2, 4, 8, 12, 13, 14, 15, 16}

func (e *EvaluationService) evalSecondPass(p *Position, side int) Score {
	var s Score
	var x uint64
	var sq int
	var sign int
	var forward int
	if side == SideWhite {
		sign = 1
		forward = 8
	} else {
		sign = -1
		forward = -8
	}
	var allPieces = p.AllPieces()
	var US = side
	var friendly = p.Colours(US)
	var THEM = side ^ 1

	for x = e.passed & friendly; x != 0; x &= x - 1 {
		sq = FirstOne(x)
		if forwardFileMasks[US][sq]&^SquareMask[sq]&p.Pawns != 0 {
			continue
		}
		var rank = relativeRankOf(US, sq)
		var keySq = sq + forward
		var bitboard = SquareMask[keySq]
		var canAdvance = allPieces&bitboard == 0
		var safeAdvance = e.attacked[THEM]&bitboard == 0
		var index = 0
		if canAdvance {
			index |= 1
		}
		if safeAdvance {
			index |= 2
		}
		s += e.Value(fPawnPassed+index*8+rank, sign)
		//if forwardFileMasks[US][sq]&^SquareMask[sq]&p.Pawns == 0 {
		//s += e.Value(fPassedFriendlyDistance+rank, sign*distanceBetween[keySq][e.kingSq[US]])
		//s += e.Value(fPassedEnemyDistance+rank, sign*distanceBetween[keySq][e.kingSq[THEM]])
		//}

		var r = Max(0, relativeRankOf(side, sq)-Rank3)
		s += e.Value(fPassedFriendlyDistance+8*r+distanceBetween[keySq][e.kingSq[US]], sign)
		s += e.Value(fPassedEnemyDistance+8*r+distanceBetween[keySq][e.kingSq[THEM]], sign)
	}

	{
		//king safety

		var kingSq = e.kingSq[US]
		var kingArea = kingAreaMasks[US][kingSq]

		var weak = e.attacked[THEM] &
			^e.attackedBy2[US] &
			(^e.attacked[US] | e.attackedBy[US][Queen] | e.attackedBy[US][King])

		var safe = ^p.Colours(THEM) &
			(^e.attacked[US] | (weak & e.attackedBy2[THEM]))

		var occupied = p.AllPieces()
		var knightThreats = KnightAttacks[kingSq]
		var bishopThreats = BishopAttacks(kingSq, occupied)
		var rookThreats = RookAttacks(kingSq, occupied)
		var queenThreats = bishopThreats | rookThreats

		var knightChecks = knightThreats & safe & e.attackedBy[THEM][Knight]
		var bishopChecks = bishopThreats & safe & e.attackedBy[THEM][Bishop]
		var rookChecks = rookThreats & safe & e.attackedBy[THEM][Rook]
		var queenChecks = queenThreats & safe & e.attackedBy[THEM][Queen]

		var val = sign * kingAttackWeight[Min(len(kingAttackWeight)-1, e.kingAttackersCount[US])]

		s += e.Value(fSafetyWeakSquares, val*PopCount(weak&kingArea))
		s += e.Value(fSafetySafeQueenCheck, val*PopCount(queenChecks))
		s += e.Value(fSafetySafeRookCheck, val*PopCount(rookChecks))
		s += e.Value(fSafetySafeBishopCheck, val*PopCount(bishopChecks))
		s += e.Value(fSafetySafeKnightCheck, val*PopCount(knightChecks))
	}

	{
		// threats

		var minors = friendly & (p.Knights | p.Bishops)
		var rooks = friendly & p.Rooks
		var queens = friendly & p.Queens

		var attacksByPawns = e.attackedBy[THEM][Pawn]
		var attacksByMinors = e.attackedBy[THEM][Knight] | e.attackedBy[THEM][Bishop]
		var attacksByMajors = e.attackedBy[THEM][Rook] | e.attackedBy[THEM][Queen]

		var poorlyDefended = (e.attacked[THEM] & ^e.attacked[US]) |
			(e.attackedBy2[THEM] & ^e.attackedBy2[US] & ^e.attackedBy[US][Pawn])

		s += e.Value(fThreatWeakPawn, sign*PopCount(friendly&p.Pawns & ^attacksByPawns & poorlyDefended))
		s += e.Value(fThreatMinorAttackedByPawn, sign*PopCount(minors&attacksByPawns))
		s += e.Value(fThreatMinorAttackedByMinor, sign*PopCount(minors&attacksByMinors))
		s += e.Value(fThreatMinorAttackedByMajor, sign*PopCount(minors&poorlyDefended&attacksByMajors))
		s += e.Value(fThreatRookAttackedByLesser, sign*PopCount(rooks&(attacksByPawns|attacksByMinors)))
		s += e.Value(fThreatMinorAttackedByKing, sign*PopCount(minors&poorlyDefended&e.attackedBy[THEM][King]))
		s += e.Value(fThreatRookAttackedByKing, sign*PopCount(rooks&poorlyDefended&e.attackedBy[THEM][King]))
		s += e.Value(fThreatQueenAttackedByOne, sign*PopCount(queens&e.attacked[THEM]))
	}

	return s
}

func (e *EvaluationService) EnableTuning() {
	e.tuning = true
}

func (e *EvaluationService) StartingWeights() []float64 {
	var material = []float64{100, 100, 325, 325, 325, 325, 500, 500, 1000, 1000}
	var result = make([]float64, 2*featureSize)
	copy(result, material)
	return result
}

func (e *EvaluationService) ComputeFeatures(pos *Position) domain.TuneEntry {
	for i := range e.values {
		e.values[i] = 0
	}
	e.Evaluate(pos)
	var result = domain.TuneEntry{
		Features:         e.WeightList.Features(),
		MgPhase:          float32(e.phase) / totalPhase,
		WhiteStrongScale: float32(e.computeFactor(SideWhite, pos)) / scaleFactorNormal,
		BlackStrongScale: float32(e.computeFactor(SideBlack, pos)) / scaleFactorNormal,
	}
	result.EgPhase = 1 - result.MgPhase
	return result
}
