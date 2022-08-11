package eval

const (
	fPawnValue = iota
	fKnightValue
	fBishopValue
	fRookValue
	fQueenValue
	fBishopPair
	fPawnPST
	fKnightPST
	fBishopPST
	fRookPST
	fQueenPST
	fKingPST
	fKnightMobility
	fBishopMobility
	fRookMobility
	fQueenMobility
	fPassedPawn
	fPassedCanMove
	fPassedSafeMove
	fPassedEnemyKing
	fPassedOwnKing
	fPawnDuo
	fPawnProtected
	fPawnIsolated
	fPawnDoubled
	fKingShield
	fSafetyWeakSquares
	fSafetyKnightCheck
	fSafetyBishopCheck
	fSafetyRookCheck
	fSafetyQueenCheck
	fKingQueenTropism
	fThreatMinorAttackedByPawn
	fThreatMinorAttackedByMinor
	fThreatMinorAttackedByMajor
	fThreatRookAttackedByLesser
	fThreatMinorAttackedByKing
	fThreatRookAttackedByKing
	fThreatQueenAttackedByOne
	fThreatWeakPawn
	fKnightOutpost
	fMinorProtected
	fMinorBehindPawn
	fBishopRammedPawns
	fRookOpen
	fRookSemiopen
	fTempo
	fSize
)

type FeatureInfo struct {
	Name       string
	Size       int
	StartIndex int
}

var infos = [fSize]FeatureInfo{
	fPawnValue:                  {Name: "PawnValue"},
	fKnightValue:                {Name: "KnightValue"},
	fBishopValue:                {Name: "BishopValue"},
	fRookValue:                  {Name: "RookValue"},
	fQueenValue:                 {Name: "QueenValue"},
	fBishopPair:                 {Name: "BishopPair"},
	fPawnPST:                    {Name: "PawnPST", Size: 32},
	fKnightPST:                  {Name: "KnightPST", Size: 32},
	fBishopPST:                  {Name: "BishopPST", Size: 32},
	fRookPST:                    {Name: "RookPST", Size: 32},
	fQueenPST:                   {Name: "QueenPST", Size: 32},
	fKingPST:                    {Name: "KingPST", Size: 32},
	fKnightMobility:             {Name: "KnightMobility", Size: 9},
	fBishopMobility:             {Name: "BishopMobility", Size: 14},
	fRookMobility:               {Name: "RookMobility", Size: 15},
	fQueenMobility:              {Name: "QueenMobility", Size: 28},
	fPassedPawn:                 {Name: "PassedPawn", Size: 5},
	fPassedCanMove:              {Name: "PassedCanMove", Size: 5},
	fPassedSafeMove:             {Name: "PassedSafeMove", Size: 5},
	fPassedEnemyKing:            {Name: "PassedEnemyKing", Size: 5 * 8},
	fPassedOwnKing:              {Name: "PassedOwnKing", Size: 5 * 8},
	fPawnDuo:                    {Name: "PawnDuo", Size: 32},
	fPawnProtected:              {Name: "PawnProtected", Size: 32},
	fPawnIsolated:               {Name: "PawnIsolated"},
	fPawnDoubled:                {Name: "PawnDoubled"},
	fKingShield:                 {Name: "KingShield", Size: 12},
	fSafetyWeakSquares:          {Name: "SafetyWeakSquares"},
	fSafetyKnightCheck:          {Name: "SafetyKnightCheck"},
	fSafetyBishopCheck:          {Name: "SafetyBishopCheck"},
	fSafetyRookCheck:            {Name: "SafetyRookCheck"},
	fSafetyQueenCheck:           {Name: "SafetyQueenCheck"},
	fKingQueenTropism:           {Name: "KingQueenTropism"},
	fThreatMinorAttackedByPawn:  {Name: "ThreatMinorAttackedByPawn"},
	fThreatMinorAttackedByMinor: {Name: "ThreatMinorAttackedByMinor"},
	fThreatMinorAttackedByMajor: {Name: "ThreatMinorAttackedByMajor"},
	fThreatRookAttackedByLesser: {Name: "ThreatRookAttackedByLesser"},
	fThreatMinorAttackedByKing:  {Name: "ThreatMinorAttackedByKing"},
	fThreatRookAttackedByKing:   {Name: "ThreatRookAttackedByKing"},
	fThreatQueenAttackedByOne:   {Name: "ThreatQueenAttackedByOne"},
	fThreatWeakPawn:             {Name: "ThreatWeakPawn"},
	fKnightOutpost:              {Name: "KnightOutpost"},
	fMinorProtected:             {Name: "MinorProtected"},
	fMinorBehindPawn:            {Name: "MinorBehindPawn"},
	fBishopRammedPawns:          {Name: "BishopRammedPawns"},
	fRookOpen:                   {Name: "RookOpen"},
	fRookSemiopen:               {Name: "RookSemiopen"},
	fTempo:                      {Name: "Tempo"},
}

var totalFeatureSize int

func init() {
	var startIndex = 0
	for i := range infos {
		if infos[i].Name == "" {
			continue
		}
		if infos[i].Size == 0 {
			infos[i].Size = 1
		}
		infos[i].StartIndex = startIndex
		startIndex += infos[i].Size
	}
	totalFeatureSize = startIndex
}
