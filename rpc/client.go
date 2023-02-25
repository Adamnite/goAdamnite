package rpc

import (
	"log"
	"math/big"
	"net/rpc"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
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

func (a *AdamniteClient) GetBalance(address common.Address) (*big.Int, error) {
	bytes, err := msgpack.Marshal(&address)
	if err != nil {
		return nil, err
	}
	var reply Reply
	err = a.client.Call("AdamniteServer.GetBalance", bytes, &reply)

	if err != nil {
		return nil, err
	}
	var ansInt big.Int
	err = msgpack.Unmarshal(reply.Data, &ansInt)
	if err != nil {
		return nil, err
	}
	return &ansInt, nil
}

// create a new RPC client that will call to the following point.
func NewAdamniteClient(listenPoint string) *AdamniteClient {
	// fmt.Println(listenPoint)
	client, err := rpc.Dial("tcp", listenPoint)
	if err != nil {
		log.Fatal(err)
	}

	return &AdamniteClient{listenPoint, *client}
}
