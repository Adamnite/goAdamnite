package bargossip

import (
	"bufio"
	"fmt"
	"os"

	"github.com/adamnite/go-adamnite/internal/bargossip/msg"
	"github.com/adamnite/go-adamnite/log"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type RemoteNode struct {
	addrInfo peer.AddrInfo
	stream   network.Stream
	peerType uint8

	rw       *bufio.ReadWriter
	curReqID uint64

	// Channels
	chDisconPeer chan RemoteNode
	stopCh       chan bool
	responseCh   chan msg.BargossipMsg
}

func CreateRemoteNode(addrInfo peer.AddrInfo, chDisconPeer chan RemoteNode) RemoteNode {
	remoteNode := RemoteNode{
		addrInfo:     addrInfo,
		chDisconPeer: chDisconPeer,
		stopCh:       make(chan bool),
		responseCh:   make(chan msg.BargossipMsg),
		curReqID:     0,
	}

	return remoteNode
}

func (n *RemoteNode) Start() {
	n.rw = bufio.NewReadWriter(bufio.NewReader(n.stream), bufio.NewWriter(n.stream))
	log.Trace("RW was opened")
	go n.writeData(n.rw)
	go n.readData(n.rw)
}

func (n *RemoteNode) Stop() {
	n.stream.Close()
	// n.stopCh <- true
}

func (n *RemoteNode) readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			// TODO: Remove peer at peers
			n.chDisconPeer <- *n
			return
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func (n *RemoteNode) writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func (n *RemoteNode) sendCurrentBlockNumberRequest() (*msg.BargossipMsg, error) {
	curReqBlockNum := msg.NewCreateCurrentBlockRequest(n.curReqID)
	byMsg, err := curReqBlockNum.ToSerialize()
	if err != nil {
		return nil, err
	}
	_, err = n.rw.Write(byMsg)
	if err != nil {
		return nil, err
	}

	err = n.rw.Flush()
	if err != nil {
		return nil, err
	}

	response := <-n.responseCh

	return &response, nil
}
