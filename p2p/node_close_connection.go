package p2p

import (
	"encoding/json"
	"io"
	"log"
)

func (n *Node) CloseConnection(conn io.Writer) {
	msg := Msg{
		MsgType: -1,
	}

	msgb, err := json.Marshal(&msg)
	if err != nil {
		log.Println("Error could not convert go type to json")
		return
	}

	msgb = append(msgb, '\n')
	if _, err := conn.Write(msgb); err != nil {
		log.Println("Could not send msg request")
	}
	// need to close the connection as well
}
