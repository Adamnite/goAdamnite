package gossip

import (
	"fmt"
	"math/big"
)

// Table represents the hash table in the Koorde protocol.
type Table struct {
	LocalNode *Node
	LocalTable *LocalNodeTable
	DeBruijn   *DeBruijnGraph
	Nodes      map[string]*Node
}

// LocalNodeTable represents the local node's table of successors and predecessors.
type LocalNodeTable struct {
	Successor    *Node
	Predecessor  *Node
	FingerTable  []*Node
}

// Node represents a node in the Koorde protocol.
type Node struct {
	Key      *Key
	Location string
}

// Key represents a key in the keyspace.
type Key struct {
	ID *big.Int
}

// DeBruijnGraph represents the DeBruijn graph used in Koorde.
type DeBruijnGraph struct {
	Size  int
	Shift int
}

// NewTable initializes a new hash table in the Koorde protocol.
func NewTable(localKey *Key, size int) *Table {
	deBruijn := NewDeBruijnGraph(size)
	localNode := &Node{
		Key:      localKey,
		Location: "Local",
	}
	localTable := &LocalNodeTable{
		Successor:   localNode,
		Predecessor: localNode,
		FingerTable: make([]*Node, deBruijn.Size),
	}
	return &Table{
		LocalNode:  localNode,
		LocalTable: localTable,
		DeBruijn:   deBruijn,
		Nodes:      make(map[string]*Node),
	}
}

// NewDeBruijnGraph initializes a new DeBruijn graph with the given size.
func NewDeBruijnGraph(size int) *DeBruijnGraph {
	shift := 0
	for (1 << shift) < size {
		shift++
	}
	return &DeBruijnGraph{
		Size:  size,
		Shift: shift,
	}
}

// JoinNode joins a new node to the hash table.
func (ht *Table) JoinNode(node *Node) {
	ht.Nodes[node.Key.String()] = node
}

// LeaveNode removes a node from the hash table.
func (ht *Table) LeaveNode(node *Node) {
	delete(ht.Nodes, node.Key.String())
}

// Stabilize performs stabilization in the hash table to ensure consistent routing.
func (ht *Table) Stabilize() {
	// Stabilize performs stabilization in the hash table to ensure consistent routing.
func (ht *HashTable) Stabilize() {
	for _, node := range ht.Nodes {
		successor, ok := ht.DeBruijn.GetSuccessor(node.Key)
		if !ok {
			fmt.Printf("No successor found for node with key: %s\n", node.Key)
			continue
		}

		// Check if the current node is the immediate successor of the successor node.
		// If not, update the successor and perform finger table updates.
		if !ht.isImmediateSuccessor(node.Key, successor.Key) {
			// Update the successor
			successor = ht.findSuccessor(node.Key)

			// Perform finger table updates
			ht.updateFingerTable(node, successor)
		}

		// Check if the successor is still valid using the circle method
		if !ht.isAlive(successor) {
			// Find a new successor using the circle method
			successor = ht.findSuccessor(node.Key)

			// Perform finger table updates
			ht.updateFingerTable(node, successor)
		}

		// Perform routing efficiency checks and updates here...
		// Example: Update routing based on finger table entries.
		// You can use node.Key and successor.Key to perform routing checks.

		// Example code to demonstrate routing efficiency update
		fingerTable := ht.buildFingerTable(node)
		for i, finger := range fingerTable {
			if ht.shouldUpdateFinger(finger, successor) {
				// Update finger table entry at index i with the successor
				fingerTable[i] = successor
			}
		}
	}
}

// isImmediateSuccessor checks if key1 is the immediate successor of key2 in the keyspace.
func (ht *HashTable) isImmediateSuccessor(key1, key2 *Key) bool {
	return key1.ID.Cmp(key2.ID) == 0
}

// findSuccessor finds the successor of the given key in the hash table.
func (ht *HashTable) findSuccessor(key *Key) *Node {
	successor, ok := ht.DeBruijn.GetSuccessor(key)
	if !ok {
		return nil
	}

	return ht.Nodes[successor]
}

// isAlive checks if the given node is still alive in the hash table.
func (ht *HashTable) isAlive(node *Node) bool {
	_, ok := ht.Nodes[node.Key]
	return ok
}

// updateFingerTable updates the finger table entries of the given node with the new successor.
func (ht *HashTable) updateFingerTable(node, successor *Node) {
	fingerTable := ht.buildFingerTable(node)
	for i := range fingerTable {
		start := ht.DeBruijn.ComputeFingerStart(node.Key, i)
		if ht.isImmediateSuccessor(start, successor.Key) {
			fingerTable[i] = successor
		}
	}
}

// buildFingerTable builds the finger table entries for the given node.
func (ht *HashTable) buildFingerTable(node *Node) []*Node {
	fingerTable := make([]*Node, ht.DeBruijn.Size())
	for i := range fingerTable {
		fingerStart := ht.DeBruijn.ComputeFingerStart(node.Key, i)
		fingerTable[i] = ht.findSuccessor(fingerStart)
	}
	return fingerTable
}

// shouldUpdateFinger checks if the finger table entry should be updated with the new successor.
func (ht *HashTable) shouldUpdateFinger(finger, successor *Node) bool {
	return ht.isImmediateSuccessor(finger.Key, successor.Key) && !ht.isAlive(finger)
}


	}

// stabilizeLocalTable performs stabilization for the local node's table.
func (ht *Table) stabilizeLocalTable() {
	predecessor := ht.LocalTable.Predecessor
	successor := ht.LocalTable.Successor

	if successor != ht.LocalNode {
		if !ht.isImmediateSuccessor(successor.Key, ht.LocalNode.Key) {
			successor = ht.findSuccessor(ht.LocalNode.Key)
			ht.updateFingerTable(successor)
		}

		if !ht.isAlive(successor) {
			successor = ht.findSuccessor(ht.LocalNode.Key)
			ht.updateFingerTable(successor)
		}
	}

	if predecessor != ht.LocalNode {
		if !ht.isAlive(predecessor) || ht.isImmediateSuccessor(ht.LocalNode.Key, predecessor.Key) {
			predecessor = ht.findPredecessor(ht.LocalNode.Key)
			ht.updatePredecessor(predecessor)
		}

		if !ht.isAlive(predecessor) {
			predecessor = ht.findPredecessor(ht.LocalNode.Key)
			ht.updatePredecessor(predecessor)
		}
	}
}

// stabilizeRemoteTable performs stabilization for a remote node's table.
func (ht *Table) stabilizeRemoteTable(node *Node) {
	successor := node
	if !ht.isImmediateSuccessor(successor.Key, node.Key) {
		successor = ht.findSuccessor(node.Key)
	}

	if !ht.isAlive(successor) {
		successor = ht.findSuccessor(node.Key)
	}

	if successor != node {
		ht.updateRemoteSuccessor(node, successor)
	}
}

// updateFingerTable updates the finger table of the local node.
func (ht *Table) updateFingerTable(successor *Node) {
	for i := 0; i < ht.DeBruijn.Size; i++ {
		start := ht.calculateStart(ht.LocalNode.Key, i)
		finger := ht.LocalTable.FingerTable[i]

		if ht.shouldUpdateFinger(finger, successor) {
			finger = successor
		}

		if !ht.isAlive(finger) || ht.isImmediateSuccessor(start, finger.Key) {
			finger = ht.findSuccessor(start)
		}

		if !ht.isAlive(finger) {
			finger = ht.findSuccessor(start)
		}

		ht.LocalTable.FingerTable[i] = finger
	}
}

// updatePredecessor updates the predecessor of the local node.
func (ht *Table) updatePredecessor(predecessor *Node) {
	ht.LocalTable.Predecessor = predecessor
}

// updateRemoteSuccessor updates the successor of a remote node.
func (ht *Table) updateRemoteSuccessor(node, successor *Node) {
	node.Successor = successor
}

// isImmediateSuccessor checks if key1 is the immediate successor of key2 in the keyspace.
func (ht *Table) isImmediateSuccessor(key1, key2 *Key) bool {
	return key1.ID.Cmp(key2.ID) == 1
}

// isAlive checks if the node is alive and present in the hash table.
func (ht *Table) isAlive(node *Node) bool {
	_, exists := ht.Nodes[node.Key.String()]
	return exists
}

// findSuccessor finds the successor node for the given key.
func (ht *Table) findSuccessor(key *Key) *Node {
	predecessor := ht.findPredecessor(key)
	successor, err := ht.getSuccessor(predecessor)
	if err != nil {
		fmt.Println(err)
	}
	return successor
}

// findPredecessor finds the predecessor node for the given key.
func (ht *Table) findPredecessor(key *Key) *Node {
	node := ht.LocalNode
	successor := ht.LocalTable.Successor

	for !ht.isInInterval(key, node.Key, successor.Key) {
		node = ht.closestPrecedingFinger(key)
		successor, _ = ht.getSuccessor(node)
	}

	return node
}

// closestPrecedingFinger finds the closest preceding finger node for the given key.
func (ht *Table) closestPrecedingFinger(key *Key) *Node {
	for i := ht.DeBruijn.Size - 1; i >= 0; i-- {
		finger := ht.LocalTable.FingerTable[i]
		if finger != nil && ht.isInInterval(finger.Key, ht.LocalNode.Key, key) {
			return finger
		}
	}
	return ht.LocalNode
}

// getSuccessor returns the successor of the given node.
func (ht *Table) getSuccessor(node *Node) (*Node, error) {
	successor := node.Successor
	if successor == nil {
		return nil, fmt.Errorf("successor not found for node %v", node.Key)
	}
	return successor, nil
}

// isInInterval checks if key is in the interval (start, end].
func (ht *Table) isInInterval(key, start, end *Key) bool {
	if start.ID.Cmp(end.ID) < 0 {
		return key.ID.Cmp(start.ID) > 0 && key.ID.Cmp(end.ID) <= 0
	}
	return key.ID.Cmp(start.ID) > 0 || key.ID.Cmp(end.ID) <= 0
}

// shouldUpdateFinger checks if the finger needs to be updated.
func (ht *Table) shouldUpdateFinger(finger, successor *Node) bool {
	if finger == nil {
		return true
	}
	return ht.isInInterval(successor.Key, ht.LocalNode.Key, finger.Key)
}

// calculateStart calculates the start key for a finger index.
func (ht *Table) calculateStart(key *Key, index int) *Key {
	exp := big.NewInt(int64(index))
	shift := big.NewInt(int64(ht.DeBruijn.Shift))
	size := big.NewInt(int64(ht.DeBruijn.Size))
	start := new(big.Int).Exp(big.NewInt(2), new(big.Int).Mul(exp, shift), size)
	start.Add(key.ID, start)
	return &Key{ID: start}
}


// getRandomNodes returns a slice of random nodes from the local table.
func (ht *Table) getRandomNodes(numNodes int) []*Node {
	nodes := make([]*Node, 0, numNodes)

	// Get all nodes from the local table
	for _, finger := range ht.LocalTable.FingerTable {
		if finger != nil {
			nodes = append(nodes, finger)
		}
	}
	if ht.LocalTable.Predecessor != nil {
		nodes = append(nodes, ht.LocalTable.Predecessor)
	}
	nodes = append(nodes, ht.LocalNode)

	// Shuffle the nodes randomly
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	// Return the requested number of random nodes
	if numNodes < len(nodes) {
		nodes = nodes[:numNodes]
	}

	return nodes
}

// reset resets the table to its default values.
func (ht *Table) reset() {
	ht.LocalNode = nil
	ht.LocalTable = nil
	ht.Nodes = make(map[string]*Node)
}

// Initialize initializes the hash table with the given local node.
func (ht *Table) Initialize(localNode *Node) {
	ht.reset()

	ht.LocalNode = localNode
	ht.LocalTable = &LocalNodeTable{
		Predecessor: nil,
		Successor:   localNode,
		FingerTable: make([]*Node, ht.DeBruijn.Size),
	}

	ht.Nodes[localNode.Key.String()] = localNode
}

// AddPeer adds a new peer to the hash table.
func (t *Table) AddPeer(peer *Node) {
	t.Successors = append(t.Successors, peer)
}

// DeletePeer deletes a peer from the hash table.
func (t *Table) DeletePeer(peer *Node) {
	for i, p := range t.Successors {
		if p.ID.Cmp(peer.ID) == 0 {
			t.Successors = append(t.Successors[:i], t.Successors[i+1:]...)
			break
		}
	}
}

// GetClosestSuccessor returns the closest successor of a given key.
func (t *Table) GetClosestSuccessor(key *big.Int) *Node {
	// Find the closest successor based on the key
	var closest *Node
	for _, successor := range t.Successors {
		if closest == nil || t.Ring.Sub(key, successor.ID).Cmp(t.Ring.Sub(key, closest.ID)) < 0 {
			closest = successor
		}
	}

	return closest
}