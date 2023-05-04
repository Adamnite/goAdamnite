package adm

//Fix for proper imports within ADM
import (
	"fmt"
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
	"crypto/sha512"
    "encoding/hex"
)




type transactionHash [32]byte

type Transaction struct {
    Type            string
    Amount          uint64
    SenderAddress   common.Address
    RecipientAddress common.Address //Reciepient is not present in a message call.
    MaxTransactionFee uint64
    Message         string
    Hash            transactionHash
	Nonce 			common.number
	Round			common.Round
	MaxRound		common.Round
	//Following parameters are only present if the type is a message call
	ContractAddress common.Address
    ContractMethod  string
    ContractParams  []interface{}

}

type fetchBalances interface {
    UpdateBalance(address common.Address, balance uint64) error
    GetBalance(address common.Address) (uint64, error)
}




func (t *Transaction) GetTransactionHash() string {
    // Calculate transaction hash using SHA-256 algorithm
    // ...
    return "<transaction hash>"
}

func (t *Transaction) GetType() string {
    return t.Type
}

func (t *Transaction) GetAmount() int {
    return t.Amount
}

func (t *Transaction) GetSender() common.Address {
    return t.Sender
}

func (t *Transaction) GetReceiver() common.Address {
    return t.Receiver
}

func (t *Transaction) GetMaxFee() int {
    return t.MaxFee
}

func (t *Transaction) GetMessage() string {
    return t.Message
}

func (t *Transaction) GetRound() int{
	return t.Round
}

func (t *Transaction) GetMaxRound() int{
	return t.MaxRound
}

func 

func SignTx(tx *Transaction, privateKey *rsa.PrivateKey) (string, error) {
    // Serialize the transaction
    txBytes, err := json.Marshal(tx)
    if err != nil {
        return "", err
    }

    // Hash the serialized transaction
    hash := sha512.Sum512_256(txBytes)

    // Sign the hash using the private key
	//Use actual crypto library instead of this 
    signature, err := crypto.sign(rand.Reader, privateKey, crypto.SHA512_256, hash[:])
    if err != nil {
        return "", err
    }

    // Convert the signature to a hex-encoded string
    signatureStr := hex.EncodeToString(signature)

    return signatureStr, nil
}





func apply_transaction(tx *Transaction) error {
    // Get sender and recipient account balances
    sender_balance := get_account_balance(tx.sender)
    recipient_balance := get_account_balance(tx.recipient)

    // Check if sender has enough balance to make the transaction
    if sender_balance < tx.amount+tx.max_fee {
        return fmt.Errorf("sender does not have enough balance to make the transaction")
    }

    // Update sender and recipient balances
    update_account_balance(tx.sender, sender_balance-tx.amount-tx.max_fee)
    update_account_balance(tx.recipient, recipient_balance+tx.amount)

    // Update balance table with transaction fee
    update_account_balance(common.fee_collector_address, get_account_balance(common.fee_collector_address)+tx.max_fee)

    return nil
}
