package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"reflect"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/ugorji/go/codec"
)

type AdamniteClient struct {
	callAddress string
	client      rpc.Client
}

func (a *AdamniteClient) CallAsync() {

}
func (a *AdamniteClient) Close() {
	a.client.Close()
}

type SendingAddress struct {
	Value common.Address
}

func (a *AdamniteClient) GetChainID() (*big.Int, error) {
	fmt.Println("starting GetChainID client side")
	var reply BigIntRPC
	err := a.client.Call(getChainIDEndpoint, nil, &reply)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return reply.toBigInt(), nil
}

func (a *AdamniteClient) GetBalance(address common.Address) (*big.Int, error) {
	fmt.Println("starting GetBalanceClient side")

	var reply BigIntRPC
	err := a.client.Call(getBalanceEndpoint, address, &reply)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return reply.toBigInt(), nil
}
func (a *AdamniteClient) GetBlockByHash(hash common.Hash) (*types.Block, error) {
	fmt.Println("starting GetBlockByHash Client side")
	var reply types.Block
	err := a.client.Call(getBlockByHashEndpoint, hash, &reply)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &reply, nil
}

// create a new RPC client that will call to the following point.
func NewAdamniteClient(listenPoint string) *AdamniteClient {
	fmt.Println("new client generated")
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	conn, err := net.Dial("tcp", listenPoint)

	rpcCodec := codec.GoRpc.ClientCodec(conn, &mh)
	if err != nil {
		log.Println(err)
	}
	client := rpc.NewClientWithCodec(rpcCodec)
	return &AdamniteClient{listenPoint, *client}
}
