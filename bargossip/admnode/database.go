package admnode

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/adamnite/go-adamnite/log15"
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
	now := time.Now()
	nodes := make([]*GossipNode, 0, count)
	iter := db.levelDB.NewIterator(nil, nil)
	var id NodeID
	var err error

	defer iter.Release()

seek:
	for seeks := 0; len(nodes) < count && seeks < count*5; seeks++ {
		rand.Read(id[:])
		iter.Seek(getNodeKey(id))

		var node *GossipNode
		if iter.Next() {
			node, err = isValidNode(iter.Key(), iter.Value())
			if err != nil {
				continue seek
			}
		} else {
			continue seek
		}

		if now.Sub(db.PongReceived(*node.ID(), node.IP())) > maxAge {
			continue seek
		}

		for i := range nodes {
			if *nodes[i].ID() == *node.ID() {
				continue seek
			}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// PongReceived returns the last pong received time from node.
func (db *NodeDB) PongReceived(id NodeID, ip net.IP) time.Time {
	if ip = ip.To16(); ip == nil {
		return time.Time{}
	}

	return time.Unix(db.getInt64(getNodePongTimeKey(id, ip)), 0)
}

func (db *NodeDB) UpdatePongReceived(id NodeID, ip net.IP, receivedTime time.Time) error {
	if ip = ip.To16(); ip == nil {
		return errors.New("invalid IP")
	}
	return db.storeInt64(getNodePongTimeKey(id, ip), receivedTime.Unix())
}

func (db *NodeDB) UpdateFindFails(id NodeID, ip net.IP, failed int) error {
	if ip = ip.To16(); ip == nil {
		return errors.New("invalid IP")
	}
	return db.storeInt64(getNodeItemKey(id, ip, []byte(dbNodeFindFailPrefix)), int64(failed))
}

func (db *NodeDB) FindFails(id NodeID, ip net.IP) int {
	if ip = ip.To16(); ip == nil {
		return 0
	}
	return int(db.getInt64(getNodeItemKey(id, ip, []byte(dbNodeFindFailPrefix))))
}

func (db *NodeDB) startCleanupExpiers() {
	db.isRunning.Do(func() {
		go db.runCleanup()
	})
}

func (db *NodeDB) runCleanup() {
	interval := time.NewTicker(dbCleanupDuration)
	defer interval.Stop()

	for {
		select {
		case <-interval.C:
			db.removeExpireNodes()
		case <-db.quit:
			return
		}
	}
}

// removeExpireNodes deletes all nodes that have not been seen for some time.
func (db *NodeDB) removeExpireNodes() {
	iter := db.levelDB.NewIterator(util.BytesPrefix([]byte(dbNodeKeyPrefix)), nil)
	defer iter.Release()

	var expireStart = time.Now().Add(-dbNodeExpireTime).Unix()

	for iter.Next() {
		id, ip, field := getNodeKeyItems(iter.Key())
		if field == dbNodePongPrefix {
			pongTime, _ := binary.Varint(iter.Value())
			if pongTime < expireStart {
				db.deleteRange(getNodeItemKey(id, ip, []byte("")))
			}
		}
	}
}

func (db *NodeDB) Close() {
	close(db.quit)
	db.levelDB.Close()
}

// getNodeKey generates the NodeDB key of GossipNode information.
func getNodeKey(id NodeID) []byte {
	key := append([]byte(dbNodeKeyPrefix), id[:]...)
	return key
}

func getNodePongTimeKey(id NodeID, ip net.IP) []byte {
	return getNodeItemKey(id, ip, []byte(dbNodePongPrefix))
}

func getNodeItemKey(id NodeID, ip net.IP, field []byte) []byte {
	ip16 := ip.To16()
	if ip16 == nil {
		panic(fmt.Errorf("GossipNodeDB: invalid IP"))
	}

	key := bytes.Join([][]byte{getNodeKey(id), ip16, field}, []byte{':'})
	return key
}

// getNodeKeyItems get the id, ip and field from key
// gn:{ID}
// gn:{ID}:{IP}:{FIELD} => len(ID) = 32, len(IP) = 16
func getNodeKeyItems(key []byte) (id NodeID, ip net.IP, field string) {
	if !bytes.HasPrefix(key, []byte(dbNodeKeyPrefix)) {
		return NodeID{}, nil, ""
	}

	item := key[len(dbNodeKeyPrefix):] // {ID}:{IP}:{FIELD}
	copy(id[:], item[:len(id)])

	if len(item) == len(id) {
		return id, nil, ""
	}
	item = item[len(id)+1:] // {IP}:{FIELD}
	if len(item) == 0 {
		return id, nil, ""
	}
	ip = item[:16]
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	if len(item) == 0 {
		return id, ip, ""
	}
	item = item[16+1:] // {FIELD}
	field = string(item)
	return id, ip, field
}

func isValidNode(id, data []byte) (*GossipNode, error) {
	node := new(GossipNode)
	if err := msgpack.Unmarshal(data, &node); err != nil {
		log15.Error("can't decode p2p node in DB", "err", err)
		return nil, err
	}

	if bytes.Equal(id, node.id[:]) {
		log15.Error("node db id mismatch", "id", id)
		return nil, ErrDBIdMismatch
	}
	return node, nil
}

func (db *NodeDB) Node(id NodeID) *GossipNode {
	data, err := db.levelDB.Get(getNodeKey(id), nil)
	if err != nil {
		return nil
	}
	return parseDecodeNode(id[:], data)
}

// UpdateNode inserts - potentially overwriting - a node into the peer database.
func (db *NodeDB) UpdateNode(node *GossipNode) error {
	key := getNodeKey(*node.ID())
	
	val, err := msgpack.Marshal(node)
	if err != nil {
		log15.Error("Failed to encode witness list", "err", err)
		return err
	}

	if err := db.levelDB.Put(key, val, nil); err != nil {
		log15.Error("Failed to insert witness list on db", "err", err)
		return err
	}
	return nil
}

func parseDecodeNode(id, data []byte) *GossipNode {
	node := new(GossipNode)
	if err := msgpack.Unmarshal(data, &node); err != nil {
		panic(fmt.Errorf("parse admnode db failed: %v", err))
	}

	copy(node.id[:], id)
	return node
}
