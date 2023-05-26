package merkle

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
)

// Data represents the data that is stored and verified by the tree.
type Data interface {
	Hash() ([]byte, error)
	Equals(data Data) (bool, error)
}

type MerkleTree struct {
	Root         *Node
	Leafs        []*Node
	hashStrategy func() hash.Hash
}

type Node struct {
	Tree      *MerkleTree
	Parent    *Node
	Left      *Node
	Right     *Node
	Hash      []byte
	duplicate bool
	D         Data
}

// verify walks down the tree and returns the current node hash
func (n *Node) verify() ([]byte, error) {
	if n.Left == nil && n.Right == nil {
		// leaf node
		return n.D.Hash()
	}

	leftHash, err := n.Left.verify()
	if err != nil {
		return nil, err
	}

	rightHash, err := n.Right.verify()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftHash, rightHash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// calculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node) calculateNodeHash() ([]byte, error) {
	if n.Left == nil && n.Right == nil {
		// leaf node
		return n.D.Hash()
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// NewEmptyTree creates a new empty Merkle Tree.
func NewEmptyTree() *MerkleTree {
	return &MerkleTree{
		hashStrategy: sha256.New,
	}
}

// NewTree creates a new Merkle Tree using the provided data.
func NewTree(data []Data) (*MerkleTree, error) {
	m := &MerkleTree{
		hashStrategy: sha256.New,
	}
	err := m.buildWithData(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// NewTreeWithHashStrategy creates a new Merkle Tree using the provided data and hashing strategy
func NewTreeWithHashStrategy(data []Data, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	m := &MerkleTree{
		hashStrategy: hashStrategy,
	}
	err := m.buildWithData(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetMerklePath gets Merkle path and indexes (left leaf or right leaf)
func (m *MerkleTree) GetMerklePath(data Data) ([][]byte, []int64, error) {
	for _, current := range m.Leafs {
		ok, err := current.D.Equals(data)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			currentParent := current.Parent
			var merklePath [][]byte
			var index []int64
			for currentParent != nil {
				if bytes.Equal(currentParent.Left.Hash, current.Hash) {
					merklePath = append(merklePath, currentParent.Right.Hash)
					index = append(index, 1) // right leaf
				} else {
					merklePath = append(merklePath, currentParent.Left.Hash)
					index = append(index, 0) // left leaf
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, index, nil
		}
	}
	return nil, nil, nil
}

func (m *MerkleTree) buildWithData(data []Data) error {
	if len(data) == 0 {
		return errors.New("Cannot construct Merkle tree with no data")
	}
	var leafs []*Node
	for _, d := range data {
		hash, err := d.Hash()
		if err != nil {
			return err
		}

		leafs = append(leafs, &Node{
			Hash: hash,
			D:    d,
			Tree: m,
		})
	}

	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Tree:      m,
			Hash:      leafs[len(leafs)-1].Hash,
			D:         leafs[len(leafs)-1].D,
			duplicate: true,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, m)
	if err != nil {
		return err
	}

	m.Root = root
	m.Leafs = leafs
	return nil
}

// buildIntermediate constructs the intermediate and root levels of the tree.
func buildIntermediate(ns []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(ns); i += 2 {
		h := t.hashStrategy()
		var left, right int = i, i + 1
		if i+1 == len(ns) {
			right = i
		}
		hash := append(ns[left].Hash, ns[right].Hash...)
		if _, err := h.Write(hash); err != nil {
			return nil, err
		}
		n := &Node{
			Left:  ns[left],
			Right: ns[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}
		nodes = append(nodes, n)
		ns[left].Parent = n
		ns[right].Parent = n
		if len(ns) == 2 {
			return n, nil
		}
	}
	return buildIntermediate(nodes, t)
}

// Rebuild rebuilds the tree by reusing the data held in the leaves.
func (m *MerkleTree) Rebuild() error {
	var data []Data
	for _, l := range m.Leafs {
		data = append(data, l.D)
	}
	err := m.buildWithData(data)
	if err != nil {
		return err
	}
	return nil
}

// RebuildWithData replaces the data in the tree.
func (m *MerkleTree) RebuildWithData(data []Data) error {
	err := m.buildWithData(data)
	if err != nil {
		return err
	}
	return nil
}

// Verify verifies the tree by validating the hashes at each level of the tree.
// Returns true if the resulting hash at the root of the tree matches the resulting root hash.
// Returns false otherwise.
func (m *MerkleTree) Verify() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verify()
	if err != nil {
		return false, err
	}

	if bytes.Equal(m.Root.Hash, calculatedMerkleRoot) {
		return true, nil
	}
	return false, nil
}

// VerifyData indicates whether the data is in the tree and the hashes are valid for that data.
// Returns true if the expected hash is equivalent to the hash calculated for a given data.
// Returns false otherwise.
func (m *MerkleTree) VerifyData(data Data) (bool, error) {
	for _, l := range m.Leafs {
		ok, err := l.D.Equals(data)
		if err != nil {
			return false, err
		}

		if ok {
			currentParent := l.Parent
			for currentParent != nil {
				h := m.hashStrategy()
				rightBytes, err := currentParent.Right.calculateNodeHash()
				if err != nil {
					return false, err
				}

				leftBytes, err := currentParent.Left.calculateNodeHash()
				if err != nil {
					return false, err
				}

				if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
					return false, err
				}
				if !bytes.Equal(h.Sum(nil), currentParent.Hash) {
					return false, nil
				}
				currentParent = currentParent.Parent
			}
			return true, nil
		}
	}
	return false, nil
}

func (n *Node) String() string {
	return fmt.Sprintf("%t %v %s", n.duplicate, n.Hash, n.D)
}

// Only leaf nodes are included in resulting string.
func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprintln(l)
	}
	return s
}