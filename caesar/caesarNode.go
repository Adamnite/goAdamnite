package caesar

import (
	"sort"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type CaesarNode struct {
	netHandler        *networking.NetNode
	signerSet         *accounts.Account
	msgByHash         map[string]*utils.CaesarMessage
	msgByRecipient    map[common.Address][]*utils.CaesarMessage
	msgBySender       map[common.Address][]*utils.CaesarMessage
	NewMessageUpdater func(*utils.CaesarMessage)
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

// adds a bouncer, so web based users can access messages through this
func (cn *CaesarNode) StartBouncer() {
	cn.netHandler.AddBouncerServer(nil, nil, 0)
	cn.netHandler.SetBounceServerMessaging(cn.GetMessagesBetween)
}
func (cn CaesarNode) GetConnectionPoint() string {
	return cn.netHandler.GetConnectionString()
}
func (cn CaesarNode) GetMessagesBetween(a, b common.Address) []*utils.CaesarMessage {
	ansMessages := []*utils.CaesarMessage{}
	for _, msg := range cn.msgBySender[a] {
		if msg.To.Address == b {
			ansMessages = append(ansMessages, msg)
		}
	}
	for _, msg := range cn.msgBySender[b] {
		if msg.To.Address == a {
			ansMessages = append(ansMessages, msg)
		}
	}
	//sort them to be in order, cause why not
	sort.Slice(ansMessages, func(i, j int) bool {
		return ansMessages[i].InitialTime < ansMessages[j].InitialTime
	}) //TODO: why not might be for important performance reasons...

	return ansMessages
}

// connect this to a seed node that it can propagate from
func (cn *CaesarNode) ConnectToNetworkFrom(seedConnectionPoint string) error {
	//TODO: once we have a seed node running, we should add a " seedConnectionPoint == '', use our known default"
	if err := cn.netHandler.ConnectToSeed(seedConnectionPoint); err != nil {
		return err
	}
	return cn.FillNetworking(false)
}

// handle all the network stuff, and fill this nodes connections (pretty wide considering how little processor power this server type should take)
func (cn *CaesarNode) FillNetworking(overrideSetLimits bool) error {
	if overrideSetLimits {
		cn.netHandler.SetMaxGreyList(0)
		cn.netHandler.SetMaxConnections(64)
	}
	err := cn.netHandler.SprawlConnections(2, 0)
	if err != nil && err != networking.ErrNoNewConnectionsMade {
		//reaching every node on the network isn't the worst, and no further problems would occur filling the connections then
		return err
	}

	return cn.netHandler.FillOpenConnections()
}

func (cn *CaesarNode) AddMessage(msg *utils.CaesarMessage) {
	if _, exists := cn.msgByHash[string(msg.Hash())]; exists {
		return
	}
	if cn.NewMessageUpdater != nil {
		cn.NewMessageUpdater(msg)
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
