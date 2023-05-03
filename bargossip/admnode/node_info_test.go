package admnode

import (
	"net"
	"testing"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	privKey, _ = crypto.HexToECDSA("a8438e283ab8987645c090bd890e098f0980e098da0980b0980789685765e765")
	pubKey     = &privKey.PublicKey
)

func TestSignatureAndVerify(t *testing.T) {
	var nodeInfo NodeInfo

	nodeInfo.SetVersion(NodeInfoVersionV0)
	nodeInfo.SetTCP(30)
	nodeInfo.SetUDP(20)
	nodeInfo.SetIP(net.IPv4(10, 10, 10, 10))

	require.NoError(t, Sign(&nodeInfo, privKey, nil, TypeURLV1))
}

func TestSerialize(t *testing.T) {
	node := NewWithParams(pubKey, net.IPv4(192, 168, 109, 100), 90, 90)

	val, err := msgpack.Marshal(node.NodeInfo())
	if err != nil {
		t.Error("Failed to encode node info")
	}

	var nodeInfo NodeInfo

	if err := msgpack.Unmarshal(val, &nodeInfo); err != nil {
		t.Error("Failed to decode node info")
	}

	require.Equal(t, node.Pubkey(), nodeInfo.GetPubKey())
}
