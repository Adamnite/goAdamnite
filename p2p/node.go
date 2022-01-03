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
	MsgType  int        `json:"type"` // type of messge, 0 -> peer-sync-request, give me ur peer list please
	Peers    []PeerNode `json:"peers"`
	BlockMsg Block      `json:"block"`
}
type Node struct {
	BootStrapNodes []string
	Port           string
	Addr           string
	Dailer         *net.Conn
	Peers          []PeerNode
	Mode           string
	DBPath         string
}

// New creates a new peer and returns it
func New() Node {
	var n Node
	n.Peers = []PeerNode{}
	n.Addr = "0.0.0.0:6969"
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
	for _, v := range n.Peers {
		peerList := n.GetPeerList(v.IP + ":" + v.Port)
		for _, v := range peerList {
			if !n.checkIfPeerExists(v.IP) {
				n.Peers = append(n.Peers, v)
			}

		}
	}
}

// GetPeerList communiates with the node, passed as arguement and gets peer list
func (n *Node) GetPeerList(v string) []PeerNode {
	var err error
	peerList := []PeerNode{}

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

func (n *Node) checkIfPeerExists(peer string) bool {
	for _, v := range n.Peers {
		if v.IP == peer {
			return true
		}
	}
	return false
}
