package rpc

import (
	"fmt"
	"log"
	"math/big"
	"net/rpc"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

const clientPreface = "[Adamnite RPC client] %v \n"

func (a *AdamniteClient) print(methodName string) {
	if a.DebugOutput {
		log.Printf(clientPreface, methodName)
	}
}
func (a *AdamniteClient) printError(methodName string, err error) {
	log.Printf(clientPreface, fmt.Sprintf("%v\tError: %s", methodName, err))
}

type AdamniteClient struct {
<<<<<<< Updated upstream
	DebugOutput       bool
	endpoint          string
	client            rpc.Client
	callerAddress     *common.Address
	hostingServerPort string //the string version of the port that our Server is running on.
=======
	DebugOutput            bool
	endpoint               string
	client                 *rpc.Client
	callerAddress          *bytes.Address
	hostingServerConnector string //the string version of the port that our Server is running on.
>>>>>>> Stashed changes
}

func (a *AdamniteClient) SetAddressAndHostingPort(add *bytes.Address, hostingPort string) {
	a.callerAddress = add
	a.hostingServerPort = hostingPort
}

func (a *AdamniteClient) Close() {
	a.client.Close()
}
func (a *AdamniteClient) SendTransaction(transaction *utils.Transaction) error {
	log.Printf(clientPreface, "Send Transaction")
	data, err := encoding.Marshal(&transaction)
	if err != nil {
		return err
	}

	return a.client.Call(SendTransactionEndpoint, &data, &[]byte{})
}
func (a *AdamniteClient) ForwardMessage(content ForwardingContent, reply *[]byte) error {
	a.print("Forward Message")
	// just pass this message along until someone who needs it finds it
	// var forwardingBytes []byte
	forwardingBytes, err := encoding.Marshal(&content)
	if err != nil {
		return err
	}
	return a.client.Call(forwardMessageEndpoint, forwardingBytes, &reply)
}

func (a *AdamniteClient) GetVersion() (*AdmVersionReply, error) {
	a.print("Get Version")
	var reply []byte = []byte{}
	var versionReceived AdmVersionReply

	if a.callerAddress == nil {
		return nil, ErrNoAccountSet
	}
	sendingData := struct {
<<<<<<< Updated upstream
		Address           common.Address
		HostingServerPort string
	}{Address: *a.callerAddress, HostingServerPort: a.hostingServerPort}
=======
		Address                 bytes.Address
		HostingServerConnection string
	}{Address: *a.callerAddress, HostingServerConnection: a.hostingServerConnector}
>>>>>>> Stashed changes

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
	a.print("Get Contact List")
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

<<<<<<< Updated upstream
func (a *AdamniteClient) GetChainID() (*big.Int, error) {
	a.print("Get chain id")
	reply := []byte{}
	err := a.client.Call(getChainIDEndpoint, []byte{}, &reply)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var output big.Int

	if err := encoding.Unmarshal(reply, &output); err != nil {
		a.printError("Get chain id", err)
		return nil, err
	}

	return &output, nil
}

func (a *AdamniteClient) GetBalance(address common.Address) (*string, error) {
=======
func (a *AdamniteClient) GetBalance(address bytes.Address) (*string, error) {
>>>>>>> Stashed changes
	a.print("Get balance")

	input := struct {
		Address string
	}{Address: address.String()}

	data, err := encoding.Marshal(input)
	if err != nil {
		a.printError("Get balance", err)
		return nil, err
	}

	var reply []byte
	if err := a.client.Call(getBalanceEndpoint, data, &reply); err != nil {
		a.printError("Get balance", err)
		return nil, err
	}

	var balance string

	if err := encoding.Unmarshal(reply, &balance); err != nil {
		a.printError("Get balance", err)
		return nil, err
	}

	return &balance, nil
}
func (a *AdamniteClient) GetAccounts() (*[]string, error) {
	a.print("Get Accounts")

	var reply []byte
	if err := a.client.Call(getAccountsEndpoint, nil, &reply); err != nil {
		a.printError("Get Accounts", err)
		return nil, err
	}

	output := struct {
		Accounts []string
	}{}

	if err := encoding.Unmarshal(reply, &output); err != nil {
		a.printError("Get Accounts", err)
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
		DebugOutput: false,
		endpoint:    endpoint,
		client:      *client,
	}, nil
}
