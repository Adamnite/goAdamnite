import (
	"fmt"
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
	"crypto/sha512"
    "encoding/hex"
	"github.com/vmihailenco/msgpack"


)





type TransactionWithSignature struct {
    Transaction *Transaction
    Signature []byte
}

func WrapTxWithSig(tx *Transaction, privateKey []byte) (*TransactionWithSignature, error) {
    txBytes, err := msgpack.Marshal(tx)
    if err != nil {
        return nil, err
    }

    sig, err := SignTx(txBytes, privateKey)
    if err != nil {
        return nil, err
    }

    return &TransactionWithSignature{
        Transaction: tx,
        Signature: sig,
    }, nil
}

func EncodeTx(tx *TransactionWithSignature) ([]byte, error) {
    var buf bytes.Buffer
    enc := msgpack.NewEncoder(&buf)

    err := enc.Encode(tx)
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}