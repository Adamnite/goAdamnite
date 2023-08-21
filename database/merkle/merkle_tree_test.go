package merkle

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

// TestData implements the Data interface provided by merkle and represents the data stored in the tree.
type TestData struct {
	s string
}

func (t TestData) Hash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.s)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (t TestData) Equals(other Data) (bool, error) {
	return t.s == other.(TestData).s, nil
}

func hashData(data []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

var (
	data = []Data{
		TestData{
			s: "Foo",
		},
		TestData{
			s: "Bar",
		},
		TestData{
			s: "Baz",
		},
	}
	missingData = TestData{
		s: "FooBar",
	}
	expectedHash = []byte{ 193, 183, 74, 3, 178, 72, 106, 229, 160, 223, 230, 101, 137, 196, 109, 160, 97, 123, 72, 183, 41, 168, 170, 19, 238, 52, 236, 163, 226, 99, 133, 2 }
)

func TestNewTree(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if bytes.Compare(tree.Root.Hash, expectedHash) != 0 {
		t.Errorf("Expected: %v, actual: %v", expectedHash, tree.Root.Hash)
	}
}

func TestNewTreeWithHashingStrategy(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if bytes.Compare(tree.Root.Hash, expectedHash) != 0 {
		t.Errorf("Expected: %v, actual: %v", expectedHash, tree.Root.Hash)
	}
}

func TestMerkleRoot(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if bytes.Compare(tree.Root.Hash, expectedHash) != 0 {
		t.Errorf("Expected: %v, actual: %v", expectedHash, tree.Root.Hash)
	}
}

func TestRebuildTree(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	err = tree.Rebuild()
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if bytes.Compare(tree.Root.Hash, expectedHash) != 0 {
		t.Errorf("Expected: %v, actual: %v", expectedHash, tree.Root.Hash)
	}
}

func TestRebuildTreeWith(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	err = tree.RebuildWithData(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if bytes.Compare(tree.Root.Hash, expectedHash) != 0 {
		t.Errorf("Expected: %v, actual: %v", expectedHash, tree.Root.Hash)
	}
}

func TestVerify(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	v1, err := tree.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if v1 != true {
		t.Errorf("Tree should be valid")
	}
	tree.Root.Hash = []byte{1}
	v2, err := tree.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if v2 != false {
		t.Errorf("Tree should be invalid")
	}
}

func TestVerifyData(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if len(data) > 0 {
		v, err := tree.VerifyData(data[0])
		if err != nil {
			t.Fatal(err)
		}
		if !v {
			t.Errorf("Data should be valid")
		}
	}
	if len(data) > 1 {
		v, err := tree.VerifyData(data[1])
		if err != nil {
			t.Fatal(err)
		}
		if !v {
			t.Errorf("Data should be valid")
		}
	}
	if len(data) > 2 {
		v, err := tree.VerifyData(data[2])
		if err != nil {
			t.Fatal(err)
		}
		if !v {
			t.Errorf("Data should be valid")
		}
	}
	if len(data) > 0 {
		tree.Root.Hash = []byte{1}
		v, err := tree.VerifyData(data[0])
		if err != nil {
			t.Fatal(err)
		}
		if v {
			t.Errorf("Data should be invalid")
		}
		if err := tree.Rebuild(); err != nil {
			t.Fatal(err)
		}
	}
	v, err := tree.VerifyData(missingData)
	if err != nil {
		t.Fatal(err)
	}
	if v {
		t.Errorf("Data should be invalid")
	}
}

func TestString(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	if tree.String() == "" {
		t.Errorf("String should be non-empty")
	}
}

func TestMerklePath(t *testing.T) {
	tree, err := NewTree(data)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	for i := 0; i < len(data); i++ {
		merklePath, index, _ := tree.GetMerklePath(data[i])

		hash, err := tree.Leafs[i].calculateNodeHash()
		if err != nil {
			t.Errorf("Error %v", err)
		}
		h := sha256.New()
		for j := 0; j < len(merklePath); j++ {

			if index[j] == 1 {
				hash = append(hash, merklePath[j]...)
			} else {
				hash = append(merklePath[j], hash...)
			}
			if _, err := h.Write(hash); err != nil {
				t.Errorf("Error %v", err)
			}
			hash, err = hashData(hash)
			if err != nil {
				t.Errorf("Error %v", err)
			}
		}
		if bytes.Compare(tree.Root.Hash, hash) != 0 {
			t.Errorf("Expected: %v, actual: %v", hash, tree.Root.Hash)
		}
	}
}