package eval

type featureInfo struct {
	Name  string
	Index int
	Size  int
}

var features []featureInfo

var featureSize int

func addFeature(size int, name string) int {
	var f = featureInfo{
		Name:  name,
		Index: featureSize,
		Size:  size,
	}
	features = append(features, f)
	featureSize += size
	return f.Index
}

var (
	fPawnValue                  = addFeature(1, "PawnValue")
	fKnightValue                = addFeature(1, "KnightValue")
	fBishopValue                = addFeature(1, "BishopValue")
	fRookValue                  = addFeature(1, "RookValue")
	fQueenValue                 = addFeature(1, "QueenValue")
	fBishopPair                 = addFeature(1, "BishopPair")
	fPawnPST                    = addFeature(32, "PawnPST")
	fKnightPST                  = addFeature(32, "KnightPST")
	fBishopPST                  = addFeature(32, "BishopPST")
	fRookPST                    = addFeature(32, "RookPST")
	fQueenPST                   = addFeature(32, "QueenPST")
	fKingPST                    = addFeature(32, "KingPST")
	fKnightMobility             = addFeature(9, "KnightMobility")
	fBishopMobility             = addFeature(14, "BishopMobility")
	fRookMobility               = addFeature(15, "RookMobility")
	fQueenMobility              = addFeature(28, "QueenMobility")
	fKingShelter                = addFeature(2*8*8, "KingShelter")
	fPawnDuo                    = addFeature(32, "PawnDuo")
	fPawnProtected              = addFeature(32, "PawnProtected")
	fPawnPassed                 = addFeature(4*8, "PawnPassed")
	fPassedFriendlyDistance     = addFeature(5*8, "PassedFriendlyDistance")
	fPassedEnemyDistance        = addFeature(5*8, "PassedEnemyDistance")
	fPawnIsolated               = addFeature(1, "PawnIsolated")
	fPawnDoubled                = addFeature(1, "PawnDoubled")
	fRookOpen                   = addFeature(1, "RookOpen")
	fRookSemiopen               = addFeature(1, "RookSemiopen")
	fBishopRammedPawns          = addFeature(1, "BishopRammedPawns")
	fSafetyWeakSquares          = addFeature(1, "SafetyWeakSquares")
	fSafetySafeQueenCheck       = addFeature(1, "SafetySafeQueenCheck")
	fSafetySafeRookCheck        = addFeature(1, "SafetySafeRookCheck")
	fSafetySafeBishopCheck      = addFeature(1, "SafetySafeBishopCheck")
	fSafetySafeKnightCheck      = addFeature(1, "SafetySafeKnightCheck")
	fKnightOutpost              = addFeature(1, "fKnightOutpost")
	fMinorProtected             = addFeature(1, "MinorProtected")
	fMinorBehindPawn            = addFeature(1, "MinorBehindPawn")
	fThreatMinorAttackedByPawn  = addFeature(1, "ThreatMinorAttackedByPawn")
	fThreatMinorAttackedByMinor = addFeature(1, "ThreatMinorAttackedByMinor")
	fThreatMinorAttackedByMajor = addFeature(1, "ThreatMinorAttackedByMajor")
	fThreatRookAttackedByLesser = addFeature(1, "ThreatRookAttackedByLesser")
	fThreatMinorAttackedByKing  = addFeature(1, "ThreatMinorAttackedByKing")
	fThreatRookAttackedByKing   = addFeature(1, "ThreatRookAttackedByKing")
	fThreatQueenAttackedByOne   = addFeature(1, "ThreatQueenAttackedByOne")
	fThreatWeakPawn             = addFeature(1, "ThreatWeakPawn")
	fTempo                      = addFeature(1, "Tempo")
)
