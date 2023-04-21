package rpc

import (
	"log"
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

func (a *AdamniteClient) GetChainID() (*string, error) {
	log.Println("[Adamnite RPC client] Get chain ID")

	var reply []byte
	if err := a.client.Call(getChainIDEndpoint, nil, &reply); err != nil {
		log.Fatal("[Adamnite RPC Client] Call error: ", err)
		return nil, err
	}

	output := struct {
		ChainID string
	}{}

	if err := msgpack.Unmarshal(reply, &output); err != nil {
		log.Fatal("[Adamnite RPC Client] Decode error: ", err)
		return nil, err
	}

	return &output.ChainID, nil
}

func (a *AdamniteClient) GetBalance(address common.Address) (*string, error) {
	log.Println("[Adamnite RPC client] Get balance")

	input := struct {
		Address string
	}{Address: address.String()}

	data, err := msgpack.Marshal(input)
	if err != nil {
		log.Fatal("[Adamnite RPC Client] Encode error: ", err)
		return nil, err
	}

	var reply []byte
	if err := a.client.Call(getBalanceEndpoint, data, &reply); err != nil {
		log.Fatal("[Adamnite RPC Client] Call error: ", err)
		return nil, err
	}

	var balance string

	if err := msgpack.Unmarshal(reply, &balance); err != nil {
		log.Fatal("[Adamnite RPC Client] Decode error: ", err)
		return nil, err
	}

	return &balance, nil
}
func (a *AdamniteClient) GetAccounts() (*[]string, error) {
	log.Println("[Adamnite RPC client] Get block by hash")

	var reply []byte
	if err := a.client.Call(getAccountsEndpoint, nil, &reply); err != nil {
		log.Fatal("[Adamnite RPC Client] Call error: ", err)
		return nil, err
	}

	output := struct {
		Accounts []string
	}{}

	if err := msgpack.Unmarshal(reply, &output); err != nil {
		log.Fatal("[Adamnite RPC Client] Decode error: ", err)
		return nil, err
	}

	return &output.Accounts, nil
}

func NewAdamniteClient(endpoint string) (AdamniteClient, error) {
	client, err := rpc.Dial("tcp", endpoint)
	if err != nil {
		log.Fatal("[Adamnite RPC Client] Error while creating new client: ", err)
		return AdamniteClient{}, err
	}
	return AdamniteClient{endpoint, *client}, nil
}
