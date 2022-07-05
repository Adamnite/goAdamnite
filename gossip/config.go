package gossip

import (
	"crypto/ecdsa"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/gossip/admnode"
	"github.com/adamnite/go-adamnite/gossip/nat"
	"github.com/adamnite/go-adamnite/log15"
)

type Config struct {
	// ServerPrvKey is the key to generate the PeerID
	ServerPrvKey *ecdsa.PrivateKey

	// ListenAddr is the listening address, incomming connection from other nodes
	ListenAddr string

	// MaxInboundConnection is the maximum number of incomming peer connections
	MaxInboundConnection int

	// MaxOutboundConnection is the maximum number of outbounding peer connections
	MaxOutboundConnection int

	// BootstrapNodes are used to establish connectivity with the rest of the network
	BootstrapNodes []*admnode.GossipNode

	// NodeDatabase is the path to the database containing the previously seen live nodes
	NodeDatabase string

	// NAT is used to make the listening port available to the Internet
	NAT nat.Interface

	// Logger is a logger to use with gossip server
	Logger log15.Logger

	// clock will be used on gossip server
	clock mclock.Clock
}
