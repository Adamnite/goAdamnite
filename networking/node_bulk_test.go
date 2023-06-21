package networking

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestLotsOfNodes(t *testing.T) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
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
	con, _ := nodes[0].activeContactToClient.Load((&nodes[1].thisContact))
	if err := con.(*rpc.AdamniteClient).ForwardMessage(getLastNodesContacts, &[]byte{}); err != nil {
		if err.Error() != rpc.ErrAlreadyForwarded.Error() {
			t.Fatal(err)
		}
	}

	fmt.Println(getLastNodeContactsReply)
	fmt.Println(forwardingCount)
	for _, node := range nodes {

		if err := node.SprawlConnections(int(math.Log(float64(len(nodes)))), 0); err != nil && err != ErrNoNewConnectionsMade {
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
	}
	callingNodeIndex := len(nodes) / 2
	for x := 0; x < 5; x++ {
		propagateContent.FinalParams = []byte{byte(x)} //only works up to 255
		propagateContent.InitialTime = time.Now().UnixMilli()
		if err := nodes[callingNodeIndex].handleForward(propagateContent, &[]byte{}); err != nil {
			t.Fatal(err)
		}
		for nodeIndex, node := range nodes {
			if nodeIndex != callingNodeIndex { //shouldn't be calling on itself when forwarding alone
				assert.Equal(t, x+1, node.hostingServer.GetTestsCount(), "not every node received the forward.")
			}
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
	}
	for x := 0; x < 5; x++ {
		propagateContent.FinalParams = []byte{byte(x)} //only works up to 255
		propagateContent.InitialTime = time.Now().UnixMilli()
		if err := nodes[0].handleForward(propagateContent, &[]byte{}); err != nil {
			t.Fatal(err)
		}
		for i, node := range nodes {
			if i != 0 {
				assert.Equal(t, x+1, node.hostingServer.GetTestsCount(), "not every node received the forward.")
			}
		}
	}
}

func TestTransactionPropagation(t *testing.T) {
	nodes, err := generateClusteredNodes(10, 15)
	if err != nil {
		t.Fatal(err)
	}

	testerNode := NewNetNode(common.Address{0xFF, 0xFF, 0xFF, 0xFF})
	var ans = &utils.BaseTransaction{}
	testerNode.AddFullServer(&statedb.StateDB{}, &blockchain.Blockchain{}, func(foo utils.TransactionType) error {
		*ans = *foo.(*utils.BaseTransaction)
		return nil
	}, nil, nil, nil)
	if err = nodes[0][0].ConnectToContact(&testerNode.thisContact); err != nil {
		t.Fatal(err)
	}
	// outsideNode := nodes[len(nodes)-1][len(nodes[0])-1]
	outsideNode := NewNetNode(common.Address{0xAF, 0xFF, 0xFF, 0xFF})
	// outsideNode.contactBook.connectionsByContact
	outsideNode.AddFullServer(&statedb.StateDB{}, &blockchain.Blockchain{}, nil, nil, nil, nil)
	outsideNode.ConnectToContact(&nodes[len(nodes)-1][len(nodes[0])-1].thisContact)
	client := NewNetNode(common.Address{0xAF, 0xFF, 0xFF, 0x00})
	err = client.AddFullServer(nil, nil, nil, nil, nil, nil)
	client.ConnectToContact(&outsideNode.thisContact)
	// client, err := rpc.NewAdamniteClient(outsideNode.thisContact.ConnectionString)
	if err != nil {
		t.Fatal(err)
	}
	sender, _ := accounts.GenerateAccount()
	transaction, err := utils.NewBaseTransaction(
		sender,
		common.Address{0xB, 1, 2, 3, 4, 5},
		big.NewInt(1000),
		big.NewInt(1000),
	)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	log.Println("\n\nInfo")

	if err := client.Propagate(transaction); err != nil {
		t.Fatal(err)
	}

	// log.Println(ans.From.Bytes())
	// log.Println(ans.To.Bytes())
	// log.Println(ans.Amount)
	// log.Println(ans.Time)
	// log.Println(ans.Signature)
	// log.Println(ans.Equal(transaction))
	assert.True(t, ans.Equal(*transaction), "failed to return equal transaction")
	// assert.Equal(t, transaction, *ans, "not equal")
}

// generates a line where each node is connected to the one in front, and behind itself.
func generateLineOfNodes(count int) ([]*NetNode, error) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	nodes := make([]*NetNode, count)
	for i := range nodes {
		node := NewNetNode(common.BytesToAddress(big.NewInt(int64(i + 1)).Bytes()))
		nodes[i] = node
		if err := node.AddServer(); err != nil {
			return nil, err
		}
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

// generate clusters of nodes
func generateClusteredNodes(clusterCount, clusterSize int) ([][]*NetNode, error) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	nodes := make([][]*NetNode, clusterCount)
	for x := 0; x < clusterCount; x++ {
		nodeRow := []*NetNode{}
		for y := 0; y < clusterSize; y++ {
			node := NewNetNode(common.Address{byte(x), byte(y)})
			nodeRow = append(nodeRow, node)
			node.maxOutboundConnections = uint(clusterCount) + uint(clusterSize) //let one node connect to an entire row and column
			if err := node.AddServer(); err != nil {
				return nil, err
			}
			if y != 0 {
				if err := node.ConnectToContact(&nodeRow[0].thisContact); err != nil {
					return nil, err
				}

				if err := node.FillOpenConnections(); err != nil {
					return nil, err
				}
			}
			nodes[x] = nodeRow
		}
	}
	for x := range nodes {
		for y := range nodes {
			if nodes[x][0].thisContact != nodes[y][0].thisContact {
				if err := nodes[x][0].ConnectToContact(&nodes[y][0].thisContact); err != nil && err != ErrPreexistingConnection {
					return nodes, err
				}
				if err := nodes[y][0].ConnectToContact(&nodes[x][0].thisContact); err != nil && err != ErrPreexistingConnection {
					return nodes, err
				}
			}
		}
	}
	return nodes, nil
}
