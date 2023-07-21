package admpacket

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net"
	"testing"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/utils/mclock"
	"github.com/adamnite/go-adamnite/crypto"
)

var (
	testKeyA, _   = crypto.HexToECDSA("ada0000000000000000000000000000000000000000000000000000000000000")
	testKeyB, _   = crypto.HexToECDSA("ada0000000000000000000000000000000000000000000000000000000000001")
	testEphKey, _ = crypto.HexToECDSA("ada0000000000000000000000000000000000000000000000000000000000002")
	testRandomID  = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
)

// A -> B    SYN
// B -> A    ASKHANDSHAKE
// A -> B 	 FindNode with handshake
// B -> A    Node
func TestHandshake(t *testing.T) {
	t.Parallel()

	demo := newHandshakeTest()
	defer demo.close()

	// A -> B 		SYN
	packet, _ := demo.nodeA.encode(t, demo.nodeB, &Findnode{})
	resp := demo.nodeB.expectDecode(t, SYNMsg, packet)

	// A <- B		ASKHANDSHAKE
	askHandshake := &AskHandshake{
		Nonce:     resp.(*SYN).Nonce,
		RandomID:  testRandomID,
		DposRound: 8,
	}
	enc, _ := demo.nodeB.encode(t, demo.nodeA, askHandshake)
	demo.nodeA.expectDecode(t, AskHandshakeMsg, enc)

	// A -> B		FINDNODE with handshake
	findnode, _ := demo.nodeA.encodeWithChallenge(t, demo.nodeB, askHandshake, &Findnode{})
	demo.nodeB.expectDecode(t, FindnodeMsg, findnode)

	nodes, _ := demo.nodeB.encode(t, demo.nodeA, &RspNodes{Total: 1})
	demo.nodeA.expectDecode(t, RspFindnodeMsg, nodes)
	t.Log(resp)
}

type handshakeTestNode struct {
	ln    *admnode.LocalNode
	codec *SSLCodec
}

type handshakeTest struct {
	nodeA handshakeTestNode
	nodeB handshakeTestNode
	clock mclock.Simulated
}

func newHandshakeTest() *handshakeTest {
	t := new(handshakeTest)
	t.nodeA.init(testKeyA, net.IP{127, 0, 0, 1}, &t.clock)
	t.nodeB.init(testKeyB, net.IP{127, 0, 0, 1}, &t.clock)
	return t
}

func (t *handshakeTest) close() {
	t.nodeA.ln.Database().Close()
	t.nodeB.ln.Database().Close()
}

func (n *handshakeTestNode) init(key *ecdsa.PrivateKey, ip net.IP, clock mclock.Clock) {
	db, _ := admnode.OpenDB("")
	n.ln = admnode.NewLocalNode(db, key)
	n.ln.SetIP(ip)
	n.codec = NewSSLCodec(n.ln, clock, key)
}

func (n *handshakeTestNode) encode(t testing.TB, to handshakeTestNode, p ADMPacket) ([]byte, Nonce) {
	t.Helper()
	return n.encodeWithChallenge(t, to, nil, p)
}

func (n *handshakeTestNode) encodeWithChallenge(t testing.TB, to handshakeTestNode, c *AskHandshake, p ADMPacket) ([]byte, Nonce) {
	t.Helper()

	// Copy challenge and add destination node. This avoids sharing 'c' among the two codecs.
	var challenge *AskHandshake
	if c != nil {
		challengeCopy := *c
		challenge = &challengeCopy
		challenge.Node = to.n()
	}
	// Encode to destination.
	enc, nonce, err := n.codec.Encode(to.id(), to.addr(), p, challenge)
	if err != nil {
		t.Fatal(fmt.Errorf("(%s) %v", n.ln.Node().ID(), err))
	}
	t.Logf("(%s) -> (%s)   %s\n%s", n.ln.Node().ID(), to.id(), p.Name(), hex.Dump(enc))
	return enc, nonce
}

func (n *handshakeTestNode) n() *admnode.GossipNode {
	return n.ln.Node()
}

func (n *handshakeTestNode) addr() string {
	return n.ln.Node().IP().String()
}

func (n *handshakeTestNode) id() admnode.NodeID {
	return *n.ln.Node().ID()
}

func (n *handshakeTestNode) expectDecode(t *testing.T, ptype byte, p []byte) ADMPacket {
	t.Helper()

	dec, err := n.decode(p)
	if err != nil {
		t.Fatal(fmt.Errorf("(%s) %v", n.ln.Node().ID(), err))
	}
	t.Logf("(%s)", n.ln.Node().ID())
	if dec.MessageType() != ptype {
		t.Fatalf("expected packet type %d, got %d", ptype, dec.MessageType())
	}
	return dec
}

func (n *handshakeTestNode) decode(input []byte) (ADMPacket, error) {
	_, _, p, err := n.codec.Decode(input, "127.0.0.1")
	return p, err
}
