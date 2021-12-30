package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// HandleMessageRequest handles request and response from peer
func (n *Node) HandleMessageRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var msg Msg
	for {
		fmt.Println("listening for incomming message")
		line, err := reader.ReadSlice('\n')
		if err != nil {
			log.Println("Error could not read from conn", err)
			return
		}

		fmt.Println("Mesage received")
		if err = json.Unmarshal(line, &msg); err != nil {
			log.Println("Error could not unmarshall the message")
			return
		}
		switch msg.MsgType {
		// peer-sync-request
		case 0:
			resp := Msg{
				MsgType:    1,
				KnownPeers: n.KnownPeers,
				BlockMsg:   struct{}{},
			}
			fmt.Println("sending message", resp)
			respb, err := json.Marshal(resp)
			if err != nil {
				log.Println("could not craft message of  type 1, known peer response")
				continue
			}
			respb = append(respb, '\n')
			if _, err = conn.Write(respb); err != nil {
				log.Println("Could not send peer list data")
			}
		case -1:
			return
		}

	}
}
