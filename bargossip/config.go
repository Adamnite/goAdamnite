package bargossip

import (
	"crypto/ecdsa"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
)

const (
	defaultMaxInboundConnections = 15
	defaultMaxPendingConnections = 15

	inboundAtemptDuration = 60 * time.Second
)

type Config struct {
	// ServerPrvKey is the key to generate the PeerID
	ServerPrvKey *ecdsa.PrivateKey

	// ListenAddr is the listening address, incomming connection from other nodes
	ListenAddr string

	// MaxPendingConnections is the maximum number of pending incomming peer connections
	MaxPendingConnections int

	// MaxInboundConnection is the maximum number of incomming peer connections
	MaxInboundConnections int

	// MaxOutboundConnection is the maximum number of outbounding peer connections
	MaxOutboundConnections int

	// PeerBlackList is the black list of the peers
	PeerBlackList *utils.IPNetList

	// PeerWhiteList is the white list of the peers
	PeerWhiteList *utils.IPNetList

	// BootstrapNodes are used to establish connectivity with the rest of the network
	BootstrapNodes []*admnode.GossipNode

	// NodeDatabase is the path to the database containing the previously seen live nodes
	NodeDatabase string

	// NAT is used to make the listening port available to the Internet
	NAT nat.Interface

	// ChainProtocol is used to communicate with peers about the blockchain information
	ChainProtocol []SubProtocol

	// Logger is a logger to use with gossip server
	Logger log15.Logger

	// clock will be used on gossip server
	clock mclock.Clock
}
