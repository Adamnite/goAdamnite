package admnode

import (
	"bytes"
	"encoding/binary"
	"os"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const (
	dbNodeKeyPrefix = "gn:"
)

// NodeDB is the node database, storing previously seen live nodes
type NodeDB struct {
	levelDB   *leveldb.DB
	isRunning sync.Once
	quit      chan struct{}
}

func OpenDB(path string) (*NodeDB, error) {
	if path == "" {
		return newMemoryNodeDB()
	}
	return newDiskNodeDB(path)
}

func newMemoryNodeDB() (*NodeDB, error) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return nil, err
	}

	return &NodeDB{levelDB: db, quit: make(chan struct{})}, nil
}

func newDiskNodeDB(path string) (*NodeDB, error) {
	options := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(path, options)
	if _, isCorrupted := err.(*errors.ErrCorrupted); isCorrupted {
		db, err = leveldb.RecoverFile(path, nil)
	}
	if err != nil {
		return nil, err
	}

	curVersion := make([]byte, binary.MaxVarintLen64)
	curVersion = curVersion[:binary.PutVarint(curVersion, int64(dbVersion))]

	// Check the database version if exist
	byVersion, err := db.Get([]byte(dbVersionKey), nil)
	if err == nil {
		if !bytes.Equal(byVersion, curVersion) {
			db.Close()
			if err = os.RemoveAll(path); err != nil {
				return nil, err
			}
			return newDiskNodeDB(path)
		}
	} else if err == leveldb.ErrNotFound {
		if err := db.Put([]byte(dbVersionKey), curVersion, nil); err != nil {
			db.Close()
			return nil, err
		}
	}

	return &NodeDB{levelDB: db, quit: make(chan struct{})}, nil
}

// QueryNodes retrieves random nodes to be used as bootstrap node.
func (db *NodeDB) QueryRandomNodes(count int, maxAge time.Duration) []*GossipNode {
	// now := time.Now()
	// nodes := make([]*GossipNode, 0, count)
	// iter := db.levelDB.NewIterator(nil, nil)
	// var id NodeID
	// defer iter.Release()

	// for seeks := 0; len(nodes) < count && seeks < count*5; seeks++ {
	// 	rand.Read(id[:])
	// 	iter.Seek(getNodeKey(id))

	// }
	return nil
}

func getNodeKey(id NodeID) []byte {
	key := append([]byte(dbNodeKeyPrefix), id[:]...)
	return key
}

func isValidNode(id, data []byte) *GossipNode {
	// node := new(GossipNode)
	return nil
}
