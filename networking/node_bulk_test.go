package networking

import (
	"fmt"
	"math"
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

func TestLinearForward(t *testing.T) {
	fmt.Println("seed node has been spun up")

	forwardingCount := big.NewInt(0)
	nodes, err := generateLineOfNodes(10)
	if err != nil {
		t.Fatal(err)
	}
	getLastNodeContactsReply := []byte{}
	getLastNodesContacts := rpc.ForwardingContent{
		FinalEndpoint:   "AdamniteServer.GetContactList",
		DestinationNode: &nodes[len(nodes)-1].thisContact.NodeID,
		FinalParams:     []byte{},
		FinalReply:      getLastNodeContactsReply,
		InitialSender:   nodes[0].thisContact.NodeID,
	}
	if err := nodes[0].activeContactToClient[(&nodes[1].thisContact)].ForwardMessage(getLastNodesContacts, &[]byte{}); err != nil {
		if err.Error() != rpc.ErrAlreadyForwarded.Error() {
			t.Fatal(err)
		}
	}

	fmt.Println(getLastNodeContactsReply)
	fmt.Println(forwardingCount)
	for _, node := range nodes {

		if err := node.SprawlConnections(int(math.Log(float64(len(nodes)))), 0); err != nil {
			t.Fatal(err)
		}
	}
	for _, node := range nodes {
		assert.Equal(t, len(nodes), len(node.contactBook.connections), "nodes couldn't find everyone.")

	}
}

func TestLinearPropagationFromCenter(t *testing.T) {
	nodes, err := generateLineOfNodes(400)
	if err != nil {
		t.Fatal(err)
	}
	propagateContent := rpc.ForwardingContent{
		FinalEndpoint:   rpc.TestServerEndpoint,
		DestinationNode: nil,
		FinalParams:     []byte{},
		FinalReply:      []byte{},
		InitialSender:   nodes[0].thisContact.NodeID,
		Signature:       common.Hash{0},
	}
	for x := 0; x < 5; x++ {
		propagateContent.Signature = common.HexToHash(fmt.Sprintf("0x%X", x)) //this only works up to 99, and not well lol
		if err := nodes[len(nodes)/2].handleForward(propagateContent, []byte{}); err != nil {
			t.Fatal(err)
		}
		for _, node := range nodes {
			assert.Equal(t, x+1, node.hostingServer.GetTestsCount(), "not every node received the forward.")
		}
	}

}
func TestLinearPropagationFromSide(t *testing.T) {
	nodes, err := generateLineOfNodes(400)
	if err != nil {
		t.Fatal(err)
	}
	propagateContent := rpc.ForwardingContent{
		FinalEndpoint:   rpc.TestServerEndpoint,
		DestinationNode: nil,
		FinalParams:     []byte{},
		FinalReply:      []byte{},
		InitialSender:   nodes[0].thisContact.NodeID,
		Signature:       common.Hash{0},
	}
	for x := 0; x < 5; x++ {
		propagateContent.Signature = common.HexToHash(fmt.Sprintf("0x%X", x)) //this only works up to 99, and not well lol
		if err := nodes[0].handleForward(propagateContent, []byte{}); err != nil {
			t.Fatal(err)
		}
		for _, node := range nodes {
			assert.Equal(t, x+1, node.hostingServer.GetTestsCount(), "not every node received the forward.")
		}
	}
}

// generates a line where each node is connected to the one in front, and behind itself.
func generateLineOfNodes(count int) ([]*NetNode, error) {
	nodes := make([]*NetNode, count)
	for i := range nodes {
		node := NewNetNode(common.BytesToAddress(big.NewInt(int64(i + 1)).Bytes()))
		nodes[i] = node
		if err := node.AddServer(); err != nil {
			return nil, err
		}
		node.maxInboundConnections = 2
		node.maxOutboundConnections = 2
	}
	nodes[0].ConnectToContact(&nodes[1].thisContact)
	for i := 1; i < len(nodes)-1; i++ {
		node := nodes[i]
		if err := node.ConnectToContact(&nodes[i+1].thisContact); err != nil {
			return nil, err
		}
		if err := node.ConnectToContact(&nodes[i-1].thisContact); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}
