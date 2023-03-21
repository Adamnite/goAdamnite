package state

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

// This interface will be implemented by every datatype that could be stored inside the trie
type TrieElement interface {
	Value(key []byte) interface{}
	Hash(off int) []byte
	BinaryKey(key []byte) binkey
}

// ==========================================================================================
// An account that will be stored inside the MBT - state trie
type AdmAccountBranch struct {
	branch
	Balance 	*big.Int
	Nonce		 uint64
	CodeHash	*common.Hash
	StorageHash *common.Hash 
}

func (a *AdmAccountBranch) isRegularAccount() bool {
	return a.CodeHash == nil
}

func (a *AdmAccountBranch) GetCodeHash() {
	// Make a request to the offchain database with the .CodeHash 
	// And retrieve the code
}

func (a *AdmAccountBranch) GetStorage() {
	// Make a request to the offchain database with the .StorageHash
}

func (a *AdmAccountBranch) GetBalance() *big.Int {
	return big.NewInt(29)
}


func (a *AdmAccountBranch) GetNonce() uint64 {
	return 10
}

func (a *AdmAccountBranch) Value(key []byte) interface{} {
	if a.value == nil || len(key) != 32 {
		panic(fmt.Sprintf("trying to get the value of an internal node %d : value = %p", len(key), a.value))
	}

	if len(key) != 32 {
		panic(fmt.Sprintf("trying to get the value of an internal node %d : value = %p", len(key), a.value))
	}

	switch key[31] & 0x3 {
	case 0:
		if a.Balance == nil {
			panic(fmt.Sprintf("nil value at %x", key))
		}
		return a.Balance
	case 1:
		return a.Nonce
	case 2:
		return a.CodeHash
	case 3:
		// The root of the storage trie is stored in
		// the account level node's right value.
		return common.BytesToHash(a.right.hash(256))
	default:
		panic("property not found")
	}
}

func (a *AdmAccountBranch) Commit() error {
	if a.CommitCh == nil {
		return fmt.Errorf("commit channel missing")
	}
	a.hash(0)
	close(a.CommitCh)
	return nil
}

func (a *AdmAccountBranch) Hash() []byte {
	return a.hash(0)
}

func (a *AdmAccountBranch) hash(off int) []byte {
	// Special hashing case if this is called
	// at data-level
	if off+len(a.prefix) > 0 {
		return a.HashAccount(off)
	}
	return a.hash(off)
}

func (a *AdmAccountBranch) HashAccount(off int) []byte {
	var hash []byte
	var hasher *trieHasher = GetHasher()
	defer PutHasher(hasher)

	// Write the balance
	hasher.sha.Write(a.key[0:254])
	hasher.sha.Write([]byte{0, 0})
	kh := hasher.sha.Sum(nil)
	hasher.sha.Reset()
	hasher.sha.Write(a.Balance.Bytes())
	hash = hasher.sha.Sum(nil)
	hasher.sha.Reset()
	hasher.sha.Write(kh)
	hasher.sha.Write(hash)
	hash = hasher.sha.Sum(nil)
	hasher.sha.Reset()

	// Write the nonce
	hasher.sha.Write(a.key[0:254])
	hasher.sha.Write([]byte{0, 1})
	kh = hasher.sha.Sum(nil)
	hasher.sha.Reset()
	var serialized [32]byte
	binary.LittleEndian.PutUint64(serialized[:], a.Nonce)
	hasher.sha.Write(serialized[:])
	hash = hasher.sha.Sum(nil)
	hasher.sha.Reset()
	hasher.sha.Write(kh)
	hasher.sha.Write(hash)
	hash = hasher.sha.Sum(nil)
	hasher.sha.Reset()


	// Non regular account case, add the prefix length
	ok := bytes.Equal(a.StorageHash[:], emptyRootSha512[:])
	if len(a.CodeHash) == 0 && ok {
		hasher.sha.Write([]byte{255}) // depth is known
		hasher.sha.Write(zero32[:31])
		hasher.sha.Write(hash)
		hash = hasher.sha.Sum(nil)
		return hash
	}

	// Contract with storage case
	if len(a.CodeHash) != 0 && !ok {
		// Write the code hash
		hasher.sha.Write(a.key[0:254])
		hasher.sha.Write([]byte{1, 0})
		kh = hasher.sha.Sum(nil)
		hasher.sha.Reset()
		hasher.sha.Write(a.CodeHash[:])
		hash2 := hasher.sha.Sum(nil)
		hasher.sha.Reset()
		hasher.sha.Write(kh)
		hasher.sha.Write(hash2)
		hash2 = hasher.sha.Sum(nil)
		hasher.sha.Reset()

		// Write the storage hash
		hasher.sha.Write(a.key[0:254])
		hasher.sha.Write([]byte{1, 1})
		kh = hasher.sha.Sum(nil)
		hasher.sha.Reset()
		hasher.sha.Write(a.right.hash(256)[:]) // storage hash is the right child of account "branch"
		hash2 = hasher.sha.Sum(nil)
		hasher.sha.Reset()
		hasher.sha.Write(kh)
		hasher.sha.Write(hash2)
		hash2 = hasher.sha.Sum(nil)
		hasher.sha.Reset()

		// merge the two sides of the branch at bit #254
		hasher.sha.Write(hash)
		hasher.sha.Write(hash2)
		hash = hasher.sha.Sum(nil)

		if len(a.prefix) > 0 {
			hasher.sha.Reset()
			hasher.sha.Write([]byte{254}) // depth is known
			hasher.sha.Write(zero32[:31])
			hasher.sha.Write(hash)
			hash = hasher.sha.Sum(nil)
		}

		return hash
	}
	panic("Can't proceed with hashing")
}


func NewAdmAccountBranch(prefix binkey, key []byte, value []byte) *AdmAccountBranch {
	br := &branch{
		prefix:     prefix,
		left:       empty(struct{}{}),
		right:      empty(struct{}{}),
		key:        key,
		value:      value,
		childCount: 0,
	}

	acc := &AdmAccountBranch{
		branch: *br,
	}

	// use msgpack here for encoding accounts
	return acc
}
// TryUpdate will set the trie's leaf at `key` to `value`. If there is
// no leaf at `key`, it will be created.
func (a *AdmAccountBranch) TryUpdate(key, value []byte, bt *BinaryTrie) error {
	bk := bt.keyCreationFn(key)
	off := 0 // Number of key bits that've been walked at current iteration

	// Go through the storage, find the parent node to
	// insert this (key, value) into.
	var currentNode *AdmAccountBranch
	switch bt.root.(type) {
	case empty:
		// This is when the trie hasn't been inserted
		// into, so initialize the root as a value.
		bt.root = NewAdmAccountBranch(bk, key, value)
		// bt.db.insert(key, value)
		return nil
	case *AdmAccountBranch:
		currentNode = bt.root.(*AdmAccountBranch)
	case hashBinaryNode:
		return errInsertIntoHash
	}
	for {
		if bk.SamePrefix(currentNode.prefix, off) {
			// The key matches the full node prefix, iterate
			// at  the child's level.
			var childNode *AdmAccountBranch
			off += len(currentNode.prefix)
			if bk[off] == 0 {
				childNode = bt.resolveNodeFn(currentNode.left, bk, off+1).(*AdmAccountBranch)
			} else {
				childNode = bt.resolveNodeFn(currentNode.right, bk, off+1).(*AdmAccountBranch)
			}
			var isLeaf bool
			if childNode == nil {
				childNode = NewAdmAccountBranch(bk[off+1:], nil, value)
				isLeaf = true
			}
			childNode.parent = &currentNode.branch

			// If the leaf count goes above the threshold,
			// mark it as used in the cache. Because every
			// new insert updates the path to the new leaf,
			// the least used branches will be pruned first.
			currentNode.childCount++ // one more child
			if currentNode.childCount > leafThreshold {
				bt.cache.Add(bk[:off], currentNode)
			}

			// Update the parent node's child reference
			if bk[off] == 0 {
				currentNode.left = childNode
			} else {
				currentNode.right = childNode
			}

			if isLeaf {
				break
			}
			currentNode = childNode
			off++
		} else {

			split := bk[off:].commonLength(currentNode.prefix)

			// A split is needed
			midNode := NewAdmAccountBranch(currentNode.prefix[split+1:], currentNode.key, currentNode.value)
			midNode.left = currentNode.left
			midNode.right = currentNode.right
			midNode.parent = &currentNode.branch

			currentNode.prefix = currentNode.prefix[:split]
			currentNode.value = nil
			childNode := NewAdmAccountBranch(bk[off+split+1:], key, value)
			childNode.parent = &currentNode.branch

			// Update the cache if the leaf count goes
			// over the threshold.
			currentNode.childCount++
			if currentNode.childCount > leafThreshold {
				bt.cache.Add(bk[:off], currentNode)
			}

			// Set the child node to the correct branch.
			if bk[off+split] == 1 {
				// New node goes on the right
				currentNode.left = midNode
				currentNode.right = childNode
			} else {
				// New node goes on the left
				currentNode.right = midNode
				currentNode.left = childNode
			}

			break
		}
	}

	// Update the list of dirty values.
	a.insert(key, value, bt)

	return nil
}


// TryGet returns the value for a key stored in the trie.
func (a *AdmAccountBranch) TryGet(key []byte, bt *BinaryTrie) (interface{}, error) {
	bk := bt.keyCreationFn(key)
	off := 0
	var currentNode *AdmAccountBranch
	switch bt.root.(type) {
	case empty:
		return nil, errKeyNotFound
	case *AdmAccountBranch:
		currentNode = bt.root.(*AdmAccountBranch)
	case hashBinaryNode:
		return nil, errReadFromHash
	}

	for {
		// If the element doesn't have the same prefix we abort
		if !bk.SamePrefix(currentNode.prefix, off) {
			return nil, errKeyNotFound
		}

		// Only leaf nodes have values
		if currentNode.isLeaf() {
			return currentNode.Value(key), nil
		}

		// This node is a fork, get the child node
		var childNode *AdmAccountBranch
		if bk[off+len(currentNode.prefix)] == 0 {
			childNode = bt.resolveNodeFn(currentNode.left, bk, off+1).(*AdmAccountBranch)
		} else {
			childNode = bt.resolveNodeFn(currentNode.right, bk, off+1).(*AdmAccountBranch)
		}
		childNode.parent = &currentNode.branch

		// if no child node could be found, the key
		// isn't present in the trie.
		if childNode == nil {
			return nil, errKeyNotFound
		}
		off += len(currentNode.prefix) + 1
		currentNode = childNode
	}
}


// Update does the same thing as TryUpdate except it panics if it encounters
// an error.
func (a *AdmAccountBranch) Update(key, value []byte, bt *BinaryTrie) {
	if err := a.TryUpdate(key, value, bt); err != nil {
		fmt.Sprintf("Unhandled trie error: %v", err)
	}
}

// subTreeFromPath rebuilds the subtrie rooted at path `path` from the db.
func (a *AdmAccountBranch) subTreeFromPath(path binkey, bt *BinaryTrie) *AdmAccountBranch {
	subtrie := NewBinaryTrie(nil)
	kv := *bt.kvdb

	it := kv.Iterator(nil)
	for it.Next() {
		a.TryUpdate(it.Key(), it.Value(), bt)
	}
	rootbranch := subtrie.root.(*AdmAccountBranch)
	rootbranch.prefix = rootbranch.prefix[:len(path)]
	return rootbranch
}

func (a *AdmAccountBranch) resolveNode(childNode BinaryNode, bk binkey, off int, bt *BinaryTrie) *AdmAccountBranch {
	// Check if the node has already been resolved,
	// otherwise, resolve it.
	switch childNode := childNode.(type) {
	case empty:
		return nil
	case hashBinaryNode:
		if bytes.Equal(childNode[:], emptyRootSha512[:]) {
			return nil
		}
		return a.subTreeFromPath(bk[:off], bt)
	}
	return childNode.(*AdmAccountBranch)
}

func (a *AdmAccountBranch) insert(key, value []byte, bt *BinaryTrie) error {
	if len(key) != 32  {
		return errors.New("can only insert keys at depth 32")
	}

	// the only value associated to a key length of 512 bits is
	// if the subtree selector is 0b11.
	itemSelector := key[31] & 3
	if len(key) == 64 && itemSelector != 3 {
		return errors.New("bintrie: trying to write at an invalid depth")
	}

	var accountKey common.Hash
	copy(accountKey[:], key[:32])
	accountKey[31] &= 0xFC

	kv := *bt.kvdb
	accDirt := kv.(*AdmKVStorage)
	account, ok := accDirt.dirties[accountKey]
	if !ok {
		account = &AdmAccountDirties{
			balance: big.NewInt(0),
			dirties: make(map[common.Hash]common.Hash),
		}
		accDirt.dirties[accountKey] = account
	}

	switch itemSelector {
	case 0: // Balance
		account.balance.SetBytes(value)
	case 1: // nonce
		if len(value) > 8 {
			return errors.New("trying to write nonce larger than u68")
		}
		account.nonce = binary.BigEndian.Uint64(value)
	case 2: // code
		account.codeHash = common.BytesToHash(value)
	case 3:
		if len(value) != 32 {
			return errors.New("bintrie: invalid value length in slot write")
		}
		slotKey := common.BytesToHash(key[32:64])
		account.dirties[slotKey] = common.BytesToHash(value)
	default:
		return errors.New("bintrie: range reserved for future use")
	}

	account.flags |= (1 << itemSelector)

	return nil
}