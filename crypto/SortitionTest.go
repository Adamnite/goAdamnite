//This is essentially pseudocode; Tsiamfei, TrueThinker, and LastAstronaut, please work to implement this into the actual DPOS Algorithm.
//Tsiamfei and I developed these ideas together.


package vrf

import (
  "crypto/ed25519" //Please try to implement this with SHA-512 using secp256k1
  "encoding/hex"
  "log"
  "testing"
  "math/rand"
)


const Genesis_Seed = "o20c468db4a665w3d53db9f0e9f08155a8052cabdddc8326a2c3bd2d90e42fea"


//Test the Sorition Mechanism with a weighted average
func TestSoritionMechanism(t *testing.T){
    test_number := 100
    var success := 0
    for i := 0; i < test_number; i++ {
    //Implement a random salt that appends itself to Genesis Seed everytime a new round is calculated
      const message = Genesis_Seed
      public, private := VrfKeygen()
      const testVotesFor = random(10000)
      const testPercentBlocksCorrect = random(100)
      const stake = random(20000)
      const nonce = random(1000)
      const weighted_average = (testVotesFor * 0.25 + testPercentBlocksCorrect * 0.55 + stake * .1 + validator_nonce * 0.25)
      _, Beta, err := Prove(private,message)
      Output, err := Verify(public,message)
      if check(Output,weighted_average){
        success++
    }
  }
   return success
}
