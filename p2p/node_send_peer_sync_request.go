package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// SendPeerSyncRequest sends peer sync request
func (n *Node) SendPeerSyncRequest(conn *net.Conn) map[string]PeerNode {

	// Peer list
	peerList := map[string]PeerNode{}

	// MstType 0 indicates peer sync request
	msg := Msg{
		MsgType: 0,
	}

	fmt.Println("Sending connect message")
	// marshallig to send data
	msgb, err := json.Marshal(&msg)
	if err != nil {
		log.Println("could not craft peer request message")
		return peerList
	}

	msgb = append(msgb, '\n')
	writer := bufio.NewWriter(*conn)

	if _, err = writer.Write(msgb); err != nil {
		log.Println("Could not send msg request")
	}
	writer.Flush()
	fmt.Println("writer wrote the message")
	// waiting for response
	reader := bufio.NewReader(*conn)
	line, err := reader.ReadSlice('\n')
	if err != nil {
		log.Println("Could not read meessage response")
		return peerList
	}

	// Receiving response
	var resp Msg

	if err = json.Unmarshal(line, &resp); err != nil {
		log.Println("Could not understand peer list response")
		return peerList
	}
	fmt.Println("Received", resp)
	peerList = resp.KnownPeers
	return peerList
}
