package caesar

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)


type CaesarNode struct {
	netHandler     *networking.NetNode
	signerSet      *accounts.Account
	msgByHash      map[string]*utils.CaesarMessage
	msgByRecipient map[common.Address][]*utils.CaesarMessage
	msgBySender    map[common.Address][]*utils.CaesarMessage
}

func NewCaesarNode(sendingKey *accounts.Account) *CaesarNode {
	cn := CaesarNode{
		msgByHash:      make(map[string]*utils.CaesarMessage),
		msgByRecipient: make(map[common.Address][]*utils.CaesarMessage),
		msgBySender:    make(map[common.Address][]*utils.CaesarMessage),
	}
	if sendingKey == nil {
		cn.signerSet, _ = accounts.GenerateAccount()
	} else {
		cn.signerSet = sendingKey
	}

	cn.netHandler = networking.NewNetNode(cn.signerSet.Address)

	return &cn
}

func (cn *CaesarNode) Startup() error {
	cn.netHandler.AddMessagingCapabilities(
		cn.AddMessage,
	)

	return nil
}

func (cn *CaesarNode) AddMessage(msg *utils.CaesarMessage) {
	if _, exists := cn.msgByHash[string(msg.Hash())]; exists {
		return
	}

	cn.msgByHash[string(msg.Hash())] = msg
	if msgArray, exists := cn.msgByRecipient[msg.To.Address]; exists {
		cn.msgByRecipient[msg.To.Address] = []*utils.CaesarMessage{msg}
	} else {
		cn.msgByRecipient[msg.To.Address] = append(msgArray, msg)
	}
	if msgArray, exists := cn.msgBySender[msg.From.Address]; exists {
		cn.msgBySender[msg.From.Address] = []*utils.CaesarMessage{msg}
	} else {
		cn.msgBySender[msg.From.Address] = append(msgArray, msg)
	}
}
func (cn *CaesarNode) SendMessage(msg *utils.CaesarMessage) error {
	cn.AddMessage(msg)
	return cn.netHandler.Propagate(msg)
}
func (cn *CaesarNode) Send(to accounts.Account, message string) error {
	msg, err := utils.NewCaesarMessage(to, *cn.signerSet, message)
	if err != nil {
		return err
	}
	if err := msg.Sign(); err != nil {
		return err
	}
	return cn.SendMessage(msg)
}
