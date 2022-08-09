package tests

import (
	"net"
	"testing"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	privKey, _ = crypto.HexToECDSA("a8438e283ab8987645c090bd890e098f0980e098da0980b0980789685765e765")
	pubKey     = &privKey.PublicKey
)

func checkGossipNodeMsgPack(t *testing.T, node *admnode.GossipNode) {
	b, err := msgpack.Marshal(node)
	if err != nil {
		t.Error(err)
	}

	var decNode admnode.GossipNode
	err = msgpack.Unmarshal(b, &decNode)
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, decNode.ID(), node.ID())
	require.Equal(t, decNode.TCP(), node.TCP())
	require.Equal(t, decNode.UDP(), node.UDP())
	require.Equal(t, decNode.Pubkey(), node.Pubkey())
	require.Equal(t, decNode.Version(), node.Version())
	require.Equal(t, decNode.IP(), node.IP())
}

func TestGossipNodeMsgPackEncoding(t *testing.T) {
	node := admnode.NewWithParams(pubKey, net.IPv4(10, 10, 10, 10), uint16(90), uint16(90))
	node1 := admnode.NewWithParams(pubKey, net.IPv6loopback, uint16(90), uint16(90))
	checkGossipNodeMsgPack(t, node)
	checkGossipNodeMsgPack(t, node1)
}
