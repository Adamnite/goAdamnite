package p2p

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	KnownPeers     map[string]PeerNode `json:"known-peers"`
	BootStrapNodes []string            `json:"boot-strap-nodes"`
}

func (p *Node) LoadKnownPeers() {
	content, err := os.ReadFile("./known-peers.json")
	if err != nil {
		log.Fatalf("Could not open known-peers.json due to error %v\n", err)
	}

	var c Config

	if err = json.Unmarshal(content, &c); err != nil {
		log.Fatalf("Could not unmarshall due to error %v\n", err)
	}

	// we do not want nil assignments
	if len(c.KnownPeers) != 0 {
		p.KnownPeers = c.KnownPeers
	}
	p.BootStrapNodes = c.BootStrapNodes

}
