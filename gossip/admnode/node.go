package admnode

import (
	"crypto/ecdsa"
	"errors"
	"net"

	"github.com/vmihailenco/msgpack/v5"
)

// GossipNode represents a host on the Adamnite network.
type GossipNode struct {
	id   *NodeID
	info *NodeInfo
}

// New wraps a node information.
func New(nodeInfo *NodeInfo) (*GossipNode, error) {
	if err := nodeInfo.Verify(); err != nil {
		return nil, err
	}

	node := &GossipNode{info: nodeInfo}

	node.id = nodeInfo.GetNodeID()
	if len(node.id) != len(NodeID{}) {
		return nil, errors.New("invalid node id")
	}

	return node, nil
}

func NewWithParams(pubKey *ecdsa.PublicKey, ip net.IP, tcp uint16, udp uint16) *GossipNode {
	var nodeInfo NodeInfo

	nodeInfo.version = NodeInfoVersionV1

	if len(ip) > 0 {
		nodeInfo.SetIP(ip)
	}
	if udp != 0 {
		nodeInfo.SetUDP(udp)
	}
	if tcp != 0 {
		nodeInfo.SetTCP(tcp)
	}
	if err := Sign(&nodeInfo, nil, pubKey, TypeCompatURLV1); err != nil {
		panic(err)
	}

	node, err := New(&nodeInfo)
	if err != nil {
		panic(err)
	}
	return node
}

func (n *GossipNode) ID() *NodeID {
	return n.id
}

func (n *GossipNode) Version() NodeInfoVersion {
	return n.info.GetVersion()
}

func (n *GossipNode) TCP() uint16 {
	return n.info.GetTCP()
}

func (n *GossipNode) UDP() uint16 {
	return n.info.GetUDP()
}

func (n *GossipNode) IP() net.IP {
	return n.info.GetIP()
}

func (n *GossipNode) Pubkey() *ecdsa.PublicKey {
	return n.info.GetPubKey()
}

func (n *GossipNode) IsValidate() error {
	if n.IP() == nil {
		return errors.New("missing IP address")
	}
	if n.UDP() == 0 {
		return errors.New("missing UDP port")
	}

	if n.IP().IsMulticast() || n.IP().IsUnspecified() {
		return errors.New("invalid IP")
	}

	if len(n.info.pubKey) == 0 {
		return errors.New("invalid pubkey")
	}
	return nil
}

var _ msgpack.CustomEncoder = (*GossipNode)(nil)
var _ msgpack.CustomDecoder = (*GossipNode)(nil)

func (n *GossipNode) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(n.id, n.info)
}

func (n *GossipNode) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&n.id, &n.info)
}

var _ msgpack.Marshaler = (*NodeID)(nil)
var _ msgpack.Unmarshaler = (*NodeID)(nil)

func (n *NodeID) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(n[:])
}

func (n *NodeID) UnmarshalMsgpack(b []byte) error {
	var id []byte
	if err := msgpack.Unmarshal(b, &id); err != nil {
		return err
	}
	copy(n[:], id)
	return nil
}
