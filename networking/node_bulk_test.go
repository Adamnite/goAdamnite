package networking

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/stretchr/testify/assert"
)

func TestLotsOfNodes(t *testing.T) {
	seedNode := NewNetNode(common.Address{0})
	seedNode.AddServer()

	fmt.Println("seed node has been spun up")
	seedContact := seedNode.thisContact

	//seed known are nodes that the seed knows of directly.
	seedKnowNodes := make([]*NetNode, 50)
	for i := 0; i < len(seedKnowNodes); i++ {
		x := NewNetNode(common.BytesToAddress(big.NewInt(int64(i + 1)).Bytes()))
		x.AddServer()
		x.ConnectToContact(&seedContact)
		seedKnowNodes[i] = x
	}
	seedNode.hostingServer.GetContactsFunction()
	for i, n := range seedKnowNodes {
		if err := n.SprawlConnections(1, 0); err != nil {
			fmt.Printf("error in sprawling at index: %v\n with error: %v\n", i, err)
			t.Fatal(err)
		}

		if !assert.Equal(
			t,
			len(seedKnowNodes)+1, //they always know the seed. The seed is the only one to not know the seed
			len(n.contactBook.connections),
			"",
		) {
			if len(seedKnowNodes)+1 > len(n.contactBook.connections) {
				fmt.Printf("at index %v, a node has not found all nodes. Node in question: %v", i, n)

			} else {
				fmt.Printf("at index %v, a node has found an extra node. Node in question: %v", i, n)
			}

			t.Fail()
		}
	}

}

func TestLinearPropagation(t *testing.T) {
	seedNode := NewNetNode(common.Address{0})
	seedNode.AddServer()
	seedContact := seedNode.thisContact
	fmt.Println("seed node has been spun up")
	goalNode := NewNetNode(common.Address{1, 2, 3, 4, 5, 6, 7, 8, 9})
	goalNode.AddServer()

	forwardingCount := big.NewInt(0)
	nodes := [5]*NetNode{}
	for i := range nodes {
		node := NewNetNode(common.BytesToAddress(big.NewInt(int64(i + 1)).Bytes()))
		nodes[i] = node
		if err := node.AddServer(); err != nil {
			t.Fatal(err)
		}
		node.maxInboundConnections = 1
		node.maxOutboundConnections = 1
	}
	for i := 0; i < len(nodes)-1; i++ {
		node := nodes[i]
		if err := node.ConnectToContact(&nodes[i+1].thisContact); err != nil {
			t.Fatal(err)
		}
	}
	getLastNodeContactsReply := []byte{}
	getLastNodesContacts := rpc.ForwardingContent{
		FinalEndpoint:   "AdamniteServer.GetContactList",
		DestinationNode: &nodes[len(nodes)-1].thisContact.NodeID,
		FinalParams:     []byte{},
		FinalReply:      getLastNodeContactsReply,
		InitialSender:   seedContact.NodeID,
	}
	if err := seedNode.ConnectToContact(&nodes[0].thisContact); err != nil {
		t.Fatal(err)
	}
	if err := seedNode.activeContactToClient[(&nodes[0].thisContact)].ForwardMessage(getLastNodesContacts, &[]byte{}); err != nil {
		t.Fatal(err)
	}

	fmt.Println(getLastNodeContactsReply)
	fmt.Println(forwardingCount)
}
