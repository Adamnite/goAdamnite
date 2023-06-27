package consensus

//TODO: blocks are moved to utils (as they are needed in many places). Once Chain uses the correct block, this file should be removed
import (
	"time"

	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
)

// ConvertHeader converts from the old block header structure to the new one (temporary workaround)
func ConvertBlockHeader(header *types.BlockHeader) *utils.BlockHeader {
	return &utils.BlockHeader{
		Timestamp:             time.Unix(int64(header.Time), 0),
		TransactionType:       int8(networking.PrimaryTransactions), //TODO: make sure this is changed later
		ParentBlockID:         header.ParentHash,
		Witness:               crypto.PublicKey(header.Witness[:]), //TODO: this wont actually work, but enough to compile
		WitnessMerkleRoot:     header.WitnessRoot,
		TransactionMerkleRoot: header.TransactionRoot,
		StateMerkleRoot:       header.StateRoot,
		Number:                header.Number,
	}
}

// ConvertBlock converts from the old block structure to the new one (temporary workaround)
func ConvertBlock(block *types.Block) *utils.Block {
	return utils.NewBlock(
		block.Header().ParentHash,
		crypto.PublicKey(block.Header().Witness[:]), //TODO: this wont actually work, but enough to compile
		block.Header().WitnessRoot,
		block.Header().TransactionRoot,
		block.Header().StateRoot,
		block.Number(),
		ConvertTransactions(block.Body().Transactions),
	)
}
