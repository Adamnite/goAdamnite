package rawdb

import (
	"github.com/adamnite/go-adamnite/databaseDeprecated"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"

	encoding "github.com/vmihailenco/msgpack/v5"
	log "github.com/sirupsen/logrus"
)

func ReadHeaderHash(db adamnitedb.AdamniteDBReader, blockNum uint64) (common.Hash, error) {

	data, _ := db.Get(blockHeaderHashKey(blockNum))

	if len(data) == 0 {
		return common.Hash{}, nil
	}

	return common.BytesToHash(data), nil
}

func WriteTrieNode(db adamnitedb.AdamniteDBWriter, hash common.Hash, node []byte) {
	if err := db.Insert(hash[:], node); err != nil {
		log.Fatal("Failed to store trie node", "err", err)
	}
}

// ReadTrieNode retrieves the trie node of the provided hash.
func ReadTrieNode(db adamnitedb.AdamniteDBReader, hash common.Hash) []byte {
	data, _ := db.Get(hash[:])
	return data
}

func DeleteTrieNode(db adamnitedb.AdamniteDBWriter, hash common.Hash) {
	if err := db.Delete(hash[:]); err != nil {
		log.Fatal("Failed to delete trie node", "err", err)
	}
}

// ReadPreimage retrieves a single preimage of the provided hash.
func ReadPreimage(db adamnitedb.AdamniteDBReader, hash common.Hash) []byte {
	data, _ := db.Get(preimageKey(hash))
	return data
}

// WritePreimages writes the provided set of preimages to the database.
func WritePreimages(db adamnitedb.AdamniteDBWriter, preimages map[common.Hash][]byte) {
	for hash, preimage := range preimages {
		if err := db.Insert(preimageKey(hash), preimage); err != nil {
			log.Fatal("Failed to store trie preimage", "err", err)
		}
	}
}

func WriteEpochNumber(db adamnitedb.AdamniteDBWriter, epochNum uint64) {
	data, err := encoding.Marshal(epochNum)
	if err != nil {
		log.Fatal("Failed to encode epoch number", "err", err)
	}

	if err := db.Insert(epochKey(), data); err != nil {
		log.Fatal("Failed to store epoch number", "err", err)
	}
}

func ReadEpochNumber(db adamnitedb.AdamniteDBReader) uint64 {
	data, err := db.Get(epochKey())
	if err != nil {
		log.Fatal("Failed to get epoch number from store", "err", err)
	}

	if len(data) == 0 {
		return 0
	}

	var epochNum uint64

	if err := encoding.Unmarshal(data, &epochNum); err != nil {
		log.Fatal("Failed to decode epoch enc data", "err", err)
	}

	return epochNum
}

func WriteBlock(db adamnitedb.AdamniteDBWriter, block *types.Block) {
	WriteBody(db, block.Hash(), block.Numberu64(), block.Body())
	WriteHeader(db, block.Header())
}

func WriteBody(db adamnitedb.AdamniteDBWriter, hash common.Hash, blockNum uint64, body *types.Body) {

	data, err := encoding.Marshal(body)
	if err != nil {
		log.Fatal("Failed to encode body", "err", err)
	}
	WriteBodyMsgPack(db, hash, blockNum, data)
}

func WriteBodyMsgPack(db adamnitedb.AdamniteDBWriter, hash common.Hash, blockNum uint64, msgpack encoding.RawMessage) {
	if err := db.Insert(blockBodyKey(blockNum, hash), msgpack); err != nil {
		log.Fatal("Failed to store block body", "err", err)
	}
}

func WriteHeader(db adamnitedb.AdamniteDBWriter, header *types.BlockHeader) {
	var (
		hash   = header.Hash()
		number = header.Number.Uint64()
	)

	WriteHeaderNumber(db, hash, number)

	data, err := encoding.Marshal(header)
	if err != nil {
		log.Fatal("Failed to encode header", "err", err)
	}

	key := headerKey(number, hash)
	if err := db.Insert(key, data); err != nil {
		log.Fatal("Failed to store header", "err", err)
	}
}

func WriteHeaderNumber(db adamnitedb.AdamniteDBWriter, hash common.Hash, number uint64) {
	key := headerNumberKey(hash)
	enc := encodeBlockNumber(number)
	if err := db.Insert(key, enc); err != nil {
		log.Fatal("Failed to store hash to number mapping", "err", err)
	}
}
