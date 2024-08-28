package domain

type DatasetItem struct {
	Fen    string
	Target float64
}

type TuneEntry struct {
	Features         []FeatureInfo
	MgPhase          float32
	EgPhase          float32
	WhiteStrongScale float32
	BlackStrongScale float32
}

type FeatureInfo struct {
	Index int16
	Value int16
}
