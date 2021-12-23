package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
)

// just a place holder for struct right now
type nodes string

// sample Block
type Block struct{}

type Node struct {
	BootStrapNodes []nodes
	Port           string
	Addr           *net.UDPAddr
	Dailer         *net.Conn
}

// New creates a new peer and returns it
func New() Node {
	var n Node
	return n
}

// Send sends a message to specific host
func (n *Node) Send(msg []byte, dest net.Addr) {
	fmt.Fprintf(*n.Dailer, string(msg))
}
func (n *Node) Status(b *Block, dest net.Addr) {
	msg := bytes.Buffer{}
	enc := gob.NewEncoder(&msg)

	if err := enc.Encode(b); err != nil {
		// err
	}

	n.Send(msg.Bytes(), dest)
}

func (n *Node) Listen() {
	listener, err := net.ListenUDP("UDP", n.Addr)
	if err != nil {
		// ...
	}

	msg := []byte{}
	for {
		listener.Read(msg)
	}
}
