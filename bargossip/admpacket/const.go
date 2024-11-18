package admpacket

import (
	"encoding/binary"
	"time"
)

var admPacketProtocolID = [10]byte{'B', 'A', 'R', '-', 'G', 'O', 'S', 'S', 'I', 'P'} // BAR-GOSSIP

const (
	handshakeTimeout = time.Millisecond * 1000

	admPacketVersionV1 = 1

	sizeOfIV            = 16
	randomPacketMsgSize = 20
)

const (
	PingMsg         = byte(0)
	PongMsg         = byte(1)
	FindnodeMsg     = byte(2)
	RspFindnodeMsg  = byte(3)
	AskHandshakeMsg = byte(4)
	SYNMsg          = byte(5)
)

const (
	messagePHT       = byte(0)
	askHandshakePHT  = byte(1)
	handshakeBodyPHT = byte(2)
)

var (
	sizeofMessageAuthData      = binary.Size(messageAuthData{})
	sizeofStaticHeader         = binary.Size(StaticHeader{})
	sizeofAskHandshakeAuthData = binary.Size(askHandshakeAuthData{})
	sizeofHandshakeAuthData    = binary.Size(handshakeAuthData{}.SrcID) + binary.Size(handshakeAuthData{}.PubkeySize) + binary.Size(handshakeAuthData{}.SignatureSize)
	sizeofStaticPacketHeader   = sizeOfIV + sizeofStaticHeader
)
