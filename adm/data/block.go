package adm



//Fix for proper imports within ADM
import (
	"fmt"
	"time"
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
	"crypto/sha512"
    "encoding/hex"

)



//Replace with the normal hash from crypto
type BlockHash [32]byte



type BlockHeader struct {
    Round         uint64
    PreviousBlock BlockHash
    ConsensusSeed []byte
    ChainID       uint32
    Timestamp     int64
    WitnessState  []byte
    GenesisHash   BlockHash
    TxRoot        []byte
    BatchUpdate   bool
    DBWitnessRoot []byte // present only if BatchUpdate is true
    ContractState BlockHash
	ProtocolVote 
    UpgradeState  uint32
}




//Add in ate/gas/fee structure, And rewards. More documentation to come later.

func (bh *BlockHeader) Hash() BlockHash {
    h := sha512.New512_256()
    h.Write([]byte{byte(bh.Round >> 56), byte(bh.Round >> 48), byte(bh.Round >> 40), byte(bh.Round >> 32),
        byte(bh.Round >> 24), byte(bh.Round >> 16), byte(bh.Round >> 8), byte(bh.Round)})
    h.Write(bh.PreviousBlock[:])
    h.Write(bh.ConsensusSeed)
    h.Write([]byte{byte(bh.ChainID >> 24), byte(bh.ChainID >> 16), byte(bh.ChainID >> 8), byte(bh.ChainID)})
    h.Write([]byte{byte(bh.Timestamp >> 56), byte(bh.Timestamp >> 48), byte(bh.Timestamp >> 40), byte(bh.Timestamp >> 32),
        byte(bh.Timestamp >> 24), byte(bh.Timestamp >> 16), byte(bh.Timestamp >> 8), byte(bh.Timestamp)})
    h.Write(bh.WitnessState)
    h.Write(bh.GenesisHash[:])
    h.Write(bh.TxRoot)
    h.Write([]byte{0})
    if bh.BatchUpdate {
        h.Write(bh.DBWitnessRoot)
    }
    h.Write(bh.ContractState[:])
    h.Write([]byte{byte(bh.UpgradeState >> 24), byte(bh.UpgradeState >> 16), byte(bh.UpgradeState >> 8), byte(bh.UpgradeState)})
    var hash BlockHash
    copy(hash[:], h.Sum(nil))
    return hash
}

func (bh *BlockHeader) HashString() string {
    return hex.EncodeToString(bh.Hash()[:])
}



type UpgradeVote struct {
    CurrentVersion   string
    ProposedVersion  string
    IsApproved       bool
}



type UpgradeState struct {
    currentVersion string
    pastUpgrades map[string]bool
    newUpgrade string
    newUpgradeApproval bool
    validFromConsensusRound uint64
}



func GetRound(blockHeader BlockHeader) uint64 {
    return blockHeader.Round
}

func GetPreviousBlockHash(blockHeader BlockHeader) Hash {
    return blockHeader.PreviousBlockHash
}

func GetSeed(blockHeader BlockHeader) Hash {
    return blockHeader.Seed
}

func GetChainID(blockHeader BlockHeader) string {
    return blockHeader.ChainID
}

func GetTimestamp(blockHeader BlockHeader) time.Time {
    return blockHeader.Timestamp
}

func GetWitnessStateRoot(blockHeader BlockHeader) Hash {
    return blockHeader.WitnessStateRoot
}

func GetGenesisHash(blockHeader BlockHeader) Hash {
    return blockHeader.GenesisHash
}

func GetTransactionRoot(blockHeader BlockHeader) Hash {
    return blockHeader.TransactionRoot
}

func GetBatchUpdate(blockHeader BlockHeader) bool {
    return blockHeader.BatchUpdate
}

func GetDBWitnessRoot(blockHeader BlockHeader) Hash {
    return blockHeader.DBWitnessRoot
}

func GetContractStateHash(blockHeader BlockHeader) Hash {
    return blockHeader.ContractStateHash
}

func GetUpgradeState(blockHeader BlockHeader) uint64 {
    return blockHeader.UpgradeState
}


type Block struct {
    Hash   BlockHash
    Header BlockHeader
	Transactions []transaction //placeholder
}



type BlockHeader struct {
    Round         uint64
    PrevBlockHash BlockHash
    Seed          []byte
    ChainID       uint64
    Timestamp     int64
    StateRoot     []byte
    GenesisHash   []byte
    TxRoot        []byte
    BatchUpdate   bool
    DBWitnessRoot []byte
    ContractState []byte
    UpgradeState  uint64
}

type BlockHash [sha512.Size256]byte

type Block struct {
    Hash   BlockHash
    Header BlockHeader
}

func createBlock(round uint64, prevBlockHash BlockHash, seed []byte, chainID uint64, timestamp int64, stateRoot []byte, genesisHash []byte, txRoot []byte, batchUpdate bool, dbWitnessRoot []byte, contractState []byte, upgradeState uint64) Block {
    header := BlockHeader{
        Round:         round,
        PrevBlockHash: prevBlockHash,
        Seed:          seed,
        ChainID:       chainID,
        Timestamp:     timestamp,
        StateRoot:     stateRoot,
        GenesisHash:   genesisHash,
        TxRoot:        txRoot,
        BatchUpdate:   batchUpdate,
        DBWitnessRoot: dbWitnessRoot,
        ContractState: contractState,
        UpgradeState:  upgradeState,
    }
    hash := calculateBlockHash(header)
    return Block{hash, header}
}

func checkBlockHeader(header BlockHeader, prevBlockHash BlockHash, genesisID uint64, genesisHash []byte, prevTimestamp int64) error {
    if header.Timestamp <= 0 {
        return errors.New("invalid timestamp")
    }
    if header.Timestamp <= prevTimestamp {
        return errors.New("timestamp must be greater than previous block")
    }
    if header.GenesisHash != genesisHash {
        return errors.New("invalid genesis hash")
    }
    if header.ChainID != genesisID {
        return errors.New("invalid chain id")
    }
    if header.PrevBlockHash != prevBlockHash {
        return errors.New("invalid previous block hash")
    }
    // Check other parameters as needed
    return nil
}

func calculateBlockHash(header BlockHeader) BlockHash {
    h := sha512.New512_256()
    // Add header fields to hash
    // Use placeholders for items not yet implemented
    h.Write([]byte("Round:"))
    h.Write([]byte(strconv.FormatUint(header.Round, 10)))
    h.Write([]byte("|PrevBlockHash:"))
    h.Write(header.PrevBlockHash[:])
    h.Write([]byte("|Seed:"))
    h.Write(header.Seed)
    h.Write([]byte("|ChainID:"))
    h.Write([]byte(strconv.FormatUint(header.ChainID, 10)))
    h.Write([]byte("|Timestamp:"))
    h.Write([]byte(strconv.FormatInt(header.Timestamp, 10)))
    h.Write([]byte("|StateRoot:"))
    h.Write(header.StateRoot)
    h.Write([]byte("|GenesisHash:"))
    h.Write(header.GenesisHash)
    h.Write([]byte("|TxRoot:"))
    h.Write(header.TxRoot)
    h.Write([]byte("|BatchUpdate:"))
    h.Write([]byte(strconv.FormatBool(header.BatchUpdate)))
    h.Write([]byte("|DBWitnessRoot:"))
    h.Write(header.DBWitnessRoot)
    h.Write([]byte("|ContractState:"))
    h.Write(header.ContractState)
    h.Write([]byte("|UpgradeState:"))
    h.Write([]byte(strconv.FormatUint(header.UpgradeState, 10)))
    var hash BlockHash
    copy(hash[:], h.Sum(nil))
    return hash
}



type SignedTransaction struct {
    // transaction fields here
}

// Split transactions into groups
func splitTransactions(transactions []SignedTransaction, groupSize int) [][]SignedTransaction {
    numGroups := (len(transactions) + groupSize - 1) / groupSize
    groups := make([][]SignedTransaction, numGroups)
    for i := range groups {
        start := i * groupSize
        end := start + groupSize
        if end > len(transactions) {
            end = len(transactions)
        }
        groups[i] = transactions[start:end]
    }
    return groups
}

// Encode a group of transactions for inclusion into a block
func encodeTransactions(transactions []SignedTransaction) ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(transactions)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
