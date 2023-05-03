package admpacket

import (
	"testing"
	"net"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/stretchr/testify/require"
)
var(
	privKey, _ = crypto.HexToECDSA("a8438e283ab8987645c090bd890e098f0980e098da0980b0980789685765e765")
	pubKey     = &privKey.PublicKey

	privKey1, _ = crypto.HexToECDSA("58438e283ab8987645c090bd890e098f0980e098da0980b0980789685765e765")
	pubKey1     = &privKey1.PublicKey
)
func TestPingPongSerialize(t *testing.T) {
	from := PeerEndpoint{
		IP: net.IP{192, 168, 112, 109},
		UDP: 	3009,
		TCP:		3009,
	}
	p := Ping{
		Version: 1,
		PubKey:      pubKey,
		From: 		from,
		To:         from,
		ReqID: 		[]byte("AAAAAAAAAA"),
	}
	byPacket, err := msgpack.Marshal(p)
	if err != nil {
		t.Log(err)
	}
	var ping Ping
	if p.MessageType() == PingMsg {
		

		if err := msgpack.Unmarshal(byPacket, &ping); err != nil {
			t.Log(err)
		}
	}

	require.Equal(t, p.PubKey, ping.PubKey)
}