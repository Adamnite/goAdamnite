package adamnitedb

import "io"

const IdealBatchSize = 100 * 1024

type Batch interface {
	AdamniteDBWriter

	// ValuSize retrieves the amount of data queued up for writing.
	ValueSize() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch for reuse.
	Reset()

	// Replay replays the batch contents.
	Replay(w AdamniteDBWriter) error
}

type Batcher interface {
	NewBatch() Batch
}

type AdamniteDBReader interface {
	// Get retrieves the given key if it's present in the key-value data store.
	Get(key []byte) ([]byte, error)
}

type AdamniteDBWriter interface {
	// Insert inserts the given value into the key-value data store.
	Insert(key []byte, value []byte) error

	// Delete removes the key from the key-value store.
	Delete(key []byte) error
}

type Iterator interface {
	// Next moves the iterator to the next key-value pair.
	Next() bool

	// Error returns occured errors.
	Error() error

	// Key returns the key of the current key-value pair
	Key() []byte

	// Value returns the value of the current key-value pair
	Value() []byte

	// Release releases resources.
	Release()
}

type Iteratee interface {
	NewIterator(prefix []byte, start []byte) Iterator
}

type Stater interface {
	// Stat returns a particular internal stat of the database.
	Stat(property string) (string, error)
}

// Compacter wraps the Compact method of a backing data store.
type Compacter interface {
	// Compact flattens the underlying data store for the given key range. In essence,
	// deleted and overwritten versions are discarded, and the data is rearranged to
	// reduce the cost of operations needed to access them.
	//
	// A nil start is treated as a key before all keys in the data store; a nil limit
	// is treated as a key after all keys in the data store. If both is nil then it
	// will compact entire data store.
	Compact(start []byte, limit []byte) error
}

type Database interface {
	AdamniteDBReader
	AdamniteDBWriter
	Iteratee
	Batcher
	Stater
	Compacter
	io.Closer
}

type HookedBatch struct {
	Batch

	OnPut    func(key []byte, value []byte) // Callback if a key is inserted
	OnDelete func(key []byte)               // Callback if a key is deleted
}

// Put inserts the given value into the key-value data store.
func (b HookedBatch) Put(key []byte, value []byte) error {
	if b.OnPut != nil {
		b.OnPut(key, value)
	}
	return b.Batch.Insert(key, value)
}

// Delete removes the key from the key-value data store.
func (b HookedBatch) Delete(key []byte) error {
	if b.OnDelete != nil {
		b.OnDelete(key)
	}
	return b.Batch.Delete(key)
}
