package rawdb

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/vmihailenco/msgpack/v5"
)

func ReadHeaderHash(db adamnitedb.AdamniteDBReader, blockNum uint64) (common.Hash, error) {

	data, _ := db.Get(blockHeaderHashKey(blockNum))

	if len(data) == 0 {
		return common.Hash{}, nil
	}

	return common.BytesToHash(data), nil
}

func WriteEpochNumber(db adamnitedb.AdamniteDBWriter, epochNum uint64) {
	data, err := msgpack.Marshal(epochNum)
	if err != nil {
		log15.Crit("Failed to encode epoch number", "err", err)
	}

	if err := db.Insert(epochKey(), data); err != nil {
		log15.Crit("Failed to store epoch number", "err", err)
	}
}

func WriteBlock(db adamnitedb.AdamniteDBWriter, block *types.Block) {
	WriteBody(db, block.Hash(), block.Numberu64(), block.Body())
	WriteHeader(db, block.Header())
}

func WriteBody(db adamnitedb.AdamniteDBWriter, hash common.Hash, blockNum uint64, body *types.Body) {

	data, err := msgpack.Marshal(body)
	if err != nil {
		log15.Crit("Failed to encode body", "err", err)
	}
	WriteBodyMsgPack(db, hash, blockNum, data)
}

func WriteBodyMsgPack(db adamnitedb.AdamniteDBWriter, hash common.Hash, blockNum uint64, msgpack msgpack.RawMessage) {
	if err := db.Insert(blockBodyKey(blockNum, hash), msgpack); err != nil {
		log15.Crit("Failed to store block body", "err", err)
	}
}

func WriteHeader(db adamnitedb.AdamniteDBWriter, header *types.BlockHeader) {
	var (
		hash   = header.Hash()
		number = header.Number.Uint64()
	)

	WriteHeaderNumber(db, hash, number)

	data, err := msgpack.Marshal(header)
	if err != nil {
		log15.Crit("Failed to encode header", "err", err)
	}

	key := headerKey(number, hash)
	if err := db.Insert(key, data); err != nil {
		log15.Crit("Failed to store header", "err", err)
	}
}

func WriteHeaderNumber(db adamnitedb.AdamniteDBWriter, hash common.Hash, number uint64) {
	key := headerNumberKey(hash)
	enc := encodeBlockNumber(number)
	if err := db.Insert(key, enc); err != nil {
		log15.Crit("Failed to store hash to number mapping", "err", err)
	}
}
