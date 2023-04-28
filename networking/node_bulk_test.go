package networking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLotsOfNodes(t *testing.T) {
	seedNode := NewNetNode()
	seedNode.AddServer()
	seedContact := seedNode.thisContact

	//seed known are nodes that the seed knows of directly.
	seedKnowNodes := make([]*NetNode, 50)
	for i := 0; i < len(seedKnowNodes); i++ {
		x := NewNetNode()
		x.AddServer()
		seedNode.contactBook.AddConnection(&x.thisContact)
		x.ConnectToContact(&seedContact)
		seedKnowNodes[i] = x
	}
	for i, n := range seedKnowNodes {
		fmt.Println(i)
		if err := n.SprawlConnections(1, 0); err != nil {
			fmt.Println("error in sprawling")
			t.Fatal(err)
		}

		assert.Equal(
			t,
			len(seedKnowNodes),
			len(n.contactBook.connections),
			"it appears at least one contact is missing, or overly added.",
		)
	}

}
