package state

import (
	"math/big"

	adamnitedb "github.com/adamnite/go-adamnite/adm/adamnitedb/leveldb"
	"github.com/adamnite/go-adamnite/common"
)


type AdmAccountDirties struct {
	flags   		byte
	balance 		*big.Int
	nonce   		uint64
	codeHash  		common.Hash
	storageHash 	common.Hash
	// list of dirty slots. The key to the slot
	// needs to be hashed.
	dirties map[common.Hash]common.Hash
}


type AdmKVStorage struct {
	adamnitedb.KVBatchStorage
	dirties map[common.Hash] *AdmAccountDirties
}

