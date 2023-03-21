package state

import (
	"errors"
	"fmt"
	adamnitedb "github.com/adamnite/go-adamnite/adm/adamnitedb/leveldb"
	"github.com/adamnite/go-adamnite/common"
	lru "github.com/hashicorp/golang-lru"
)

// BinaryNode represents any node in a binary trie.
type BinaryNode interface {
	Hash() []byte
	hash(off int) []byte
	Commit() error
	isLeaf() bool
	Value([]byte) interface{}
	// save(path binkey) ([]byte, error)
}

// BinaryHashPreimage represents a tuple of a hash and its preimage
type BinaryHashPreimage struct {
	Key   []byte
	Value []byte
}

type hashType int
const leafThreshold = 2
var emptyRootSha512 = common.FromHex("cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e")

// BinaryTrie represents a multi-level binary trie.
//
// Nodes with only one child are compacted into a "prefix"
// for the first node that has two children.
type KeyCreation func([]byte) binkey
type ResolveNode func(childNode BinaryNode, bk binkey, off int) interface{}

type BinaryTrie struct {
	root     		BinaryNode
	kvdb     		*adamnitedb.KVBatchStorage
	// A cache of nodes with at least N children in memory
	cache 			*lru.Cache
	keyCreationFn 	KeyCreation
	resolveNodeFn	ResolveNode
}

type (
	branch struct {
		left  BinaryNode
		right BinaryNode

		key   []byte
		value []byte
		// Used to send (hash, preimage) pairs when hashing
		CommitCh chan BinaryHashPreimage

		prefix binkey
		childCount int

		parent *branch // pointer to parent, nil for root
		hashStoredData interface{}
	}

	hashBinaryNode []byte

	empty struct{}
)

var (
	errInsertIntoHash    = errors.New("trying to insert into a hash")
	errReadFromHash      = errors.New("trying to read from a hash")
	errReadFromEmptyTree = errors.New("reached an empty subtree")
	errKeyNotFound     	 = errors.New("key doesn't exist inside the tree")

	// 0_32
	zero32 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

func (br *branch) Hash() []byte {
	return br.hash(0)
}

func (br *branch) hash(off int) []byte {
	var hasher *trieHasher
	var hash []byte
	if br.value == nil {
		// This is a branch node, so the rule is
		// branch_hash = hash(left_root_hash || right_root_hash)
		lh := br.left.hash(off + len(br.prefix) + 1)
		rh := br.right.hash(off + len(br.prefix) + 1)
		hasher = GetHasher()
		defer PutHasher(hasher)
		hasher.sha.Write(lh)
		hasher.sha.Write(rh)
		hash = hasher.sha.Sum(nil)

		// If br.CommitCh isn't nil, then it has been
		// called by the eviction process. This won't
		// always be true, and some extra parameter
		// should be introduced FIXME
		// Ongoing eviction process: replace the child
		// subtries with their hashes.
		if br.CommitCh != nil {
			br.left = hashBinaryNode(lh)
			br.right = hashBinaryNode(rh)
		}
	} else {
		hasher = GetHasher()
		defer PutHasher(hasher)
		// This is a leaf node, so the hashing rule is
		// leaf_hash = hash(hash(key) || hash(leaf_value))
		hasher.sha.Write(br.key)
		kh := hasher.sha.Sum(nil)
		hasher.sha.Reset()

		hasher.sha.Write(br.value)
		hash = hasher.sha.Sum(nil)
		hasher.sha.Reset()

		hasher.sha.Write(kh)
		hasher.sha.Write(hash)
		hash = hasher.sha.Sum(nil)

		// Store the leaf
		if br.CommitCh != nil {
			br.CommitCh <- BinaryHashPreimage{Key: br.key, Value: br.value}
		}
	}

	if len(br.prefix) > 0 {
		hasher.sha.Reset()
		fpLen := len(br.prefix) + off
		hasher.sha.Write([]byte{byte(fpLen), byte(fpLen >> 8)})
		hasher.sha.Write(zero32[:30])
		hasher.sha.Write(hash)
		hash = hasher.sha.Sum(nil)
	}

	return hash
}


func (br *branch) isLeaf() bool {
	_, ok := br.left.(empty)
	return ok
}

// branchCacheEvictionCallback is called when the trie is taking too much
// memory and the oldest, deepest nodes must be evicted. It will calculate
// the hash of the subtrie, save it to disk.
func branchCacheEvictionCallback(bt *BinaryTrie) func(interface{}, interface{}) {
	return func(key interface{}, value interface{}) {
		br := value.(*branch)

		// Allocate channel to receive dirty values,
		// and store each received value to disk.
		oldChannel := br.CommitCh
		defer func() {
			br.CommitCh = oldChannel
		}()
		br.CommitCh = make(chan BinaryHashPreimage)
		br.Commit()
		go func() {
			for kv := range br.CommitCh {
				k := *bt.kvdb
				k.Set(kv.Key, kv.Value)
			}
		}()

		// Decrease child leaf counts in parents
		childCount := br.childCount
		for currentNode := br; currentNode.parent != nil; currentNode = currentNode.parent {
			currentNode.childCount -= childCount
		}
	}
}

// NewBinaryTrie creates a binary trie using sha512 for hashing.
func NewBinaryTrie(db *adamnitedb.KVBatchStorage) *BinaryTrie {
	bt := &BinaryTrie{
		root: empty(struct{}{}),
		keyCreationFn: newBinKey,
		kvdb: db,
	}
	bt.cache, _ = lru.NewWithEvict(100, branchCacheEvictionCallback(bt))
	return bt
}


// Save saves the top of the trie in the DB, which can then be loaded
// from the db when starting the client.
func (bt *BinaryTrie) Save() ([]byte, error) {
	//return bt.root.save(nil)
	panic("Not implemented yet")
}

// Load loads a serialized trie, which is typically stored in the DB
// when exiting geth.
func (bt *BinaryTrie) Load(serialized []byte) {
	keylen := serialized[0]
	path := serialized[1 : keylen+1]
	next := keylen + 33
	h := serialized[keylen+1 : next]
	root := &branch{
		prefix: path[:len(path)-1],
	}
	if path[len(path)-1] == 0 {
		root.left = hashBinaryNode(h)
	} else {
		panic("first hash should always be a left child")
	}

	for {
		keylen = serialized[next]
		// root.insertHash(serialized[next+1:next+1+keylen], serialized[next+1+keylen:next+33+keylen], 0)
		next = next + keylen + 33
	}
}

// Hash returns the root hash of the binary trie, with the merkelization
// rule described in EIP-3102.
func (bt *BinaryTrie) Hash() []byte {
	return bt.root.Hash()
}

// Commit stores all the values in the binary trie into the database.
// This version does not perform any caching, it is intended to perform
// the conversion from hexary to binary.
// It basically performs a hash, except that it makes sure that there is
// a channel to stream the intermediate (hash, preimage) values to.
func (br *branch) Commit() error {
	if br.CommitCh == nil {
		return fmt.Errorf("commit channel missing")
	}
	br.Hash()
	close(br.CommitCh)
	return nil
}

// Commit does not commit anything, because a hash doesn't have
// its accompanying preimage.
func (h hashBinaryNode) Commit() error {
	return nil
}

// Hash returns itself
func (h hashBinaryNode) Hash() []byte {
	return h
} 

func (h hashBinaryNode) hash(off int) []byte {
	return h
}

func (h hashBinaryNode) tryGet(key []byte, depth int) ([]byte, error) {
	if depth >= 8*len(key) {
		return []byte(h), nil
	}
	return nil, errReadFromEmptyTree
}

func (h hashBinaryNode) gv(path string) (string, string) {
	me := fmt.Sprintf("h%s", path)
	return fmt.Sprintf("%s [label=\"H\"]\n", me), me
}

func (h hashBinaryNode) isLeaf() bool                { panic("calling isLeaf on a hash node") }
func (h hashBinaryNode) Value([]byte) interface{}    { panic("calling value on a hash node") }
func (h hashBinaryNode) save(binkey) ([]byte, error) { panic("calling save on a hash node") }

func (e empty) Hash() []byte {
	return emptyRootSha512[:]
}

func (e empty) hash(off int) []byte {
	return emptyRootSha512[:]
}

func (e empty) Commit() error {
	return errors.New("can not commit empty node")
}

func (e empty) tryGet(key []byte, depth int) ([]byte, error) {
	return nil, errReadFromEmptyTree
}

func (e empty) isLeaf() bool                { panic("calling isLeaf on an empty node") }
func (e empty) Value([]byte) interface{}    { panic("calling value on an empty node") }
func (e empty) save(binkey) ([]byte, error) { panic("calling save on empty node") }

func (e empty) gv(path string) (string, string) {
	me := fmt.Sprintf("e%s", path)
	return fmt.Sprintf("%s [label=\"∅\"]\n", me), me
}