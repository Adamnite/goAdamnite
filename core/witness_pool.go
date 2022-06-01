package core

import "github.com/adamnite/go-adamnite/core/types"

type WitnessConfig struct {
	WitnessCount uint32 // The total numbers of witness on top tier
}

var DefaultWitnessConfig = WitnessConfig{
	WitnessCount: 27,
}

type WitnessCandidatePool struct {
	config            WitnessConfig
	witnessCandidates []types.Witness
}

func (wp *WitnessCandidatePool) GetCandidates() []types.Witness {
	return ChooseWitnesses(*wp)
}
