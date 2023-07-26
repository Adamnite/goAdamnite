package consensus

//TODO: once chain is converted to using new transactions, all of this can be deleted
import (
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/utils"
)

// ConvertTransactions converts between old transaction structure and new one (temporary workaround)
func ConvertTransactions(transactions []*types.Transaction) []*utils.BaseTransaction {
	var ts []*utils.BaseTransaction
	for _, t := range transactions {
		ts = append(ts, convertTransaction(t))
	}
	return ts
}

func convertTransaction(t *types.Transaction) *utils.BaseTransaction {
	ans := utils.BaseTransaction{
		Version:  utils.TRANSACTION_V0,
		Amount:   t.Amount(),
		GasLimit: t.ATEPrice(),
		Time:     t.Timestamp,
	}
	// TODO: convert following members from old transaction type
	return &ans
}
