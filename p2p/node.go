package p2p

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

const defaultListenPort = 6969

// just a place holder for struct right now
type nodes string

// sample Block
type Block struct{}

type Node struct {
	BootStrapNodes []nodes
	Port           string
	Addr           *net.UDPAddr
	Dailer         *net.Conn
	KnownPeers     map[string]PeerNode

	DBPath string
}

// New creates a new peer and returns it
func New() Node {
	var n Node
	return n
}

// Send sends a message to specific host
func (n *Node) Send(msg []byte, dest net.Addr) {
	fmt.Fprintf(*n.Dailer, string(msg))
}
func (n *Node) Status(b *Block, dest net.Addr) {
	msg := bytes.Buffer{}
	enc := gob.NewEncoder(&msg)

	if err := enc.Encode(b); err != nil {
		// err
	}

	n.Send(msg.Bytes(), dest)
}

func (n *Node) Listen() {

	fmt.Println("Listening on a port for new peers", n.Port)

	// load its db state somehow

	ctx, err := context.WithCancel(context.Background())
	if err != nil {
		// this is a fatal error my guy
	}

	go n.sync(ctx)

	listener, err := net.ListenUDP("UDP", n.Addr)
	if err != nil {
		// ...
	}

	msg := []byte{}
	for {
		listener.Read(msg)
	}
}

// it finds other peers
func (n *Node) Sync(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			// searching for new peers
			n.FetchNewBlockAndPeers()
		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) FetchNewBlockAndPeers() {
	for _, peer := range n.KnownPeers {
		status, err := GetPeerInfo(peer)
		if err != nil {
			// Could not connect with peer
		}
		// else update local block stuff

		/*updating block, consensus algorithm code here*/

		// checking if any new peer was found
		for peer, _ := range status.KnownPeers {
			p, ok := n.KnownPeers[peer.GetTCPAddress()]
			if !ok {
				// new peer, updating peer list
				n.KnownPeers[peer.GetTCPAddress()] = p.PeerNode
			}
		}

	}

}
