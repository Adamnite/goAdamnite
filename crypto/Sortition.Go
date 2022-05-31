//This is essentially pseudocode; Tsiamfei, TrueThinker, and LastAstronaut, please work to implement this into the actual DPOS Algorithm.
//Tsiamfei and I developed these ideas together.

package vrf

import (
  "log"
  "math/big"
)
const testVotesFor = 4500
const testPercentBlocksCorrect = 90
const stake = 3000
const validator_nonce = 44


const weighted_average = (testVotesFor * 0.25 + testPercentBlocksCorrect * 0.55 + stake * .1 + validator_nonce * 0.25)
func Sortition_determine(ratio float64) bool{
  if ratio > weighted_average{
    return false
  }
  else{
    return true
  }
}


func VRFHash(Beta []byte) float64 {
  t := &big.Int{}
	t.SetBytes(vrfOutput[:])

	precision := uint(8 * (len(vrfOutput) + 1))
	max, b, err := big.ParseFloat("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Fatal("failed to parse big float constant for sortition")
	}

	//hash value as int expression.
	//hval, _ := h.Float64() to get the value
	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	ratio := big.Float{}
	cratio, _ := ratio.Quo(&h, max).Float64()
  return cratio
}
