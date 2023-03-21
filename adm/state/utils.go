package state

import (
	"bytes"
	"crypto/sha512"
	"hash"
	"sync"
)

type binkey []byte

// trieHasher is a type used for the trie Hash operation. A hasher has some
// internal preallocated temp space
type trieHasher struct {
	sha      hash.Hash
	tmp      []byte
	parallel bool
}


func (b binkey) SamePrefix(other binkey, off int) bool {
	return bytes.Equal(b[off:off+len(other)], other[:])
}
func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func (b binkey) commonLength(other binkey) int {
	length := min(len(b), len(other))
	for i := 0; i < length; i++ {
		if b[i] != other[i] {
			return i
		}
	}
	return length
}

var hashPoolSha512 = sync.Pool{
	New: func() interface{} {
		return &trieHasher{
			tmp: make([]byte, 0, 550),
			sha: sha512.New(),
		}
	},
}

func NewHasher512(parallel bool) *trieHasher {
	h := hashPoolSha512.Get().(*trieHasher)
	h.parallel = parallel
	return h
}

func GetHasher() *trieHasher {
	hasher := NewHasher512(false)
	hasher.sha.Reset()
	return hasher
}

func PutHasher(h *trieHasher) {
	hashPoolSha512.Put(h)
}

func newBinKey(key []byte) binkey {
	bits := make([]byte, 8*len(key))
	for i, kb := range key {
		// might be best to have this statement first, as compiler bounds-checking hint
		bits[8*i+7] = kb & 0x1
		bits[8*i] = (kb >> 7) & 0x1
		bits[8*i+1] = (kb >> 6) & 0x1
		bits[8*i+2] = (kb >> 5) & 0x1
		bits[8*i+3] = (kb >> 4) & 0x1
		bits[8*i+4] = (kb >> 3) & 0x1
		bits[8*i+5] = (kb >> 2) & 0x1
		bits[8*i+6] = (kb >> 1) & 0x1
	}
	return binkey(bits)
}