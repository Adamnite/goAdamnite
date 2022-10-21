package core

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/params"
)

func TestVRF(t *testing.T) {
	witnessAddrs := []common.Address{
		common.HexToAddress("0x5d8124bb42734acb442b6992c73ecad2651612cd"),
		common.HexToAddress("0x5117dd7283175dfd686757784de62197bd2179a2"),
	}
	witnessPriv := []crypto.PrivateKey{
		[]byte{129, 50, 8, 58, 48, 196, 228, 240, 112, 217, 229, 206, 85, 67, 205, 15, 40, 217, 153, 238, 218, 79, 48, 169, 112, 187, 33, 209, 251, 65, 117, 60, 0, 30, 169, 254, 159, 106, 159, 223, 76, 124, 13, 161, 82, 78, 3, 97, 148, 179, 162, 197, 157, 173, 109, 98, 79, 111, 163, 101, 76, 8, 175, 235},
		[]byte{108, 99, 17, 93, 248, 20, 105, 122, 148, 157, 117, 109, 237, 91, 23, 248, 127, 185, 48, 174, 33, 224, 58, 151, 24, 212, 141, 196, 67, 117, 212, 66, 183, 147, 88, 51, 65, 77, 98, 23, 11, 5, 212, 7, 117, 9, 79, 87, 205, 213, 94, 63, 138, 127, 18, 226, 23, 33, 19, 17, 84, 228, 13, 226},
	}
	witnessPub := []crypto.PublicKey{}
	for _, x := range witnessPriv {
		foo, _ := x.Public()
		witnessPub = append(witnessPub, foo)
	}

	witnessCans := []types.Witness{
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
			PubKey: witnessPub[0],
		},
		&types.WitnessImpl{
			Address: witnessAddrs[1],
			Voters: []types.Voter{
				types.Voter{
					Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217902"),
					StakingAmount: big.NewInt(1000000000000000000),
				},
				types.Voter{
					Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217903"),
					StakingAmount: big.NewInt(9000000000000000000),
				},
				// types.Voter{
				// 	Address:       common.HexToAddress("0x5117dd7283175dfd686757784de62197bd217905"),
				// 	StakingAmount: big.NewInt(9000000000000000000),
				// },
			},
			PubKey: witnessPub[1],
		},
	}
	testSeed := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	witnessCandidatePool := NewWitnessPool(DefaultWitnessConfig, params.DemoChainConfig, nil, witnessCans, testSeed)

	vrf, proof := witnessPriv[0].Prove(testSeed)
	fmt.Println(witnessCandidatePool.IsTrustedWitness(witnessPub[0], vrf, proof))

	vrf, proof = witnessPriv[1].Prove(testSeed)
	fmt.Println(witnessCandidatePool.IsTrustedWitness(witnessPub[1], vrf, proof))
	fmt.Println(witnessCandidatePool.selectedWitnesses)
	witnesses := witnessCandidatePool.witnessCandidates
	t.Log(witnesses)
}
