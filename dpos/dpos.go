package dpos

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/adamnite/go-adamnite/accounts"
	"github.com/adamnite/go-adamnite/adm/adamnitedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"

	lru "github.com/hashicorp/golang-lru"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/params"
)

const (
	blockInterval       = 1
	EpochBlockCount     = 162
	inmemorySignatures  = 4096
	checkpointInterval  = 1024
	inmemoryWintessPool = 128
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
	ErrInvalidVotingChain       = errors.New("invalid voting chain")
)

type Config struct {
	Log log15.Logger `toml:"-"`
}

type AdamniteDPOS struct {
	config *params.ChainConfig
	db     adamnitedb.Database

	closeOnce sync.Once

	recents    *lru.ARCCache
	signatures *lru.ARCCache
}

type SignerFn func(accounts.Account, []byte) ([]byte, error)

func New(config *params.ChainConfig, db adamnitedb.Database) *AdamniteDPOS {

	signatures, _ := lru.NewARC(inmemorySignatures)
	recents, _ := lru.NewARC(inmemoryWintessPool)
	dpos := &AdamniteDPOS{
		config:     config,
		db:         db,
		recents:    recents,
		signatures: signatures,
	}

	return dpos
}

type DposData struct {
	Witnesses []types.Witness `json:"witnesses"`

	Votes map[common.Address]types.Voter `json:"votes"`
}

func (adpos *AdamniteDPOS) Close() error {
	adpos.closeOnce.Do(func() {

	})
	return nil
}

func ecrecover(header *types.BlockHeader, sigcache *lru.ARCCache) (common.Address, error) {

	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(common.Address), nil
	}

	signature := header.Extra

	// Recover the public key and the Adamnite address
	pubkey, err := crypto.Recover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}

	signer := crypto.PubkeyByteToAddress(pubkey)

	sigcache.Add(hash, signer)
	return signer, nil
}

func (adpos *AdamniteDPOS) witnesspool(chain ChainReader, number uint64, hash common.Hash, parents []*types.BlockHeader) (*WitnessPool, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers     []*types.BlockHeader
		witnessPool *WitnessPool
	)
	for witnessPool == nil {

		if number%checkpointInterval == 0 {
			if s, err := GetWitnessPool(adpos.config, adpos.signatures, adpos.db, hash); err == nil {
				log15.Info("Loaded voting snapshot from disk", "number", number, "hash", hash)
				witnessPool = s
				break
			}
		}

		if number == 0 {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()

				dposData := DposData{}
				dposDataBytes := checkpoint.Extra
				DposDataDecode(dposDataBytes, &dposData)

				tmpWitnesses := make([]types.Witness, 0)
				dposData.Witnesses = tmpWitnesses

				witnessPool = NewRoundWitnessPool(DefaultWitnessConfig, adpos.config, adpos.signatures, number, hash, dposData.Witnesses)
				if err := witnessPool.saveWitnessPool(adpos.db); err != nil {
					return nil, err
				}
				log15.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}

		var header *types.BlockHeader
		if len(parents) > 0 {
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			header = chain.GetHeader(hash, number)

			if header == nil {
				return nil, ErrUnknownAncestor
			}
		}

		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}

	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}

	witnessPool, err := witnessPool.witnessPoolFromBlockHeader(headers)
	if err != nil {
		return nil, err
	}
	adpos.recents.Add(witnessPool.Hash, witnessPool)

	if witnessPool.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = witnessPool.saveWitnessPool(adpos.db); err != nil {
			return nil, err
		}
		log15.Info("Stored voting snapshot to disk", "number", witnessPool.Number, "hash", witnessPool.Hash)
	}
	return witnessPool, err
}

func SealHash(header *types.BlockHeader) (hash common.Hash) {
	hasher := crypto.NewRipemd160State()
	encodeSignHeader(hasher, header)
	hasher.Sum(hash[:0])
	return hash
}

func encodeSignHeader(w io.Writer, header *types.BlockHeader) {
	err := msgpack.NewEncoder(w).Encode([]interface{}{
		header.ParentHash,
		header.Witness,
		header.WitnessRoot,
		header.CurrentEpoch,
		header.Number,
		header.Signature,
		header.StateRoot,
		header.Extra,
		header.Time,
		header.TransactionRoot,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
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

	if number == 0 {
		return nil
	}

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

	parent := chain.GetHeaderByHash(header.ParentHash)
	if parent == nil {
		return ErrUnknownAncestor
	}
	witnesspool, err := adpos.witnesspool(chain, number-1, parent.Hash(), nil)
	if err != nil {
		return err
	}

	dposData := DposData{
		Witnesses: []types.Witness{},
		Votes:     map[common.Address]types.Voter{},
	}
	if number%EpochBlockCount == 0 {
		dposData.Witnesses = witnesspool.CalcWitnesses()
	}
	header.Extra = append(header.Extra, DposDataEncode(dposData)...)

	return nil
}

func proceedIncentive(config *params.ChainConfig, state *statedb.StateDB, header *types.BlockHeader) {

}

func (adpos *AdamniteDPOS) Finalize(chain ChainReader, header *types.BlockHeader, state *statedb.StateDB, txs []*types.Transaction) (*types.Block, error) {
	proceedIncentive(chain.Config(), state, header)

	dposData := DposData{}
	err := DposDataDecode(header.Extra, &dposData)
	if err != nil {
		return nil, err
	}
	dposData.Votes = adpos.calVote(chain, header, state, txs)
	header.Extra = append(header.Extra, DposDataEncode(dposData)...)

	return types.NewBlock(header, txs, trie.NewStackTrie(nil)), nil
}

func (adpos *AdamniteDPOS) calVote(chain ChainReader, header *types.BlockHeader, state *statedb.StateDB, txs []*types.Transaction) (votes map[common.Address]types.Voter) {
	votes = map[common.Address]types.Voter{}

	number := header.Number.Uint64()
	var witnessPool *WitnessPool
	if number > 0 {
		witnessPool, _ = adpos.witnesspool(chain, number-1, header.ParentHash, nil)
		if witnessPool == nil {
			return
		}
	}
walk:
	for _, tx := range txs {

		sender, _ := types.Sender(types.AdamniteSigner{}, tx)

		vote := types.Voter{}
		switch tx.Type() {
		case types.VOTE_TX:

			if number != 0 && witnessPool.isVoted(sender) {
				continue walk
			}
			stakingAmount, ok := big.NewInt(0).SetString(tx.Amount().String(), 10)
			if !ok || stakingAmount.Cmp(big.NewInt(0)) < 0 {
				continue walk
			}
			vote.StakingAmount = stakingAmount
			if state.GetBalance(sender).Cmp(stakingAmount) <= 0 {
				continue walk
			}
			vote.Address = *tx.To()
			votes[sender] = vote
			log15.Info(fmt.Sprintf("vote from: %s, to: %s, stake amount: %s", sender.String(), vote.Address.String(), vote.StakingAmount.String()))
			state.SubBalance(sender, stakingAmount)

		default:
			continue walk
		}

	}
	return
}

func (adpos *AdamniteDPOS) GetRoundNumber() uint64 {
	return 0
}

func DposDataEncode(dpos DposData) []byte {
	bytes, _ := msgpack.Marshal(dpos)
	return bytes
}
func DposDataDecode(bytes []byte, data *DposData) error {
	return msgpack.Unmarshal(bytes, &data)
}
