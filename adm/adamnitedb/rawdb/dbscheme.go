package rawdb

import (
	"encoding/binary"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/metrics"
)

var (
	epochPrefix            = []byte("epoch") // epochPrefix -> epoch number

	blockBodyPrefix        = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body

	headerPrefix           = []byte("h") 
	headerNumberPrefix     = []byte("HN") 
	headerHashSuffix       = []byte("BH")

	preimagePrefix         = []byte("secure-key-") // preimagePrefix + hash -> preimage

	currentBlockNumPrefix  = []byte("curBN")

	witnessListPrefix      = []byte("WL")
	witnessBlackListPrefix = []byte("BL")

	preimageCounter        = metrics.NewRegisteredCounter("db/preimage/total", nil)
	preimageHitCounter     = metrics.NewRegisteredCounter("db/preimage/hits", nil)
)

// headerHashKey = headerPrefix + num + headerHashSuffix
func headerHashKey(number uint64) []byte {
	return append(append(headerPrefix, encodeNumber(number)...), headerHashSuffix...)
}

func encodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.LittleEndian.PutUint64(enc, number)
	return enc
}

// preimageKey = preimagePrefix + hash
func preimageKey(hash common.Hash) []byte {
	return append(preimagePrefix, hash[:]...)
}

func epochKey() []byte {
	return epochPrefix
}

// blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash
func blockBodyKey(number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, encodeNumber(number)...), hash.Bytes()...)
}

func headerNumberKey(hash common.Hash) []byte {
	return append(headerNumberPrefix, hash.Bytes()...)
}

func currentBlockNumberKey() []byte {
	return currentBlockNumPrefix;
}

// headerKey returns the level db header key.
// headerKey = headerPrefix + block_num (uint64 big endian) + block_hash
// key => value format: headerKey => Block Header
// @params: 
//     number -> the block number
//     hash   -> the block hash 
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeNumber(number)...), hash.Bytes()...)
}

func witnessListKey(epochNum uint64) []byte {
	return append(witnessListPrefix, encodeNumber(epochNum)...)
}

func witnessBlackListKey(addr common.Address) []byte {
	return append(witnessBlackListPrefix, addr.Bytes()...)
}