package rpc

import (
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

//the server items for sending and receiving a message

func (a *AdamniteServer) SetCaesarMessagingHandlers(msgH func(*utils.CaesarMessage)) {
	a.newMessageHandler = msgH
}

const newMessageEndpoint = "AdamniteServer.NewCaesarMessage"

func (a *AdamniteServer) NewCaesarMessage(params *[]byte, reply *[]byte) error {
	a.print("New Caesar Message")
	if a.newMessageHandler == nil {
		a.print("not setup to handle/log messages")
		return nil //we aren't setup to handle it, just forward it
	}
	var msg utils.CaesarMessage
	if err := encoding.Unmarshal(*params, &msg); err != nil {
		a.printError("New Candidate", err)
		return err
	}
	go a.newMessageHandler(&msg)
	return nil
}
