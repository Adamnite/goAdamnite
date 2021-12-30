package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type PeerInfo struct {
	KnownPeers     map[string]PeerNode `json:"known-peers"`
	BootStrapNodes []string            `json:"boot-strap-nodes"`
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
	if len(c.KnownPeers) != 0 {
		p.KnownPeers = c.KnownPeers
	}
	p.BootStrapNodes = c.BootStrapNodes

}
