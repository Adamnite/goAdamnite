package admnode

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDBKeys(t *testing.T) {
	nodeID := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	// keyID := getNodeKey(nodeID)
	pongKey := getNodePongTimeKey(nodeID, net.IP{192, 168, 112, 109})

	id, ip, field := getNodeKeyItems(pongKey)

	require.Equal(t, id, nodeID)
	require.Equal(t, net.IP{192, 168, 112, 109}, ip)
	require.Equal(t, "pong", field)

}

func TestDBUpdateAndGet(t *testing.T) {
	node := NewWithParams(pubKey, net.IP{192, 168, 112, 109}, 90, 90)

	pongTime := time.Now()

	db, err := OpenDB("") // use memory db
	if err != nil {
		t.Error("Open db failed")
	}

	defer db.Close()

	if rt := db.PongReceived(*node.ID(), node.IP()); rt.Unix() != 0 {
		t.Errorf("non-existing object: %v", rt)
	}

	if err := db.UpdatePongReceived(*node.ID(), node.IP(), pongTime); err != nil {
		t.Errorf("failed to update: %v", err)
	}

	if pt := db.PongReceived(*node.ID(), node.IP()); pt.Unix() != pongTime.Unix() {
		t.Errorf("pong time value mismatch: have %v, want %v", pt, pongTime)
	}
}
