package dpos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"

	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

const (
	blockInterval   = 16
	EpochBlockCount = 162
)

var (
	// When a list of signers for a block is requested, errunknownblock is returned.
	// This is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	//If the timestamp of the block is lower than errInvalidTimestamp
	//Timestamp of the previous block + minimum block period.
	ErrInvalidTimestamp         = errors.New("invalid timestamp")
	ErrWaitForPrevBlock         = errors.New("wait for last block arrived")
	ErrApplyNextBlock           = errors.New("apply the next block")
	ErrMismatchSignerAndWitness = errors.New("mismatch block signer and witness")
	ErrInvalidWitness           = errors.New("invalid witness")
	ErrInvalidApplyBlockTime    = errors.New("invalid time to apply the block")
	ErrNilBlockHeader           = errors.New("nil block header returned")
	ErrUnknownAncestor          = errors.New("unknown ancestor")
)
var (
	big0  = big.NewInt(0)
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)

	timeOfFirstBlock = int64(0)

	confirmedBlockHead = []byte("confirmed-block-head")
)

type Config struct {
	Log log15.Logger `toml:"-"`
}

type AdamniteDPOS struct {
	config    Config
	db        adamnitedb.Database
	lock      sync.Mutex
	closeOnce sync.Once
}

func New(config Config, db adamnitedb.Database) *AdamniteDPOS {
	if config.Log == nil {
		config.Log = log15.Root()
	}

	dpos := &AdamniteDPOS{
		config: config,
		db:     db,
	}

	return dpos
}

func (adpos *AdamniteDPOS) Close() error {
	adpos.closeOnce.Do(func() {

	})
	return nil
}

func (adpos *AdamniteDPOS) Witness(header *types.BlockHeader) (common.Address, error) {
	return header.Witness, nil
}

func (adpos *AdamniteDPOS) VerifyHeader(header *types.BlockHeader, chain ChainReader, blockInterval uint64) error {
	if header.Number == nil {
		return errUnknownBlock
	}

	number := header.Number.Uint64()

	parent := chain.GetHeader(header.ParentHash, number-1)

	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return ErrUnknownAncestor
	}

	if parent.Time+blockInterval > header.Time {
		return ErrInvalidTimestamp
	}
	return nil
}

func (adpos *AdamniteDPOS) Prepare(chain ChainReader, header *types.BlockHeader) error {

	number := header.Number.Uint64()

	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return ErrUnknownAncestor
	}
	witness, _ := adpos.Witness(header)
	header.Witness = witness
	return nil
}

func proceedIncentive(config *params.ChainConfig, state *statedb.StateDB, header *types.BlockHeader) {

}

func (adpos *AdamniteDPOS) Finalize(chain ChainReader, header *types.BlockHeader, state *statedb.StateDB, txs []*types.Transaction, dposEnv *types.DposEnv, witnessCandidatePool WitnessCandidatePool) (*types.Block, error) {
	proceedIncentive(chain.Config(), state, header)

	parent := chain.GetHeaderByHash(header.ParentHash)
	epochContext := &EpochContext{
		statedb:   state,
		DposEnv:   dposEnv,
		TimeStamp: int64(header.Time),
	}

	if timeOfFirstBlock == 0 {
		if firstBlockHeader := chain.GetHeaderByNumber(1); firstBlockHeader != nil {
			timeOfFirstBlock = int64(firstBlockHeader.Time)
		}
	}
	genesis := chain.GetHeaderByNumber(0)

	err := epochContext.tryElect(genesis, parent, witnessCandidatePool)
	if err != nil {
		return nil, fmt.Errorf("got error when elect next epoch, err: %s", err)
	}

	updateMintCnt(int64(parent.Time), int64(header.Time), header.Witness, dposEnv)
	header.DposEnv = dposEnv.ToProto()
	return types.NewBlock(header, txs, trie.NewStackTrie(nil)), nil
}

func PrevSlot(now int64, blockInterval uint64) int64 {
	return int64((now-1)/int64(blockInterval)) * int64(blockInterval)
}

func NextSlot(now int64, blockInterval uint64) int64 {
	return int64((now+int64(blockInterval)-1)/int64(blockInterval)) * int64(blockInterval)
}
func (adpos *AdamniteDPOS) GetRoundNumber() uint64 {
	return 0
}
func updateMintCnt(parentBlockTime, currentBlockTime int64, witness common.Address, dposEnv *types.DposEnv) {
	currentMintCntTrie := dposEnv.MintCntTrie()
	currentEpoch := parentBlockTime / EpochBlockCount
	currentEpochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentEpochBytes, uint64(currentEpoch))

	cnt := int64(1)
	newEpoch := currentBlockTime / EpochBlockCount

	if currentEpoch == newEpoch {
		iter := trie.NewIterator(currentMintCntTrie.NodeIterator(currentEpochBytes))

		if iter.Next() {
			cntBytes := currentMintCntTrie.Get(append(currentEpochBytes, witness.Bytes()...))

			if cntBytes != nil {
				cnt = int64(binary.BigEndian.Uint64(cntBytes)) + 1
			}
		}
	}

	newCntBytes := make([]byte, 8)
	newEpochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newEpochBytes, uint64(newEpoch))
	binary.BigEndian.PutUint64(newCntBytes, uint64(cnt))
	dposEnv.MintCntTrie().TryUpdate(append(newEpochBytes, witness.Bytes()...), newCntBytes)
}
