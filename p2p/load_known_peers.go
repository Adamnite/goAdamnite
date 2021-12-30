package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type PeerInfo struct {
	Peers []PeerNode `json:"peers"`
}

func (p *Node) LoadKnownPeers() {
	content, err := os.ReadFile("./known-peers.json")
	if err != nil {
		log.Fatalf("Could not open known-peers.json due to error %v\n", err)
	}

	var c PeerInfo

	if err = json.Unmarshal(content, &c); err != nil {
		log.Fatalf("Could not unmarshall due to error %v\n", err)
	}
	fmt.Println("from config file", c)
	// we do not want nil assignments
	if len(c.Peers) != 0 {
		p.Peers = c.Peers
	}
	p.Peers = c.Peers

}
