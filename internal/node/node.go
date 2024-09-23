package node

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"

	"github.com/adamnite/go-adamnite/internal/bargossip"
	"github.com/adamnite/go-adamnite/internal/blockchain"
	"github.com/adamnite/go-adamnite/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

// Node is a container on which services can be registered.
type Node struct {
	// config describes the configuration of the Adamnite node.
	config Config
	prvKey crypto.PrivKey

	bootstrapNodes []host.Host

	localNode   bargossip.LocalNode
	remoteNodes []bargossip.RemoteNode
	adamnite    blockchain.Blockchain

	// Channels
	stop     chan struct{} // Channel to wait for termination notifications
	onceStop sync.Once
	pingDone chan bool
}

func New(config *Config) (*Node, error) {
	var r io.Reader

	r = rand.Reader

	// check if the prvkey exists on the datadir

	prvKey := checkNodeKey(config)

	if prvKey == nil {
		var err error

		prvKey, _, err = crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
		if err != nil {
			log.Error("Key generation failed")
			return nil, err
		}

		saveNodeKey(config, prvKey)
	}

	// save the prvkey

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2PPort))

	server, err := libp2p.New(libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(prvKey))

	if err != nil {
		log.Error("Bargossip server hosting error!")
	}

	localNode := bargossip.CreateLocalNode(server, bargossip.Config{
		BootstrapNodes: config.BootstrapNodes,
		ProtocolID:     config.ProtocolID,
		NodeType:       config.NodeType,
	})

	adamnite := blockchain.New(config.AdamniteConfig, localNode, config.DataDir, config.ValidatorAddr)

	node := &Node{
		config:    *config,
		prvKey:    prvKey,
		stop:      make(chan struct{}),
		pingDone:  make(chan bool),
		localNode: localNode,
		adamnite:  adamnite,
	}

	return node, nil
}

func (n *Node) Wait() {
	<-n.stop
}

func (n *Node) Start() {
	log.Info("Adamnite DataDir path", "PATH", n.config.DataDir)
	log.Info("Adamnite external address", "Addr", n.localNode.GetServer().Addrs())
	log.Info("Adamnite host ID", "ID", n.localNode.GetServer().ID())

	n.localNode.Start()
	n.adamnite.Start()
}

func (n *Node) Close() {
	closeOnce := func() {
		close(n.stop)
	}

	n.onceStop.Do(closeOnce)
}
