package adamnitedb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	// minLevelDBCache is the minimum memory allocate to leveldb
	// half write, half read
	minLevelDBCache = 16 // 16 MiB

	// minLevelDBHandles is the minimum number of files handles to leveldb open files
	minLevelDBHandles = 16

	DefaultLevelDBCache               = 1024 // 1 GiB
	DefaultLevelDBHandles             = 512  // files handles to leveldb open files
	DefaultLevelDBBloomKeyBits        = 2048 // bloom filter bits (256 bytes)
	DefaultLevelDBCompactionTableSize = 4    // 4  MiB
	DefaultLevelDBCompactionTotalSize = 40   // 40 MiB
	DefaultLevelDBNoSync              = false
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

type LevelDBFactory interface {
	// set cache size
	SetCacheSize(int) LevelDBFactory

	// set handles
	SetHandles(int) LevelDBFactory

	// set bloom key bits
	SetBloomKeyBits(int) LevelDBFactory

	// set compaction table size
	SetCompactionTableSize(int) LevelDBFactory

	// set compaction table total size
	SetCompactionTotalSize(int) LevelDBFactory

	// set no sync
	SetNoSync(bool) LevelDBFactory

	// build the storage
	Build() (KVBatchStorage, error)
}

type leveldbFactory struct {
	path    string
	options *opt.Options
}

func (factory *leveldbFactory) SetCacheSize(cacheSize int) LevelDBFactory {
	cacheSize = max(cacheSize, minLevelDBCache)

	factory.options.BlockCacheCapacity = cacheSize * opt.MiB

	return factory
}

func (factory *leveldbFactory) SetHandles(handles int) LevelDBFactory {
	factory.options.OpenFilesCacheCapacity = max(handles, minLevelDBHandles)
	return factory
}

func (factory *leveldbFactory) SetBloomKeyBits(bloomKeyBits int) LevelDBFactory {
	factory.options.Filter = filter.NewBloomFilter(bloomKeyBits)

	return factory
}

func (factory *leveldbFactory) SetCompactionTableSize(compactionTableSize int) LevelDBFactory {
	factory.options.CompactionTableSize = compactionTableSize * opt.MiB
	factory.options.WriteBuffer = factory.options.CompactionTableSize * 2

	return factory
}

func (factory *leveldbFactory) SetCompactionTotalSize(compactionTotalSize int) LevelDBFactory {
	factory.options.CompactionTotalSize = compactionTotalSize * opt.MiB

	return factory
}

func (factory *leveldbFactory) SetNoSync(noSync bool) LevelDBFactory {
	factory.options.NoSync = noSync
	return factory
}

func (factory *leveldbFactory) Build() (KVBatchStorage, error) {
	db, err := leveldb.OpenFile(factory.path, factory.options)
	if err != nil {
		return nil, err
	}

	return &levelDBKV{db: db}, nil
}

// NewFactory creates the new leveldb storage factory
func NewLevelDBFactory(path string) LevelDBFactory {
	return &leveldbFactory{
		path:   path,
		options: &opt.Options{
			OpenFilesCacheCapacity:        minLevelDBHandles,
			CompactionTableSize:           DefaultLevelDBCompactionTableSize * opt.MiB,
			CompactionTotalSize:           DefaultLevelDBCompactionTotalSize * opt.MiB,
			BlockCacheCapacity:            minLevelDBCache * opt.MiB,
			WriteBuffer:                   (DefaultLevelDBCompactionTableSize * 2) * opt.MiB,
			CompactionTableSizeMultiplier: 1.1,
			Filter:                        filter.NewBloomFilter(DefaultLevelDBBloomKeyBits),
			NoSync:                        false,
			BlockSize:                     256 * opt.KiB,
			FilterBaseLg:                  19,
			DisableSeeksCompaction:        true,
		},
	}
}
