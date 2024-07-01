package node

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"

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

	server *host.Host

	prvKey crypto.PrivKey

	// Channels
	stop     chan struct{} // Channel to wait for termination notifications
	onceStop sync.Once
}

func New(config *Config) (*Node, error) {
	var r io.Reader
	r = rand.Reader

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		log.Error("Key generation failed")
		return nil, err
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2PPort))

	server, err := libp2p.New(libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(prvKey))
	if err != nil {
		log.Error("Bargossip server hosting error!")
	}

	return &Node{
		config: *config,
		prvKey: prvKey,
		server: &server,
		stop:   make(chan struct{}),
	}, nil
}

func (n *Node) Wait() {
	<-n.stop
}

func (n *Node) Start() {
	log.Info("Adamnite DataDir path", "PATH", n.config.DataDir)
}

func (n *Node) Close() {
	closeOnce := func() {
		close(n.stop)
	}

	n.onceStop.Do(closeOnce)
}
