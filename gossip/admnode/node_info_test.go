package admnode

import (
	"net"
	"testing"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, Sign(&nodeInfo, privKey, TypeURLV1))
}
