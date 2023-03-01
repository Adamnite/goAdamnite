package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"net/rpc"
	"reflect"

	"github.com/adamnite/go-adamnite/common"
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

func (a *AdamniteClient) GetBalance(address common.Address) (*big.Int, error) {
	fmt.Println("starting GetBalanceClient side")

	var reply BigIntReply
	err := a.client.Call("AdamniteServer.GetBalance", address, &reply)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return reply.toBigInt(), nil
}

// create a new RPC client that will call to the following point.
func NewAdamniteClient(listenPoint string) *AdamniteClient {
	fmt.Println("new client generated")
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	conn, err := net.Dial("tcp", listenPoint)

	rpcCodec := codec.GoRpc.ClientCodec(conn, &mh)
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClientWithCodec(rpcCodec)
	return &AdamniteClient{listenPoint, *client}
}
