package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net/rpc"

	"github.com/adamnite/go-adamnite/common"
	encoding "github.com/vmihailenco/msgpack/v5"
)

const clientPreface = "[Adamnite RPC client] %v \n"

type AdamniteClient struct {
	endpoint          string
	client            rpc.Client
	callerAddress     *common.Address
	hostingServerPort string //the string version of the port that our Server is running on.
}

func (a *AdamniteClient) SetAddressAndHostingPort(add *common.Address, hostingPort string) {
	a.callerAddress = add
	a.hostingServerPort = hostingPort
}

func (a *AdamniteClient) Close() {
	a.client.Close()
}

func (a *AdamniteClient) ForwardMessage(content ForwardingContent, reply *[]byte) error {
	log.Printf(clientPreface, "Forward Message")
	// just pass this message along until someone who needs it finds it
	// var forwardingBytes []byte
	forwardingBytes, err := encoding.Marshal(content)
	if err != nil {
		return err
	}
	return a.client.Call(forwardMessageEndpoint, forwardingBytes, &reply)
}

func (a *AdamniteClient) GetVersion() (*AdmVersionReply, error) {
	log.Printf(clientPreface, "Get Version")
	var reply []byte = []byte{}
	var versionReceived AdmVersionReply

	if a.callerAddress == nil {
		return nil, ErrNoAccountSet
	}
	sendingData := struct {
		Address           common.Address
		HostingServerPort string
	}{Address: *a.callerAddress, HostingServerPort: a.hostingServerPort}

	addressBytes, err := encoding.Marshal(sendingData)
	if err != nil {
		return nil, err
	}
	if err := a.client.Call(getVersionEndpoint, addressBytes, &reply); err != nil {
		log.Panicln(err)
		return nil, err
	}
	if err := encoding.Unmarshal(reply, &versionReceived); err != nil {
		log.Println(err)
		return nil, err
	}
	versionReceived.Timestamp = versionReceived.Timestamp.UTC()
	return &versionReceived, nil
}

func (a *AdamniteClient) GetContactList() *PassedContacts {
	log.Printf(clientPreface, "Get Contact List")
	var passed *PassedContacts
	var reply []byte
	if err := a.client.Call(getContactsListEndpoint, []byte{}, &reply); err != nil {
		log.Println(err)
		return nil
	}
	if err := encoding.Unmarshal(reply, &passed); err != nil {
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

	if err := encoding.Unmarshal(reply, &output); err != nil {
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

	data, err := encoding.Marshal(input)
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

	if err := encoding.Unmarshal(reply, &balance); err != nil {
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

	if err := encoding.Unmarshal(reply, &output); err != nil {
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
	return AdamniteClient{
		endpoint: endpoint,
		client:   *client,
	}, nil
}
