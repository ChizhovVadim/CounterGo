package eval

import (
	"github.com/ChizhovVadim/CounterGo/internal/domain"
	"github.com/ChizhovVadim/CounterGo/internal/math"
	. "github.com/ChizhovVadim/CounterGo/pkg/common"
)

const (
	totalPhase        = 24
	scaleFactorNormal = 128
	weightScale       = 10_000 //see internal::tuner::model::Print()
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
	//2024/08/28 10:48:04 mse cost: 0.052704
	var w = []int{5072, 5217, 18158, 20472, 19978, 18228, 25015, 28386, 56677, 45242, 1871, 4745, 0, 0, 0, 0, 0, 0, 0, 0, 159, 4355, 1483, 4546, 638, 5112, 929, 4496, 813, 4255, 1396, 3910, 1479, 3414, 1207, 3681, 783, 4656, 1154, 4807, 2742, 3196, 3160, 2469, 1853, 5466, 3566, 4355, 2673, 3941, 3803, 2964, 2279, 4837, 3795, 5181, 2962, 4395, 4672, 3705, 4675, 7519, 2435, 8520, 2330, 8568, 6246, 7200, 0, 0, 0, 0, 0, 0, 0, 0, 2980, 2845, 4122, 5567, 4158, 5321, 4742, 6739, 4444, 4398, 5244, 5409, 4792, 5972, 5333, 7078, 3677, 5751, 5655, 5381, 5149, 6382, 6114, 7601, 5938, 6130, 6311, 6631, 6669, 7515, 6545, 8199, 6083, 6632, 6199, 6848, 6734, 7760, 7177, 8506, 5712, 6306, 5934, 6798, 6200, 7405, 8174, 7829, 6280, 5520, 5516, 6257, 9071, 5379, 9099, 7593, -5293, 2713, -4097, 9418, -7335, 8862, 8657, 7131, 6115, 9402, 7313, 8605, 5327, 9273, 5445, 9547, 7280, 8528, 7412, 8212, 7258, 8876, 6249, 9235, 6143, 9075, 7717, 8783, 6782, 9573, 6946, 9627, 6955, 8410, 6477, 9563, 6579, 10248, 7158, 9779, 6536, 8982, 6569, 9681, 7992, 9013, 7548, 10655, 7126, 9140, 8003, 9610, 7371, 9892, 7846, 9722, 5119, 9625, 3426, 10735, 5617, 10321, 6183, 10089, 4179, 10027, 2053, 10677, -1362, 10506, 1674, 10253, 8026, 15025, 7842, 16132, 8928, 14911, 9429, 14773, 6376, 15797, 7689, 15733, 8558, 15277, 8623, 15488, 7104, 16853, 7759, 16483, 8075, 16371, 8413, 16303, 7782, 17369, 8727, 17511, 9219, 17254, 9244, 17048, 9650, 17723, 9888, 17922, 10828, 17897, 11359, 17028, 9155, 18454, 12147, 17338, 11621, 16718, 12997, 16391, 11162, 18359, 9937, 18165, 12348, 17567, 13892, 16783, 14162, 17485, 13843, 17940, 13894, 17990, 13938, 18087, 20515, 29667, 20680, 29035, 19473, 28540, 19899, 31430, 19592, 30474, 20365, 29272, 21002, 29646, 20978, 30816, 19512, 32767, 19742, 34483, 20198, 33836, 20191, 34079, 19521, 35998, 18994, 37223, 18928, 37611, 18460, 38576, 20678, 36820, 19105, 39278, 18726, 39277, 18505, 41260, 19269, 37238, 20567, 38434, 17897, 40762, 20120, 40511, 20091, 38215, 17034, 41311, 19893, 41210, 18144, 42513, 20359, 36227, 24205, 34318, 24934, 35148, 24965, 37865, 127, -3049, 1878, -3101, -67, -3843, 86, -4638, 1318, -1182, -138, -847, -986, -1286, -3766, -191, -932, -281, 70, 214, -1323, 114, -2512, 788, -753, 989, -294, 1322, 545, 793, -2508, 1797, -1854, 1274, 2323, 2394, -2536, 2796, -4381, 3278, 2930, -59, -1520, 4382, -2356, 4367, -5170, 5094, -2612, -938, -1890, 4905, -936, 4316, -866, 4776, -110, -4193, 239, 145, 4403, -1138, 1219, 1054, 5270, 444, 6094, 5536, 6709, 7891, 7426, 9495, 8154, 10152, 8701, 11393, 9206, 11237, 9680, 11212, 10243, 10738, 5193, 4188, 5670, 7346, 6864, 8782, 7007, 10595, 7727, 11322, 7971, 12070, 8376, 12608, 8244, 13007, 8671, 12927, 8889, 13149, 8628, 13281, 9646, 13196, 8455, 13786, 9858, 12992, 7109, 11752, 5403, 15719, 6703, 18006, 7321, 18831, 8249, 19813, 8458, 21061, 8159, 21321, 8505, 21625, 8978, 22088, 9306, 22320, 9693, 22292, 10250, 22648, 9816, 23391, 10313, 23288, 11618, 21976, 27228, 12404, 24893, 27926, 20764, 30833, 21817, 30277, 22023, 33100, 22072, 35517, 22257, 37693, 22499, 39011, 22660, 38829, 22690, 40151, 22700, 40320, 22876, 40841, 22870, 41383, 23001, 42269, 22993, 42643, 22543, 42618, 22815, 42873, 22567, 43097, 22946, 42073, 22783, 43440, 23027, 41680, 23771, 41191, 25852, 39072, 23090, 39777, 24544, 38483, 28446, 36206, 30323, 33338, 23366, 36193, 0, 0, 1519, -1479, 1536, -1054, 1540, -1222, 172, 189, 1039, 468, -6255, 1616, -3888, 511, 0, 0, 1621, -1183, 485, -389, 187, -675, -2335, 224, -1792, 1419, -3689, 1937, -3084, 366, 0, 0, 391, 52, -1790, 369, -1120, 63, -1231, 213, -832, 510, -1764, 2117, -3073, -167, 0, 0, 870, -495, -543, 657, 238, -197, 18, -422, -1181, -171, 2038, -907, -1336, 367, 0, 0, -1143, 923, -1118, 44, -1913, 582, -1372, 488, -6360, 1903, -1397, -102, -2853, 410, 0, 0, 1767, -1091, -1245, 428, -650, -883, 1034, -1724, 3572, -1728, -299, -256, -2351, -256, 0, 0, 1302, -154, 1047, -703, 306, -1035, -248, -399, 165, 367, -1407, -74, -2065, 170, 0, 0, 1365, -2579, 1456, -748, 1187, -1115, -24, -312, -1267, 197, -9981, 5535, -2539, 313, 0, 0, 2490, -4530, 1452, -2077, 1375, -1845, 1024, -1435, 1829, -2013, -10920, 5506, -4974, 322, 0, 0, 1932, -425, 984, -429, -276, -787, -42, -272, 1922, 1559, -6293, 5490, -3512, 206, 0, 0, 2429, 132, -1018, 1003, -492, 974, -950, 117, -4843, 2864, -1740, 4079, -2340, 233, 0, 0, -1496, 3231, -1752, 1414, -1138, 1326, -1791, 548, -3205, 2457, -3288, -403, -3099, -192, 0, 0, 196, -14, -56, 500, 50, -121, 424, -101, -4365, 1574, -348, 2161, -1680, -178, 0, 0, 1845, -284, -402, 823, 50, 143, -574, -162, -815, 415, -1373, 1596, -3101, 301, 0, 0, 1779, -328, 1331, -883, 370, -1144, -903, -402, -1374, 1316, 895, 4051, -3063, 35, 0, 0, 3072, -5109, 2179, -2873, 715, -2344, -1698, -781, -1187, 1078, -13712, 10542, -4028, 7, 0, 0, 0, 0, 0, 0, 0, 0, 958, -529, 1156, -338, 612, 159, -113, 932, 572, -547, 971, 259, 288, 803, 1126, 439, 131, 461, 929, 449, 1016, 781, 1145, 1606, 2143, 2006, 386, 2926, 2758, 3507, 1535, 3601, -261, 9035, 10635, 8974, -2018, 13313, 11491, 5162, -9125, 30284, 7715, 34271, 24354, 24949, 9076, 21951, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1075, -26, 2338, 247, 1450, 1175, 1392, 1599, 642, 207, 1707, -178, 1022, 428, 1423, 848, 1263, 717, 1009, 1195, 1869, 1237, 2698, 1462, 415, 5782, 2477, 4983, 4023, 5147, 4244, 5861, 19782, 12095, 14185, 8174, 13810, 11304, 9197, 8700, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1220, 1285, -2406, 1519, -1341, 896, 454, 233, 1745, 1851, 522, 141, 0, 0, 0, 0, -343, 268, -152, 464, 303, 1047, 890, 1610, 2382, 2241, 6843, 2928, 0, 0, 0, 0, 126, 64, -484, 619, -375, 425, -71, 524, 2011, 562, 4575, -1659, 0, 0, 0, 0, -21, -77, -863, 953, -1000, 2431, -113, 3708, 1161, 6698, -1047, 16754, 0, 0, -1615, 1815, -1413, 2288, -300, 1329, 112, -80, -63, -233, -200, -410, 1405, -1039, 1728, -1564, -2669, 5162, -5479, 5968, -2377, 3414, -769, 1469, -334, -175, 1880, -711, 3416, -1739, 2405, -802, 404, 7944, 2255, 8069, 1807, 4359, -899, 2926, -357, 531, -26, -359, 3242, -1495, 2306, -1776, 15160, 10554, 14086, 11703, 24098, 3803, 5503, 4054, 22, 2265, -1319, 1731, -1307, 1763, -123, 1971, -9494, 32653, -4520, 23191, 16339, 11180, 13361, 5990, 16596, 2591, 6897, 2617, 2673, 3183, -1855, 6502, 28, -729, 2261, 312, -592, 999, -117, 623, 537, 109, 71, 484, -1616, 2054, -2252, 2234, -3377, 513, 532, 19, 106, 765, 871, 1063, -937, 2682, -1285, 4622, -1268, 5285, -2699, 5779, -3353, -1681, 2363, -1852, 1298, 467, -16, 3531, -60, 5786, 411, 7397, -721, 8478, -1198, 9376, -2387, -3881, 2543, -3665, 3275, 397, 1748, 5139, 691, 9408, 511, 12124, 2821, 12089, 4147, 11288, -5542, -7052, 2087, -5810, 4707, -748, 2409, 6714, 2423, 10575, 5065, 12365, 4444, 14622, 7438, 12266, -463, -1371, -285, -142, 2787, -387, 1278, 692, -927, -1183, -257, 69, -514, -357, -988, 12, -170, -417, -612, 16, 2186, -369, 295, 1241, 572, 1601, -4204, -4015, -2083, -3157, -2352, -3661, -5045, -3001, -3919, -1997, -4853, -1513, -3994, -3914, -1345, -3447, 3965, 1596}
	e.WeightList.init(featureSize)
	e.WeightList.InitWeights(w)
	return e
}

func (e *EvaluationService) Evaluate(p *Position) int {
	const CentipawnScale = 144 // Считаем по определению, что 100 сантипшек это вероятность выиграть 1/3
	// == 100*reverseSigmoid(2/3)
	var res = e.evalCore(p) * CentipawnScale / weightScale
	if !p.WhiteMove {
		res = -res
	}
	return res
}

func (e *EvaluationService) evalCore(p *Position) int {
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

	var result = (score.Mg()*phase + score.Eg()*(totalPhase-phase)) / totalPhase

	var strongSide int
	if result > 0 {
		strongSide = SideWhite
	} else {
		strongSide = SideBlack
	}
	result = result * e.computeFactor(strongSide, p) / scaleFactorNormal

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

func (e *EvaluationService) FeatureSize() int {
	return featureSize
}

func (e *EvaluationService) ComputeFeatures(pos *Position) domain.TuneEntry {
	e.tuning = true
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

func (e *EvaluationService) EvaluateProb(p *Position) float64 {
	return math.Sigmoid(float64(e.evalCore(p)) / weightScale)
}
