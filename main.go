package main

import (
	"fmt"
	"os"
	"p2p/p2p"
	"time"
)

func main() {
	// Creating a new p2p instance
	p1 := p2p.New()

	fmt.Println("after loading known peers")
	// temporarily assigning these variables
	p1.Addr = "0.0.0.0:6969"

	// if anyone wants to connect with me
	if len(os.Args) < 2 || os.Args[1] == "--client-node" {
		p1.Mode = "client"
		fmt.Println("outside loop the peer list", p1.KnownPeers)
		for {
			p1.SyncPeerList()
			fmt.Println("After syncing", p1.KnownPeers)
			time.Sleep(10 * time.Second)
			fmt.Println("After syncing", p1.KnownPeers)
		}

	} else if os.Args[1] == "--full-node" {
		p1.Mode = "full-node"
		p1.Listen()
	}

}
