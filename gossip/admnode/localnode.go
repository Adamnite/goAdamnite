package admnode

import (
	"crypto/ecdsa"
	"errors"
	"net"
	"sync"
)

type LocalNode struct {
	id          NodeID
	key         *ecdsa.PrivateKey
	db          *NodeDB
	curNodeInfo *NodeInfo

	mu sync.Mutex
}

func NewLocalNode(db *NodeDB, key *ecdsa.PrivateKey) *LocalNode {
	localNode := &LocalNode{
		id:          PubkeyToNodeID(&key.PublicKey),
		db:          db,
		key:         key,
		curNodeInfo: &NodeInfo{},
	}

	return localNode
}

func (n *LocalNode) NodeInfo() *NodeInfo {
	return n.curNodeInfo
}

func (n *LocalNode) SetVersion(version NodeInfoVersion) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.curNodeInfo.SetVersion(version)
	n.doSign(TypeURLV1)
}

func (n *LocalNode) SetTCP(port uint16) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.curNodeInfo.SetTCP(port)
	n.doSign(TypeURLV1)
}

func (n *LocalNode) SetUDP(port uint16) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.curNodeInfo.SetUDP(port)
	n.doSign(TypeURLV1)
}

func (n *LocalNode) SetIP(ip net.IP) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.curNodeInfo.SetIP(ip)
	n.doSign(TypeURLV1)
}

func (n *LocalNode) GetVersion() NodeInfoVersion {
	return n.curNodeInfo.GetVersion()
}

func (n *LocalNode) GetNetType() NetType {
	return n.curNodeInfo.GetNetType()
}

func (n *LocalNode) GetTCP() uint16 {
	return n.curNodeInfo.GetTCP()
}

func (n *LocalNode) GetUDP() uint16 {
	return n.curNodeInfo.GetUDP()
}

func (n *LocalNode) GetIP() net.IP {
	return n.curNodeInfo.GetIP()
}

func (n *LocalNode) GetInfoType() NodeInfoType {
	return n.curNodeInfo.GetType()
}

func (n *LocalNode) GetPubKey() *ecdsa.PublicKey {
	return n.curNodeInfo.GetPubKey()
}

func (n *LocalNode) doSign(infoType NodeInfoType) error {
	if n.key == nil {
		return errors.New("cannot find private key")
	}

	return Sign(n.curNodeInfo, n.key, infoType)
}
