package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net/rpc"

	"github.com/adamnite/go-adamnite/common"
	"github.com/vmihailenco/msgpack/v5"
)

type AdamniteClient struct {
	endpoint string
	client   rpc.Client
}

func (a *AdamniteClient) Close() {
	a.client.Close()
}

const clientPreface = "[Adamnite RPC server] %v \n"

func (a *AdamniteClient) GetContactList() *PassedContacts {
	log.Printf(clientPreface, "Get Contact List")
	var passed *PassedContacts
	var reply []byte
	if err := a.client.Call(getContactsListEndpoint, []byte{}, &reply); err != nil { //you can, in fact, not just pass nil
		log.Println(err)
		return nil
	}
	if err := msgpack.Unmarshal(reply, &passed); err != nil {
		log.Println(err)
		return nil
	}
	return passed
}

func (a *AdamniteClient) GetChainID() (*big.Int, error) {
	log.Printf(clientPreface, "Get chain id")
	reply := []byte{}
	err := a.client.Call(getChainIDEndpoint, []byte{}, &reply)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var output big.Int

	if err := msgpack.Unmarshal(reply, &output); err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("error: %v", err))
		return nil, err
	}

	return &output, nil
}

func (a *AdamniteClient) GetBalance(address common.Address) (*string, error) {
	log.Printf(clientPreface, "Get balance")

	input := struct {
		Address string
	}{Address: address.String()}

	data, err := msgpack.Marshal(input)
	if err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("Encode error: %v", err))
		return nil, err
	}

	var reply []byte
	if err := a.client.Call(getBalanceEndpoint, data, &reply); err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("Call error: %v", err))
		return nil, err
	}

	var balance string

	if err := msgpack.Unmarshal(reply, &balance); err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("error: %v", err))
		return nil, err
	}

	return &balance, nil
}
func (a *AdamniteClient) GetAccounts() (*[]string, error) {
	log.Printf(clientPreface, "Get block by hash")

	var reply []byte
	if err := a.client.Call(getAccountsEndpoint, nil, &reply); err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("Call error: %v", err))
		return nil, err
	}

	output := struct {
		Accounts []string
	}{}

	if err := msgpack.Unmarshal(reply, &output); err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("error: %v", err))
		return nil, err
	}

	return &output.Accounts, nil
}

func NewAdamniteClient(endpoint string) (AdamniteClient, error) {
	client, err := rpc.Dial("tcp", endpoint)
	if err != nil {
		log.Fatalf(clientPreface, fmt.Sprintf("Error while creating new client: %v", err))
		return AdamniteClient{}, err
	}
	return AdamniteClient{endpoint, *client}, nil
}
