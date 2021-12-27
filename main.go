package main

import (
	"fmt"
	"p2p/p2p"
	"time"
)

func main() {
	// Creating a new p2p instance
	p1 := p2p.New()
	p1.LoadKnownPeers()
	fmt.Println("after loading known peers")
	// temporarily assigning these variables
	p1.Addr = "0.0.0.0:6969"
	go p1.Listen()
	fmt.Println("outside loop the peer list", p1.KnownPeers)
	for {
		p1.SyncPeerList()
		fmt.Println("After syncing", p1.KnownPeers)
		time.Sleep(10 * time.Second)
		fmt.Println("After syncing", p1.KnownPeers)
	}

}
