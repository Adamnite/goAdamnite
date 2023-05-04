package adm

import (
	"fmt"
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
	"crypto/sha512"
    "encoding/hex"
)




type MessageCall struct {
    TransactionHash TxHash
    ContractAddress common.Address
    ContractMethod  string
    ContractParams  []interface{}
    Fee             uint64
}

func SignMessageCall(privateKey crypto.PrivateKey, msgCall *MessageCall) ([]byte, error) {
    msgCallBytes, err := msgpack.Marshal(msgCall)
    if err != nil {
        return nil, err
    }
    sig, err := crypto.Sign(msgCallBytes, privateKey)
    if err != nil {
        return nil, err
    }
    return append(sig, msgCallBytes...), nil
}

func EncodeMessageCall(msgCall *MessageCall) ([]byte, error) {
    return msgpack.Marshal(msgCall)
}

func DecodeMessageCall(data []byte) (*MessageCall, error) {
    var msgCall MessageCall
    err := msgpack.Unmarshal(data, &msgCall)
    if err != nil {
        return nil, err
    }
    return &msgCall, nil
}

func ExecuteMessageCall(vm *VM, offDB *OffChainDB, msgCall *MessageCall) ([]byte, error) {
    contract, err := offDB.FetchContract(msgCall.ContractAddress)
    if err != nil {
        return nil, err
    }

    // Execute contract method
    result, err := vm.ExecuteMethod(contract.Code, msgCall.ContractMethod, msgCall.ContractParams)
    if err != nil {
        return nil, err
    }

    // Update contract state in off-chain database
    err = offDB.UpdateContractState(msgCall.ContractAddress, contract.State)
    if err != nil {
        return nil, err
    }

    return result, nil
}
