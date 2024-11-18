package admnode

import (
	"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *NodeDB) storeInt64(key []byte, v int64) error {
	bin := make([]byte, binary.MaxVarintLen64)
	bin = bin[:binary.PutVarint(bin, v)]
	return db.levelDB.Put(key, bin, nil)
}

func (db *NodeDB) getInt64(key []byte) int64 {
	bin, err := db.levelDB.Get(key, nil)
	if err != nil {
		return 0
	}
	v, r := binary.Varint(bin)
	if r <= 0 {
		return 0
	}
	return v
}

func (db *NodeDB) storeUInt64(key []byte, v uint64) error {
	bin := make([]byte, binary.MaxVarintLen64)
	bin = bin[:binary.PutUvarint(bin, v)]
	return db.levelDB.Put(key, bin, nil)
}

func (db *NodeDB) getUInt64(key []byte) uint64 {
	bin, err := db.levelDB.Get(key, nil)
	if err != nil {
		return 0
	}
	v, r := binary.Uvarint(bin)
	if r <= 0 {
		return 0
	}
	return v
}

func (db *NodeDB) deleteRange(prefix []byte) {
	iter := db.levelDB.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()

	for iter.Next() {
		db.levelDB.Delete(iter.Key(), nil)
	}
}
