package admnode

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/rlp"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/sha3"
)

type NetType uint8

const (
	NetTypeIPV4 NetType = 0x01
	NetTypeIPV6 NetType = 0x02
)

type NodeInfoVersion uint64

const (
	NodeInfoVersionV0 NodeInfoVersion = 0x0
	NodeInfoVersionV1 NodeInfoVersion = 0x1
)

type NodeInfoType string

const (
	TypeNil         NodeInfoType = "nil"
	TypeURLV1       NodeInfoType = "urlv1"
	TypeCompatURLV1 NodeInfoType = "compaturlv1"
)

// NodeInfo represents a node information.
type NodeInfo struct {
	version   NodeInfoVersion
	netType   NetType
	tcp       uint16
	udp       uint16
	ip        net.IP
	infoType  NodeInfoType
	pubKey    []byte
	signature []byte `msgpack:",omitempty"`
}

func (n *NodeInfo) SetVersion(version NodeInfoVersion) {
	n.invalidate()
	n.version = version
}

func (n *NodeInfo) SetTCP(port uint16) {
	n.invalidate()
	n.tcp = port
}

func (n *NodeInfo) SetUDP(port uint16) {
	n.invalidate()
	n.udp = port
}

func (n *NodeInfo) SetIP(ip net.IP) {
	n.invalidate()

	if ipV4 := ip.To4(); ipV4 != nil {
		n.ip = ip.To4()
		n.netType = NetTypeIPV4
	} else {
		n.ip = ip.To16()
		n.netType = NetTypeIPV6
	}
}

func (n *NodeInfo) GetVersion() NodeInfoVersion {
	return n.version
}

func (n *NodeInfo) GetNetType() NetType {
	return n.netType
}

func (n *NodeInfo) GetTCP() uint16 {
	return n.tcp
}

func (n *NodeInfo) GetUDP() uint16 {
	return n.udp
}

func (n *NodeInfo) GetIP() net.IP {
	return n.ip
}

func (n *NodeInfo) GetType() NodeInfoType {
	return n.infoType
}

func (n *NodeInfo) GetPubKey() *ecdsa.PublicKey {
	pubKey, err := crypto.DecompressPubkey(n.pubKey)
	if err != nil {
		return nil
	}

	return pubKey
}

// Signature returns the signature of the node info
func (n *NodeInfo) Signature() []byte {
	if n.signature == nil {
		return nil
	}

	sig := make([]byte, len(n.signature))
	copy(sig, n.signature)
	return sig
}

func (n *NodeInfo) invalidate() {
	n.signature = nil
}

func (n *NodeInfo) ToRawElements() []interface{} {
	var list []interface{}
	list = append(list, n.version)

	switch n.version {
	case NodeInfoVersionV0:
		list = append(list, n.netType)
		list = append(list, n.tcp)
		list = append(list, n.udp)
		list = append(list, n.ip)
		list = append(list, n.infoType)
		list = append(list, n.pubKey)
	default:
	}
	return list
}

func Sign(n *NodeInfo, privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey, infoType NodeInfoType) error {
	cpy := *n

	cpy.infoType = infoType
	switch infoType {
	case TypeNil:
	case TypeURLV1:
		cpy.invalidate()
		cpy.pubKey = crypto.CompressPubkey(&privKey.PublicKey)

		h := sha3.NewLegacyKeccak256()
		rlp.Encode(h, cpy.ToRawElements())

		sig, err := crypto.Sign(h.Sum(nil), privKey)
		if err != nil {
			return err
		}

		sig = sig[:len(sig)-1]

		if err := cpy.verify(sig); err != nil {
			return err
		}
		cpy.signature = sig
		*n = cpy
	case TypeCompatURLV1:
		cpy.invalidate()
		cpy.pubKey = crypto.CompressPubkey(pubKey)

		if err := cpy.verify([]byte{}); err != nil {
			return err
		}
		cpy.signature = []byte{}
		*n = cpy
	}

	return nil
}

func (n *NodeInfo) Verify() error {
	if n.signature == nil {
		return errors.New("signature field must be not null")
	}
	return n.verify(n.signature)
}

func (n *NodeInfo) verify(sig []byte) error {
	switch n.infoType {
	case TypeNil:
	case TypeURLV1:
		h := sha3.NewLegacyKeccak256()
		rlp.Encode(h, n.ToRawElements())
		if !crypto.VerifySignature(n.pubKey, h.Sum(nil), sig) {
			return ErrInvalidSig
		}
	case TypeCompatURLV1:
		_, err := crypto.DecompressPubkey(n.pubKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeInfo) GetNodeID() *NodeID {
	nodeId := NodeID{}
	switch n.infoType {
	case TypeNil:
	case TypeURLV1:
		pubKey := n.GetPubKey()
		buf := make([]byte, 64)
		math.ReadBits(pubKey.X, buf[:32])
		math.ReadBits(pubKey.Y, buf[32:])
		copy(nodeId[:], crypto.Keccak256(buf))
	case TypeCompatURLV1:
		pubKey := n.GetPubKey()
		buf := make([]byte, 64)
		math.ReadBits(pubKey.X, buf[:32])
		math.ReadBits(pubKey.Y, buf[32:])
		copy(nodeId[:], crypto.Keccak256(buf))
	}

	return &nodeId
}

func (n *NodeInfo) ToURL() string {
	nodeid := fmt.Sprintf("%x", crypto.FromECDSAPub(n.GetPubKey())[1:])

	nodeUrl := url.URL{Scheme: "gnite"}

	if n.ip != nil {
		addr := net.TCPAddr{IP: n.GetIP(), Port: int(n.GetTCP())}
		nodeUrl.User = url.User(nodeid)
		nodeUrl.Host = addr.String()

		if n.GetUDP() != n.GetTCP() {
			nodeUrl.RawQuery = "udp=" + strconv.Itoa(int(n.GetUDP()))
		}
	}

	return nodeUrl.String()
}

func ParseNodeURL(input string) (*NodeInfo, error) {
	return nil, nil
}

// ***********************************************************************************************************
// **********************************************  Serialization *********************************************
// ***********************************************************************************************************

var _ msgpack.CustomEncoder = (*NodeInfo)(nil)
var _ msgpack.CustomDecoder = (*NodeInfo)(nil)

func (n *NodeInfo) EncodeMsgpack(enc *msgpack.Encoder) error {
	ip, err := n.ip.MarshalText()
	if err != nil {
		return err
	}

	return enc.EncodeMulti(n.version, n.netType, n.tcp, n.udp, ip, n.infoType, n.pubKey, n.signature)
}

func (n *NodeInfo) DecodeMsgpack(dec *msgpack.Decoder) error {
	var byIP []byte
	err := dec.DecodeMulti(&n.version, &n.netType, &n.tcp, &n.udp, &byIP, &n.infoType, &n.pubKey, &n.signature)
	if err != nil {
		return err
	}
	var ip net.IP
	if err = ip.UnmarshalText(byIP); err != nil {
		return err
	}
	if ip.To4() != nil {
		n.ip = ip.To4()
	} else {
		n.ip = ip
	}
	return nil
}

var _ msgpack.Marshaler = (*NodeInfoVersion)(nil)
var _ msgpack.Unmarshaler = (*NodeInfoVersion)(nil)

func (n NodeInfoVersion) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(uint64(n))
}

func (n *NodeInfoVersion) UnmarshalMsgpack(b []byte) error {
	var version uint64
	if err := msgpack.Unmarshal(b, &version); err != nil {
		return err
	}
	*n = NodeInfoVersion(version)
	return nil
}

var _ msgpack.Marshaler = (*NodeInfoType)(nil)
var _ msgpack.Unmarshaler = (*NodeInfoType)(nil)

func (n NodeInfoType) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(string(n))
}

func (n *NodeInfoType) UnmarshalMsgpack(b []byte) error {
	var infoType string

	if err := msgpack.Unmarshal(b, &infoType); err != nil {
		return err
	}

	*n = NodeInfoType(infoType)
	return nil
}

var _ msgpack.Marshaler = (*NetType)(nil)
var _ msgpack.Unmarshaler = (*NetType)(nil)

func (n NetType) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(uint8(n))
}

func (n *NetType) UnmarshalMsgpack(b []byte) error {
	var netType uint8
	if err := msgpack.Unmarshal(b, &netType); err != nil {
		return err
	}
	*n = NetType(netType)
	return nil
}
