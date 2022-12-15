package adamclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/adamnite/go-adamnite/adm"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/hexutil"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/vmihailenco/msgpack/v5"
)

type Client struct {
	c *rpc.AdamniteClient
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.AdamniteClient) *Client {
	return &Client{c}
}

func (ac *Client) Close() {
	ac.c.Close()
}
func (ac *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := ac.c.CallContext(ctx, &result, "adm_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

func (ac *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return ac.getBlock(ctx, "adm_getBlockByHash", hash, true)
}

type rpcBlock struct {
	Hash         common.Hash      `json:"hash"`
	Transactions []rpcTransaction `json:"transactions"`
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (ac *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	var raw json.RawMessage
	err := ac.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, errors.New("not found")
	}
	// Decode header and transactions.
	var head *types.BlockHeader
	var body rpcBlock
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}

	if head.Hash() == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, fmt.Errorf("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.Hash() != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, fmt.Errorf("server returned empty transaction list but block header indicates transactions")
	}

	// Fill the sender cache of transactions in the block.
	txs := make([]*types.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		if tx.From != nil {
			setSenderFromServer(tx.tx, *tx.From, body.Hash)
		}
		txs[i] = tx.tx
	}
	return types.NewBlockWithHeader(head), nil
}

// HeaderByHash returns the block header with the given hash.
func (ec *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*types.BlockHeader, error) {
	var head *types.BlockHeader
	err := ec.c.CallContext(ctx, &head, "adm_getBlockByHash", hash, false)
	if err == nil && head == nil {
		err = errors.New("not found")
	}
	return head, err
}

func (ac *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	var json *rpcTransaction
	err = ac.c.CallContext(ctx, &json, "adm_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if json == nil {
		return nil, false, errors.New("not found")
	} else if _, r, _ := json.tx.RawSignature(); r == nil {
		return nil, false, fmt.Errorf("server returned transaction without signature")
	}
	if json.From != nil && json.BlockHash != nil {
		setSenderFromServer(json.tx, *json.From, *json.BlockHash)
	}
	return json.tx, json.BlockNumber == nil, nil
}

// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
func (ac *Client) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	// Try to load the address from the cache.
	sender, err := types.Sender(&senderFromServer{blockhash: block}, tx)
	if err == nil {
		return sender, nil
	}
	var meta struct {
		Hash common.Hash
		From common.Address
	}
	if err = ac.c.CallContext(ctx, &meta, "adm_getTransactionByBlockHashAndIndex", block, hexutil.Uint64(index)); err != nil {
		return common.Address{}, err
	}
	if meta.Hash == (common.Hash{}) || meta.Hash != tx.Hash() {
		return common.Address{}, errors.New("wrong inclusion block/index")
	}
	return meta.From, nil
}

// TransactionInBlock returns a single transaction at index in the given block.
func (ac *Client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	var json *rpcTransaction
	err := ac.c.CallContext(ctx, &json, "adm_getTransactionByBlockHashAndIndex", blockHash, hexutil.Uint64(index))
	if err != nil {
		return nil, err
	}
	if json == nil {
		return nil, errors.New("not found")
	} else if _, r, _ := json.tx.RawSignature(); r == nil {
		return nil, fmt.Errorf("server returned transaction without signature")
	}
	if json.From != nil && json.BlockHash != nil {
		setSenderFromServer(json.tx, *json.From, *json.BlockHash)
	}
	return json.tx, err
}
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ac *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	version := new(big.Int)
	var ver string
	if err := ac.c.CallContext(ctx, &ver, "net_version"); err != nil {
		return nil, err
	}
	if _, ok := version.SetString(ver, 10); !ok {
		return nil, fmt.Errorf("invalid net_version result %q", ver)
	}
	return version, nil
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ac *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result hexutil.Big
	err := ac.c.CallContext(ctx, &result, "adm_getBalance", account, toBlockNumArg(blockNumber))
	return (*big.Int)(&result), err
}

func (ac *Client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ac.c.CallContext(ctx, &result, "adm_getStorageAt", account, key, toBlockNumArg(blockNumber))
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ac *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ac.c.CallContext(ctx, &result, "adm_getCode", account, toBlockNumArg(blockNumber))
	return result, err
}

func (ac *Client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	var result hexutil.Big
	err := ac.c.CallContext(ctx, &result, "adm_getBalance", account, "pending")
	return (*big.Int)(&result), err
}

// PendingStorageAt returns the value of key in the contract storage of the given account in the pending state.
func (ac *Client) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	var result hexutil.Bytes
	err := ac.c.CallContext(ctx, &result, "adm_getStorageAt", account, key, "pending")
	return result, err
}

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ac *Client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	var result hexutil.Bytes
	err := ac.c.CallContext(ctx, &result, "adm_getCode", account, "pending")
	return result, err
}

// PendingTransactionCount returns the total number of transactions in the pending state.
func (ac *Client) PendingTransactionCount(ctx context.Context) (uint, error) {
	var num hexutil.Uint
	err := ac.c.CallContext(ctx, &num, "adm_getBlockTransactionCountByNumber", "pending")
	return uint(num), err
}

func (ac *Client) CallContract(ctx context.Context, msg adm.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := ac.c.CallContext(ctx, &hex, "adm_call", toCallArg(msg), toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (ac *Client) PendingCallContract(ctx context.Context, msg adm.CallMsg) ([]byte, error) {
	var hex hexutil.Bytes
	err := ac.c.CallContext(ctx, &hex, "adm_call", toCallArg(msg), "pending")
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (ac *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	data, err := msgpack.Marshal(tx)
	if err != nil {
		return err
	}
	return ac.c.CallContext(ctx, nil, "adm_sendRawTransaction", hexutil.Encode(data))
}

func toCallArg(msg adm.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
