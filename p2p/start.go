package p2p

import (
	"log"
	"os"
	"time"
)

// Start is the entry point for the main function, depending on the mode
// it calls the necessary function
func (n *Node) Start() {

	// if no arguement was supplied, or first arguement was --client-mode run as client mode
	if len(os.Args) < 2 || os.Args[1] == "--client-node" {
		n.Mode = "client"
		log.Println("[+] Started in client mode")
		for {
			n.SyncPeerList()
			log.Println("After syncing", n.Peers)
			time.Sleep(10 * time.Second)
		}

	} else if os.Args[1] == "--full-node" {
		n.Mode = "full-node"
		n.Listen()
	}
}
