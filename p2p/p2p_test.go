package p2p

import (
	"testing"
	"time"
)

func TestPeerListUpdate(t *testing.T) {

	knownPeerList := []string{"3.3.3.3", "4.4.4.4"}
	p1 := New()

	p2 := New()
	p2.Peers = []PeerNode{
		PeerNode{IP: "1.1.1.1", Port: "6969"},
		PeerNode{IP: "4.4.4.4", Port: "6969"},
	}
	p2.Addr = "0.0.0.0:6969"
	go p2.Listen()
	time.Sleep(5 * time.Second)
	// sync peer list i.e. sends sends g0et peer request message
	p1.BootStrapNodes = []string{"127.0.0.1:6969"}
	p1.SyncPeerList()

	if len(p1.Peers) == 0 {
		t.Fatal("Known peer list is empty, error")
	}

	for _, v := range knownPeerList {
		if !p2.checkIfPeerExists(v) {
			t.Fatalf("The peer %s is not present, sync failed\n", v)
		}
	}
}
