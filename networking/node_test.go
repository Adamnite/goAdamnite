package networking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTwoNodes(t *testing.T) {
	nodeA := NewNetNode()
	nodeB := NewNetNode()
	if err := nodeA.AddServer(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("first server spun up")
	if err := nodeB.ConnectToContact(&nodeA.thisContact); err != nil {
		t.Fatal(err)
	}
	fmt.Println("second node connected")
	if assert.Equal(t, 1, len(nodeB.contactBook.connections), "node did not add contact") {
		assert.Equal(t, &nodeA.thisContact, nodeB.contactBook.connections[0].contact, "nodeB appears to have not correctly added the contact.")
	}
	fmt.Println("first connection successful")
	assert.Equal(t, ErrPreexistingConnection, nodeB.ConnectToContact(&nodeA.thisContact))
	fmt.Println("all worked!")
}
func TestTwoNodesFlagChanges(t *testing.T) {
	nodeA := NewNetNode()
	nodeB := NewNetNode()
	if err := nodeA.AddServer(); err != nil {
		t.Fatal(err)
	}
	//proper connection
	if err := nodeB.ConnectToContact(&nodeA.thisContact); err != nil {
		t.Fatal(err)
	}
	wrongConnectionString := &Contact{
		NodeID:           nodeA.thisContact.NodeID,
		connectionString: "not a connection string",
	}

	//test that it does throw an error!
	if err := nodeB.ConnectToContact(wrongConnectionString); err == nil {
		t.Fatal("this should panic realizing that the connection string at this connection point is different")
	} else {
		// fmt.Println(err)
	}

	wrongNodeID := &Contact{
		NodeID:           1234567,
		connectionString: nodeA.thisContact.connectionString,
	}

	//test that it does throw an error!
	if err := nodeB.ConnectToContact(wrongNodeID); err == nil {
		t.Fatal("this should panic realizing that the nodeID at this connection point is different")
	} else {
		// fmt.Println(err)
	}
}

func TestSingleNode(t *testing.T) {
	testingNode := NewNetNode()

	fmt.Println(testingNode.thisContact.NodeID)

}
