package networking

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
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
		seedNode.contactBook.AddConnection(&x.thisContact)
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
