package admnode

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"strings"

	"github.com/adamnite/go-adamnite/common/math"
	"golang.org/x/crypto/sha3"
)

type NodeID [32]byte

// String prints as a hexadecimal number
func (id NodeID) String() string {
	return fmt.Sprintf("%x", id[:])
}

// Bytes returns byte array
func (id NodeID) Bytes() []byte {
	return id[:]
}

// StringToHexID converts a string to NodeID
func StringToHexID(strID string) NodeID {
	id, err := ParseNodeID(strID)
	if err != nil {
		panic(err)
	}
	return id
}

func ParseNodeID(strID string) (NodeID, error) {
	var id NodeID
	byteID, err := hex.DecodeString(strings.TrimPrefix(strID, "0x"))
	if err != nil {
		return id, err
	} else if len(byteID) != len(id) {
		return id, errors.New("wrong length of node id string")
	}

	copy(id[:], byteID)
	return id, nil
}

func PubkeyToNodeID(key *ecdsa.PublicKey) NodeID {
	var nodeId NodeID
	pub := make([]byte, 64)
	// pub := [64]byte{}
	math.ReadBits(key.X, pub[:32])
	math.ReadBits(key.Y, pub[32:])
	pubCopy := append([]byte{}, pub...)
	hashed := sha3.Sum512(pubCopy)
	copy(nodeId[:], hashed[:32])

	return nodeId
}

// LogDist returns the logarithmic distance between a and b
func LogDist(a, b NodeID) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += bits.LeadingZeros8(x)
			break
		}
	}
	return len(a) - lz
}

// DistanceCmp compares the distances
// -1: a is closer to target
// 0: equal
// 1: b is closer to target
func DistanceCmp(a, b, target NodeID) int {
	for i := range target {
		distanceWithA := a[i] ^ target[i]
		distanceWithB := b[i] ^ target[i]

		if distanceWithA > distanceWithB {
			return 1
		} else if distanceWithA < distanceWithB {
			return -1
		}
	}
	return 0
}
