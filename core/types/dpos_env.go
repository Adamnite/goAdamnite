package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/trie"
	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/sha3"
)

type DposEnv struct {
	epochTrie            *trie.Trie
	witnessTrie          *trie.Trie
	voteTrie             *trie.Trie
	witnessCandidateTrie *trie.Trie
	mintCntTrie          *trie.Trie

	db *trie.Database
}

var (
	epochPrefix            = []byte("epoch-")
	witnessPrefix          = []byte("witness-")
	votePrefix             = []byte("vote-")
	witnessCandidatePrefix = []byte("candidate-")
	mintCntPrefix          = []byte("mintCnt-")
)

func NewEpochTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, epochPrefix, db)
}

func NewWitnessTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, witnessPrefix, db)
}

func NewVoteTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, votePrefix, db)
}

func NewWitnessCandidateTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, witnessCandidatePrefix, db)
}

func NewMintCntTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, mintCntPrefix, db)
}

func NewDposContext(db *trie.Database) (*DposEnv, error) {
	epochTrie, err := NewEpochTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	witnessTrie, err := NewWitnessTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	voteTrie, err := NewVoteTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	witnessCandidateTrie, err := NewWitnessCandidateTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	return &DposEnv{
		epochTrie:            epochTrie,
		witnessTrie:          witnessTrie,
		voteTrie:             voteTrie,
		witnessCandidateTrie: witnessCandidateTrie,
		mintCntTrie:          mintCntTrie,
		db:                   db,
	}, nil
}

func NewDposEnv(db *trie.Database, ctxProto *DposEnvProtocol) (*DposEnv, error) {
	epochTrie, err := NewEpochTrie(ctxProto.EpochHash, db)
	if err != nil {
		return nil, err
	}
	witnessTrie, err := NewWitnessTrie(ctxProto.WitnessHash, db)
	if err != nil {
		return nil, err
	}
	voteTrie, err := NewVoteTrie(ctxProto.VoteHash, db)
	if err != nil {
		return nil, err
	}
	witnessCandidateTrie, err := NewWitnessCandidateTrie(ctxProto.WitnessCandidateHash, db)
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(ctxProto.MintCntHash, db)
	if err != nil {
		return nil, err
	}
	return &DposEnv{
		epochTrie:            epochTrie,
		witnessTrie:          witnessTrie,
		voteTrie:             voteTrie,
		witnessCandidateTrie: witnessCandidateTrie,
		mintCntTrie:          mintCntTrie,
		db:                   db,
	}, nil
}

func (d *DposEnv) Copy() *DposEnv {
	epochTrie := *d.epochTrie
	witnessTrie := *d.witnessTrie
	voteTrie := *d.voteTrie
	witnessCandidateTrie := *d.witnessCandidateTrie
	mintCntTrie := *d.mintCntTrie
	return &DposEnv{
		epochTrie:            &epochTrie,
		witnessTrie:          &witnessTrie,
		voteTrie:             &voteTrie,
		witnessCandidateTrie: &witnessCandidateTrie,
		mintCntTrie:          &mintCntTrie,
	}
}

func (d *DposEnv) Root() (h common.Hash) {
	hw := sha3.New256()
	msgpack.NewEncoder(hw).Encode(d.epochTrie.Hash())
	msgpack.NewEncoder(hw).Encode(d.witnessTrie.Hash())
	msgpack.NewEncoder(hw).Encode(d.witnessCandidateTrie.Hash())
	msgpack.NewEncoder(hw).Encode(d.voteTrie.Hash())
	msgpack.NewEncoder(hw).Encode(d.mintCntTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DposEnv) Snapshot() *DposEnv {
	return d.Copy()
}

func (d *DposEnv) RevertToSnapShot(snapshot *DposEnv) {
	d.epochTrie = snapshot.epochTrie
	d.witnessTrie = snapshot.witnessTrie
	d.witnessCandidateTrie = snapshot.witnessCandidateTrie
	d.voteTrie = snapshot.voteTrie
	d.mintCntTrie = snapshot.mintCntTrie
}

func (d *DposEnv) FromProto(dcp *DposEnvProtocol) error {
	var err error

	d.epochTrie, err = NewEpochTrie(dcp.EpochHash, d.db)
	if err != nil {
		return err
	}
	d.witnessTrie, err = NewWitnessTrie(dcp.WitnessHash, d.db)
	if err != nil {
		return err
	}
	d.witnessCandidateTrie, err = NewWitnessCandidateTrie(dcp.WitnessCandidateHash, d.db)
	if err != nil {
		return err
	}
	d.voteTrie, err = NewVoteTrie(dcp.VoteHash, d.db)
	if err != nil {
		return err
	}
	d.mintCntTrie, err = NewMintCntTrie(dcp.MintCntHash, d.db)
	return err
}

type DposEnvProtocol struct {
	EpochHash            common.Hash `json:"epochRoot"        gencodec:"required"`
	WitnessHash          common.Hash `json:"delegateRoot"     gencodec:"required"`
	WitnessCandidateHash common.Hash `json:"candidateRoot"    gencodec:"required"`
	VoteHash             common.Hash `json:"voteRoot"         gencodec:"required"`
	MintCntHash          common.Hash `json:"mintCntRoot"      gencodec:"required"`
}

func (d *DposEnv) ToProto() *DposEnvProtocol {
	return &DposEnvProtocol{
		EpochHash:            d.epochTrie.Hash(),
		WitnessHash:          d.witnessTrie.Hash(),
		WitnessCandidateHash: d.witnessCandidateTrie.Hash(),
		VoteHash:             d.voteTrie.Hash(),
		MintCntHash:          d.mintCntTrie.Hash(),
	}
}

func (p *DposEnvProtocol) Root() (h common.Hash) {
	hw := sha3.New256()
	msgpack.NewEncoder(hw).Encode(p.EpochHash)
	msgpack.NewEncoder(hw).Encode(p.WitnessHash)
	msgpack.NewEncoder(hw).Encode(p.WitnessCandidateHash)
	msgpack.NewEncoder(hw).Encode(p.VoteHash)
	msgpack.NewEncoder(hw).Encode(p.MintCntHash)

	hw.Sum(h[:0])
	return h
}

func (d *DposEnv) RemoveWitnessCandidate(witnessCandidateAddr common.Address) error {
	witnessCandidate := witnessCandidateAddr.Bytes()
	err := d.witnessCandidateTrie.TryDelete(witnessCandidate)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	iter := trie.NewIterator(d.witnessTrie.PrefixIterator(witnessCandidate))
	for iter.Next() {
		witness := iter.Value
		key := append(witnessCandidate, witness...)
		err = d.witnessTrie.TryDelete(key)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		v, err := d.voteTrie.TryGet(witness)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		if err == nil && bytes.Equal(v, witnessCandidate) {
			err = d.voteTrie.TryDelete(witness)
			if err != nil {
				if _, ok := err.(*trie.MissingNodeError); !ok {
					return err
				}
			}
		}
	}
	return nil
}

func (d *DposEnv) BecomeWitnessCandidate(witnessCandidateAddr common.Address) error {

	witnessCandidate := witnessCandidateAddr.Bytes()
	return d.witnessCandidateTrie.TryUpdate(witnessCandidate, witnessCandidate)
}

func (d *DposEnv) Delegate(witnessAddr, witnessCandidateAddr common.Address) error {
	witness, witnessCandidate := witnessAddr.Bytes(), witnessCandidateAddr.Bytes()

	candidateInTrie, err := d.witnessCandidateTrie.TryGet(witnessCandidate)
	if err != nil {
		return err
	}
	if candidateInTrie == nil {
		return errors.New("invalid candidate to delegate")
	}

	oldCandidate, err := d.voteTrie.TryGet(witness)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	if oldCandidate != nil {
		d.witnessTrie.Delete(append(oldCandidate, witness...))
	}

	if err = d.witnessTrie.TryUpdate(append(witnessCandidate, witness...), witness); err != nil {
		return err
	}

	return d.voteTrie.TryUpdate(witness, witnessCandidate)
}

func (d *DposEnv) UnDelegate(witnessAddr, witnessCandidateAddr common.Address) error {

	witness, witnessCandidate := witnessAddr.Bytes(), witnessCandidateAddr.Bytes()

	candidateInTrie, err := d.witnessCandidateTrie.TryGet(witnessCandidate)
	if err != nil {
		return err
	}

	if candidateInTrie == nil {
		return errors.New("invalid candidate to undelegate")
	}

	oldCandidate, err := d.voteTrie.TryGet(witness)
	if err != nil {
		return err
	}

	if !bytes.Equal(witnessCandidate, oldCandidate) {
		return errors.New("mismatch candidate to undelegate")
	}

	if err = d.witnessTrie.TryDelete(append(witnessCandidate, witness...)); err != nil {
		return err
	}

	return d.voteTrie.TryDelete(witness)
}

func (d *DposEnv) Commit() (*DposEnvProtocol, error) {

	epochRoot, err := d.epochTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.epochTrie.TryUpdate(epochRoot[:], d.epochTrie.Get(epochRoot[:]))

	witnessRoot, err := d.witnessTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.witnessTrie.TryUpdate(witnessRoot[:], d.witnessTrie.Get(witnessRoot[:]))

	voteRoot, err := d.voteTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.voteTrie.TryUpdate(voteRoot[:], d.voteTrie.Get(voteRoot[:]))

	candidateRoot, err := d.witnessCandidateTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.witnessCandidateTrie.TryUpdate(candidateRoot[:], d.witnessCandidateTrie.Get(candidateRoot[:]))

	mintCntRoot, err := d.mintCntTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.mintCntTrie.TryUpdate(mintCntRoot[:], d.mintCntTrie.Get(mintCntRoot[:]))

	d.db.Commit(epochRoot, true)
	d.db.Commit(witnessRoot, true)
	d.db.Commit(candidateRoot, true)
	d.db.Commit(voteRoot, true)
	d.db.Commit(mintCntRoot, true)

	return &DposEnvProtocol{
		EpochHash:            epochRoot,
		WitnessHash:          witnessRoot,
		VoteHash:             voteRoot,
		WitnessCandidateHash: candidateRoot,
		MintCntHash:          mintCntRoot,
	}, nil
}

func (d *DposEnv) WitnessCandidateTrie() *trie.Trie   { return d.witnessCandidateTrie }
func (d *DposEnv) WitnessTrie() *trie.Trie            { return d.witnessTrie }
func (d *DposEnv) VoteTrie() *trie.Trie               { return d.voteTrie }
func (d *DposEnv) EpochTrie() *trie.Trie              { return d.epochTrie }
func (d *DposEnv) MintCntTrie() *trie.Trie            { return d.mintCntTrie }
func (d *DposEnv) DB() *trie.Database                 { return d.db }
func (dc *DposEnv) SetEpoch(epoch *trie.Trie)         { dc.epochTrie = epoch }
func (dc *DposEnv) SetWitness(witness *trie.Trie)     { dc.witnessTrie = witness }
func (dc *DposEnv) SetVote(vote *trie.Trie)           { dc.voteTrie = vote }
func (dc *DposEnv) SetCandidate(candidate *trie.Trie) { dc.witnessCandidateTrie = candidate }
func (dc *DposEnv) SetMintCnt(mintCnt *trie.Trie)     { dc.mintCntTrie = mintCnt }

func (dc *DposEnv) GetWitnesses() ([]Witness, error) {
	var witnesses []Witness
	key := []byte("witness")
	witnessessVal := dc.epochTrie.Get(key)

	if err := msgpack.Unmarshal(witnessessVal, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witness: %s", err)
	}
	return witnesses, nil
}

func (dc *DposEnv) SetWitnesses(witnesses []Witness) error {
	key := []byte("validator")
	witnessessVal, err := msgpack.Marshal(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode validators to rlp bytes: %s", err)
	}
	dc.epochTrie.Update(key, witnessessVal)
	return nil
}
