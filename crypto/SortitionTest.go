//This is essentially pseudocode; Tsiamfei, TrueThinker, and LastAstronaut, please work to implement this into the actual DPOS Algorithm.
//Tsiamfei and I developed these ideas together.


package vrf

import (
  "crypto/ed25519" //Please try to implement this with SHA-512 using secp256k1
  "encoding/hex"
  "log"
  "testing"
)


const Genesis_Seed = "o20c468db4a665w3d53db9f0e9f08155a8052cabdddc8326a2c3bd2d90e42fea"
