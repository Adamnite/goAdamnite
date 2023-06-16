package gossip

import (
	"math/big"
)

// DeBruijnGraph represents the deBruijn graph in the Koorde protocol.
type DeBruijnGraph struct {
	Size   int         // Size of the graph (number of nodes)
	Lookup map[*Key]*Key // Lookup table for node connections
}

// NewDeBruijnGraph creates a new deBruijn graph with the specified size.
func NewDeBruijnGraph(size int) *DeBruijnGraph {
	graph := &DeBruijnGraph{
		Size:   size,
		Lookup: make(map[*Key]*Key),
	}

	// Generate the node connections
	for i := 0; i < size; i++ {
		key := createKeyFromIndex(i)
		successor := createKeyFromIndex(i + 1)

		graph.Lookup[key] = successor
	}

	return graph
}

// GetSuccessor returns the successor key of the given key in the deBruijn graph.
func (graph *DeBruijnGraph) GetSuccessor(key *Key) (*Key, bool) {
	successor, ok := graph.Lookup[key]
	return successor, ok
}

// createKeyFromIndex creates a key based on the given index.
func createKeyFromIndex(index int) *Key {
	id := big.NewInt(int64(index % KeyspaceSize))
	return &Key{ID: id}
}
