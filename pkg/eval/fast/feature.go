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
	fPassedPawn
	fPassedEnemyKing
	fPassedOwnKing
	fPawnDuo
	fPawnProtected
	fPawnIsolated
	fPawnDoubled
	fKingShield
	fThreatPawn
	fTempo
	fSize
)

type FeatureInfo struct {
	Name       string
	Size       int
	StartIndex int
}

var infos = [fSize]FeatureInfo{
	fPawnValue:       {Name: "PawnValue"},
	fKnightValue:     {Name: "KnightValue"},
	fBishopValue:     {Name: "BishopValue"},
	fRookValue:       {Name: "RookValue"},
	fQueenValue:      {Name: "QueenValue"},
	fBishopPair:      {Name: "BishopPair"},
	fPawnPST:         {Name: "PawnPST", Size: 32},
	fKnightPST:       {Name: "KnightPST", Size: 32},
	fBishopPST:       {Name: "BishopPST", Size: 32},
	fRookPST:         {Name: "RookPST", Size: 32},
	fQueenPST:        {Name: "QueenPST", Size: 32},
	fKingPST:         {Name: "KingPST", Size: 32},
	fPassedPawn:      {Name: "PassedPawn", Size: 5},
	fPassedEnemyKing: {Name: "PassedEnemyKing", Size: 5 * 8},
	fPassedOwnKing:   {Name: "PassedOwnKing", Size: 5 * 8},
	fPawnDuo:         {Name: "PawnDuo", Size: 32},
	fPawnProtected:   {Name: "PawnProtected", Size: 32},
	fPawnIsolated:    {Name: "PawnIsolated"},
	fPawnDoubled:     {Name: "PawnDoubled"},
	fKingShield:      {Name: "KingShield", Size: 12},
	fThreatPawn:      {Name: "ThreatPawn"},
	fTempo:           {Name: "Tempo"},
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
