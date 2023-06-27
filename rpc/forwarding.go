package rpc

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
	encoding "github.com/vmihailenco/msgpack/v5"
)

type ForwardingContent struct {
	InitialTime     int64           //when this was created recorded in unix milliseconds
	FinalEndpoint   string          //the final endpoint to call
	DestinationNode *common.Address //null if its for everyone
	FinalParams     []byte          //the params to be passed at the end
	FinalReply      []byte          //ignored if DestinationNode is nill, otherwise will attempt to link back
	InitialSender   common.Address  //who started this
}

func (fc ForwardingContent) Hash() common.Hash {
	byteForm := []byte(fc.FinalEndpoint)
	if fc.DestinationNode != nil {
		byteForm = append(byteForm, fc.DestinationNode.Bytes()...)
	}
	byteForm = append(byteForm, fc.FinalParams...)
	byteForm = append(byteForm, fc.FinalReply...)
	byteForm = append(byteForm, binary.LittleEndian.AppendUint64([]byte{}, uint64(fc.InitialTime))...)

	return common.BytesToHash(crypto.Sha512(byteForm))
}

// create a forwarding message to be sent to everyone
func CreateForwardToAll(finalMessage interface{}) (ForwardingContent, error) {
	messageBytes, err := encoding.Marshal(finalMessage)
	if err != nil {
		return ForwardingContent{}, err
	}
	forwardAns := ForwardingContent{
		InitialTime: time.Now().UnixMilli(),
		FinalParams: messageBytes,
	}
	switch finalMessage.(type) {
	case utils.CaesarMessage, *utils.CaesarMessage:
		forwardAns.FinalEndpoint = newMessageEndpoint
	case utils.Candidate, *utils.Candidate:
		forwardAns.FinalEndpoint = NewCandidateEndpoint
	case utils.Voter, *utils.Voter:
		forwardAns.FinalEndpoint = NewVoteEndpoint
	case utils.TransactionType, *utils.TransactionType:
		forwardAns.FinalEndpoint = NewTransactionEndpoint
	case utils.BlockType, *utils.Block, *utils.VMBlock:
		forwardAns.FinalEndpoint = NewBlockEndpoint
	default:
		return ForwardingContent{}, fmt.Errorf("endpoint for forwarding message could not be determined")
	}
	return forwardAns, nil
}
