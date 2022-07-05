package admnode

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/crypto"
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
	math.ReadBits(key.X, pub[:32])
	math.ReadBits(key.Y, pub[32:])
	copy(nodeId[:], crypto.Keccak256(pub))

	return nodeId
}
