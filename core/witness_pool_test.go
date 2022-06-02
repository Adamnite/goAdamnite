package core

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
)

func TestVRF(t *testing.T) {
	witnessAddrs := []common.Address{
		common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
		common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
	}
	witnessCandidatePool := WitnessCandidatePool{
		config: DefaultWitnessConfig,
		witnessCandidates: []types.Witness{
			&types.WitnessImpl{
				Address: witnessAddrs[0],
				Voters: []types.Voter{
					types.Voter{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217901"),
						StakingAmount: big.NewInt(1000000000000000000),
					},
					types.Voter{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217902"),
						StakingAmount: big.NewInt(9000000000000000000),
					},
				},
			},
			&types.WitnessImpl{
				Address: witnessAddrs[1],
				Voters: []types.Voter{
					types.Voter{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217902"),
						StakingAmount: big.NewInt(7800000000000000000),
					},
					types.Voter{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217903"),
						StakingAmount: big.NewInt(9000000000000000000),
					},
					types.Voter{
						Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217905"),
						StakingAmount: big.NewInt(9000000000000000000),
					},
				},
			},
		},
	}

	witnesses := witnessCandidatePool.GetCandidates()
	t.Log(witnesses)
}
