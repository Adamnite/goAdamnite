package rawdb

import (
	"encoding/binary"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/metrics"
)

var (
	blockHeaderPrefix     = []byte("h")
	blockHeaderHashSuffix = []byte("H")

	epochPrefix = []byte("epoch") // epochPrefix -> epoch number

	blockBodyPrefix = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body

	headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)

	preimagePrefix = []byte("secure-key-") // preimagePrefix + hash -> preimage

	preimageCounter    = metrics.NewRegisteredCounter("db/preimage/total", nil)
	preimageHitCounter = metrics.NewRegisteredCounter("db/preimage/hits", nil)
)

// blockHeaderHashKey = blockHeaderPrefix + num + headerHashSuffix
func blockHeaderHashKey(number uint64) []byte {
	return append(append(blockHeaderPrefix, encodeBlockNumber(number)...), blockHeaderHashSuffix...)
}

func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.LittleEndian.PutUint64(enc, number)
	return enc
}

func epochKey() []byte {
	return epochPrefix
}

// blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash
func blockBodyKey(number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

func headerNumberKey(hash common.Hash) []byte {
	return append(headerNumberPrefix, hash.Bytes()...)
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}
