package leveldb

import (
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	log "github.com/sirupsen/logrus"
)

const (
	minCache   = 16 // minCache is the minimum amount of memory in megabytes to allocate to leveldb read and write chaching, split half and half
	minHandles = 16 // minHandles is the minimum number of files handles to allocate to the open database files.
)

type Database struct {
	fileName string      // filename for levelDB path
	db       *leveldb.DB // levelDB instance

	quitLock sync.Mutex
	quitChan chan chan error

	logger *log.Entry
}

// New returns a wrapped LevelDB object
func New(file string, cache int, handles int, readonly bool) (*Database, error) {
	if cache < minCache {
		cache = minCache
	}

	if handles < minHandles {
		handles = minHandles
	}

	options := &opt.Options{
		Filter:                 filter.NewBloomFilter(10),
		DisableSeeksCompaction: true,
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB,
		ReadOnly:               readonly,
	}

	logger := log.WithFields(log.Fields{"database": file})
	usedCache := options.GetBlockCacheCapacity() + options.GetWriteBuffer()*2

	logger.Info("Allocated cache and file handles", "cache", common.StorageSize(usedCache), "handles", options.GetOpenFilesCacheCapacity(), "readonly", options.ReadOnly)

	db, err := leveldb.OpenFile(file, options)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	levelDB := &Database{
		fileName: file,
		db:       db,
		logger:   logger,
		quitChan: make(chan chan error),
	}

	return levelDB, nil
}

// Close flushes any pending data to disk and closes all io accesses to the underlying key-value store.
func (db *Database) Close() error {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()
	return db.db.Close()
}

// Get retrieves the given key if it's present in the key-value store.
func (db *Database) Get(key []byte) ([]byte, error) {
	value, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Insert inserts the given value into the key-value store.
func (db *Database) Insert(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

// Delete removes the given value into the key-value store.
func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

func (db *Database) NewIterator(prefix []byte, start []byte) adamnitedb.Iterator {
	return db.db.NewIterator(bytesPrefixRange(prefix, start), nil)
}

func (db *Database) Stat(property string) (string, error) {
	return db.db.GetProperty(property)
}

func (db *Database) NewBatch() adamnitedb.Batch {
	return &batch{
		db: db.db,
		b:  new(leveldb.Batch),
	}
}

func bytesPrefixRange(prefix, start []byte) *util.Range {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	return r
}

// batch is a write-only leveldb batch that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
type batch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

// Put inserts the given value into the batch for later committing.
func (b *batch) Insert(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

// Delete inserts the a key removal into the batch for later committing.
func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *batch) Write() error {
	return b.db.Write(b.b, nil)
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

// Replay replays the batch contents.
func (b *batch) Replay(w adamnitedb.AdamniteDBWriter) error {
	return b.b.Replay(&replayer{writer: w})
}

// replayer is a small wrapper to implement the correct replay methods.
type replayer struct {
	writer  adamnitedb.AdamniteDBWriter
	failure error
}

// Put inserts the given value into the key-value data store.
func (r *replayer) Put(key, value []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Insert(key, value)
}

// Delete removes the key from the key-value data store.
func (r *replayer) Delete(key []byte) {
	// If the replay already failed, stop executing ops
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Delete(key)
}

func (db *Database) Compact(start []byte, limit []byte) error {
	return db.db.CompactRange(util.Range{Start: start, Limit: limit})
}
