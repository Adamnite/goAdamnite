package rawdb

import (
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"

	"github.com/adamnite/go-adamnite/log15"
	"github.com/vmihailenco/msgpack/v5"
)

/*
|-------------------------------------------------------------------------------------------
| DB Writers
|-------------------------------------------------------------------------------------------
*/

// WriteHeaderHash writes header hash on database.
func WriteHeaderHash(db adamnitedb.AdamniteDBWriter, header *types.BlockHeader) {
	blockHeaderHash := header.Hash()
	if err := db.Insert(headerHashKey(header.Number.Uint64()), blockHeaderHash[:]); err != nil {
		log15.Crit("Failed to store block header hash", "err", err)
	}
}

func WriteTrieNode(db adamnitedb.AdamniteDBWriter, hash common.Hash, node []byte) {
	if err := db.Insert(hash[:], node); err != nil {
		log15.Crit("Failed to store trie node", "err", err)
	}
}

// WritePreimages writes the provided set of preimages to the database.
func WritePreimages(db adamnitedb.AdamniteDBWriter, preimages map[common.Hash][]byte) {
	for hash, preimage := range preimages {
		if err := db.Insert(preimageKey(hash), preimage); err != nil {
			log15.Crit("Failed to store trie preimage", "err", err)
		}
	}
	preimageCounter.Inc(int64(len(preimages)))
	preimageHitCounter.Inc(int64(len(preimages)))
}

// WriteEpochNumber writes the epoch number of the current blockchain.
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
	enc := encodeNumber(number)
	if err := db.Insert(key, enc); err != nil {
		log15.Crit("Failed to store hash to number mapping", "err", err)
	}
}

func WriteCurrentBlockNumber(db adamnitedb.AdamniteDBWriter, number uint64) {
	key := currentBlockNumberKey()
	
	val, err := msgpack.Marshal(number)
	if err != nil {
		log15.Crit("Failed to encode current block number", "err", err)
	}

	if err := db.Insert(key, val); err != nil {
		log15.Crit("Failed to change current block number", "err", err)
	}
}

func WriteWitnessList(db adamnitedb.AdamniteDBWriter, epochNum uint64, witnessList types.WitnessList) {
	key := witnessListKey(epochNum)

	val, err := msgpack.Marshal(witnessList)
	if err != nil {
		log15.Crit("Failed to encode witness list", "err", err)
	}

	if err := db.Insert(key, val); err != nil {
		log15.Crit("Failed to insert witness list on db", "err", err)
	}
}

func WriteWitnessBlackList(db adamnitedb.AdamniteDBWriter, blockNum uint64, wAddr common.Address) {
	key := witnessBlackListKey(wAddr)

	val, err := msgpack.Marshal(blockNum)
	if err != nil {
		log15.Crit("Failed to encode block number of invalid witness address", "err", err)
	}

	if err := db.Insert(key, val); err != nil {
		log15.Crit("Failed to insert witness address on BL", "err", err)
	}
}

/*
|-------------------------------------------------------------------------------------------
| DB Readers
|-------------------------------------------------------------------------------------------
*/

// ReadHeaderHash reads header hash of the block number from database.
func ReadHeaderHash(db adamnitedb.AdamniteDBReader, blockNum uint64) (common.Hash, error) {

	data, _ := db.Get(headerHashKey(blockNum))

	if len(data) == 0 {
		return common.Hash{}, ErrNoData
	}

	return common.BytesToHash(data), nil
}

// ReadTrieNode retrieves the trie node of the provided hash.
func ReadTrieNode(db adamnitedb.AdamniteDBReader, hash common.Hash) []byte {
	data, _ := db.Get(hash[:])
	return data
}

// ReadPreimage retrieves a single preimage of the provided hash.
func ReadPreimage(db adamnitedb.AdamniteDBReader, hash common.Hash) []byte {
	data, _ := db.Get(preimageKey(hash))
	return data
}

// ReadEpochNumber reads the current epoch number of the blockchain.
func ReadEpochNumber(db adamnitedb.AdamniteDBReader) uint64 {
	data, err := db.Get(epochKey())
	if err != nil {
		log15.Crit("Failed to get epoch number from store", "err", err)
	}

	if len(data) == 0 {
		return 0
	}

	var epochNum uint64

	if err := msgpack.Unmarshal(data, &epochNum); err != nil {
		log15.Crit("Failed to decode epoch enc data", "err", err)
	}

	return epochNum
}

func ReadBlock(db adamnitedb.AdamniteDBReader, blockNum uint64, blockHash common.Hash) (*types.Block, error) {
	header, err := ReadHeader(db, blockNum, blockHash)
	if err != nil {
		return nil, err
	}

	body, err := ReadBody(db, blockNum, blockHash) 
	if err != nil {
		return nil, err
	}

	block := types.NewBlock(header, body.Transactions, trie.NewStackTrie(nil))
	return block, nil
}

func ReadBlockFromHash(db adamnitedb.AdamniteDBReader, blockHash common.Hash) (*types.Block, error) {
	blockNum, err := ReadHeaderNumber(db, blockHash)
	if err != nil {
		return nil, err
	}

	return ReadBlock(db, blockNum, blockHash)
}

func ReadBlockFromNumber(db adamnitedb.AdamniteDBReader, blockNum uint64) (*types.Block, error) {
	blockHash, err := ReadHeaderHash(db, blockNum)
	if err != nil {
		return nil, err
	}

	return ReadBlock(db, blockNum, blockHash)
}

func ReadBody(db adamnitedb.AdamniteDBReader, blockNum uint64, blockHash common.Hash) (*types.Body, error) {
	packedData, err := ReadBodyMsgPack(db, blockNum, blockHash)
	if err != nil {
		return nil, err
	}

	body := new(types.Body)

	if err := msgpack.Unmarshal(packedData, body); err != nil {
		return nil, ErrMsgPackDecode
	}
	return body, nil
}

func ReadBodyMsgPack(db adamnitedb.AdamniteDBReader, blockNum uint64, blockHash common.Hash) (msgpack.RawMessage, error) {
	data, err := db.Get(blockBodyKey(blockNum, blockHash))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ReadHeader(db adamnitedb.AdamniteDBReader, blockNum uint64, blockHash common.Hash) (*types.BlockHeader, error) {
	headerKey := headerKey(blockNum, blockHash)
	data, err := db.Get(headerKey)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrNoData
	}

	header := new(types.BlockHeader)

	if err := msgpack.Unmarshal(data, header); err != nil {
		return nil, ErrMsgPackDecode
	}
	return header, nil
}

func ReadHeaderFromHash(db adamnitedb.AdamniteDBReader, blockHash common.Hash) (*types.BlockHeader, error) {
	blockNum, err := ReadHeaderNumber(db, blockHash)
	if err != nil {
		return nil, err
	}

	return ReadHeader(db, blockNum, blockHash)
}

func ReadHeaderFromNumber(db adamnitedb.AdamniteDBReader, blockNum uint64) (*types.BlockHeader, error) {
	blockHash, err := ReadHeaderHash(db, blockNum)
	if err != nil {
		return nil, err
	}

	return ReadHeader(db, blockNum, blockHash)
}

func ReadHeaderNumber(db adamnitedb.AdamniteDBReader, hash common.Hash) (uint64, error) {
	key := headerNumberKey(hash)
	data, err := db.Get(key)

	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, ErrNoData
	}

	var blockNum uint64

	if err := msgpack.Unmarshal(data, &blockNum); err != nil {
		return 0, ErrMsgPackDecode
	}

	return blockNum, nil
}

func ReadCurrentBlockNumber(db adamnitedb.AdamniteDBReader) (uint64, error) {
	key := currentBlockNumberKey()

	data, err := db.Get(key)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, ErrNoData
	}

	var currentBlockNum uint64

	if err := msgpack.Unmarshal(data, &currentBlockNum); err != nil {
		return 0, ErrMsgPackDecode
	}
	return currentBlockNum, nil
}

func ReadWitnessList(db adamnitedb.AdamniteDBReader, epochNum uint64) (types.WitnessList, error) {
	key := witnessListKey(epochNum)

	data, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrNoData
	}

	var witnessList []types.WitnessImpl

	if err := msgpack.Unmarshal(data, &witnessList); err != nil {
		return nil, ErrMsgPackDecode
	}

	var witnesses types.WitnessList

	for _, witness := range witnessList {
		witnesses = append(witnesses, witness)
	}
	return witnesses, nil
}

func ReadWitnessBlackList(db adamnitedb.AdamniteDBReader) ([]common.Address, error) {
	var blackList []common.Address

	return blackList, nil
}

/*
|-------------------------------------------------------------------------------------------
| DB Deletes
|-------------------------------------------------------------------------------------------
*/

func DeleteTrieNode(db adamnitedb.AdamniteDBWriter, hash common.Hash) {
	if err := db.Delete(hash[:]); err != nil {
		log15.Crit("Failed to delete trie node", "err", err)
	}
}


// TODO: 