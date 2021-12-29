package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

const defaultListenPort = 6969

// just a place holder for struct right now
type nodes string

// sample Block
type Block struct{}

// 0 -> peer-sync-request, give me ur peer list please
// 1 -> peer-sync-response
// -1 -> close connection
type Msg struct {
	MsgType    int                 `json:"type"` // type of messge, 0 -> peer-sync-request, give me ur peer list please
	KnownPeers map[string]PeerNode `json:"known-peers"`
	BlockMsg   Block               `json:"block"`
}
type Node struct {
	BootStrapNodes []string
	Port           string
	Addr           string
	Dailer         *net.Conn
	KnownPeers     map[string]PeerNode

	DBPath string
}

// New creates a new peer and returns it
func New() Node {
	var n Node
	n.KnownPeers = map[string]PeerNode{}
	// loading peer list from config file
	n.LoadKnownPeers()
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

// Listen handles any incommoing request
func (n *Node) Listen() {
	listener, err := net.Listen("tcp", n.Addr)
	if err != nil {
		log.Fatalf("Could not spawn message request handler due to error %v\n", err)
	}

	for {
		conn, _ := listener.Accept()
		go n.HandleMessageRequest(conn)
	}
}

// Requesting its peers to send peer list data
func (n *Node) SyncPeerList() {
	for _, v := range n.BootStrapNodes {
		peerList := n.GetPeerList(v)
		for k, v := range peerList {
			if _, ok := n.KnownPeers[k]; !ok {
				n.KnownPeers[k] = v
			}
		}
	}
}

// GetPeerList communiates with the node, passed as arguement and gets peer list
func (n *Node) GetPeerList(v string) map[string]PeerNode {
	var err error
	peerList := map[string]PeerNode{}

	fmt.Println("trying to connect with", v)

	conn, err := net.Dial("tcp", v)
	if err != nil {
		log.Println("Could not connect due to error", err)
		return peerList
	}
	defer conn.Close()

	peerList = n.SendPeerSyncRequest(&conn)

	return peerList
}

/* it finds other peers
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
}*/

/*
func (n *Node) FetchNewBlockAndPeers() {
	for _, peer := range n.KnownPeers {
		status, err := GetPeerInfo(peer)
		if err != nil {
			// Could not connect with peer
		}
		// else update local block stuff

		//updating block, consensus algorithm code here

		// checking if any new peer was found
		for peer, _ := range status.KnownPeers {
			p, ok := n.KnownPeers[peer.GetTCPAddress()]
			if !ok {
				// new peer, updating peer list
				n.KnownPeers[peer.GetTCPAddress()] = p.PeerNode
			}
		}

	}

}*/
