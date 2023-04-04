package wallet

import (
	"math/big"
	"github.com/adamnite/go-adamnite/main"
)

//Define a normal Adamnite Account

type Account struct {
	Address main.address
	Balance float64
}


//Define a Wallet, which handles the actual cryptographic signing of messages and other data
//Based heavily off of GoEthereum's current structure, may need to alter in future
type Wallet interface{
	Open() error
	Close() error
	SignData(data []byte) ([]byte, error)
	SignDataPassphrase(data []byte) ([]byte, error)
    SignText(text string) ([]byte, error)
	SignTextPassphrase(text string) ([]byte, error)
    SignTransaction(tx *types.Transaction) ([]byte, error)
	SignTransactionPassphrase(tx *types.Transaction) ([]byte, error)
	Verify(data []byte, sig []byte) (status bool, err error)

} 