package accounts

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
)

// Account represents an Adamnite account located at a specific location defined by the optional URL field
type Account struct {
	Address common.Address `json:"address"`
	URL     URL            `json:"url"`
}

// Wallet might contain one or more accounts
type Wallet interface {
	// URL indicates the path of the wallet.
	URL() URL

	// Open initializes access to a wallet instance.
	Open(passphrase string) error

	// Close releases any resources held by an open wallet instance.
	Close() error

	// Accounts retrieves the list of signing accounts the wallet is currently aware of.
	Accounts() []Account

	// Status returns a textual status to aid the user in the current state of the wallet.
	Status() (string, error)

	// SignData requests the wallet to sign the hash of the given data
	SignData(account Account, mimeType string, data []byte) ([]byte, error)

	// SignDataWithPassphrase is identical to SignData, but also takes a password
	SignDataWithPassphrase(account Account, passphrase, mimeType string, data []byte) ([]byte, error)

	// SignText requests the wallet to sign the hash of a given piece of data, prefixed
	// by the Adamnite prefix scheme
	SignText(account Account, text []byte) ([]byte, error)

	// SignTextWithPassphrase is identical to Signtext, but also takes a password
	SignTextWithPassphrase(account Account, passphrase string, hash []byte) ([]byte, error)

	// SignTx requests the wallet to sign the given transaction
	SignTx(account Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)

	// SignTxWithPassphrase is identical to SignTx, but also takes a password
	SignTxWithPassphrase(account Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)

	Verify(data []byte, sig []byte) (status bool, err error)
}
