package admpacket

import (
	"net"
	"crypto/ecdsa"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/vmihailenco/msgpack/v5"
)

type PeerEndpoint struct {
	IP  	net.IP
	UDP 	uint16
	TCP		uint16
}

type Ping struct {
	Version 	uint
	PubKey      *ecdsa.PublicKey
	From        PeerEndpoint
	To          PeerEndpoint
	ReqID 		[]byte
}

type Pong struct {
	Version     uint
	To          PeerEndpoint
	ReqID       []byte
}

func (p *Ping) Name() string {
	return "ADM-PING"
}

func (p *Ping) MessageType() byte {
	return PingMsg
}

func (p *Ping) RequestID() []byte {
	return p.ReqID
}

func (p *Ping) SetRequestID(id []byte) {
	p.ReqID = id
}

func (p *Pong) Name() string {
	return "ADM-PONG"
}

func (p *Pong) MessageType() byte {
	return PongMsg
}

func (p *Pong) RequestID() []byte {
	return p.ReqID
}

func (p *Pong) SetRequestID(id []byte) {
	p.ReqID = id
}

// ***********************************************************************************************************
// **********************************************  Serialization *********************************************
// ***********************************************************************************************************

var _ msgpack.CustomEncoder = (*Ping)(nil)
var _ msgpack.CustomDecoder = (*Ping)(nil)

func (p Ping) EncodeMsgpack(enc *msgpack.Encoder) error {
	pubkey := crypto.CompressPubkey(p.PubKey)

	return enc.EncodeMulti(p.Version, pubkey, p.From, p.To, p.ReqID)
}

func (p *Ping) DecodeMsgpack(dec *msgpack.Decoder) error {
	var pubkey []byte
	err := dec.DecodeMulti(&p.Version, &pubkey, &p.From, &p.To, &p.ReqID)
	if err != nil {
		return err
	}
	p.PubKey, err = crypto.DecompressPubkey(pubkey)
	if err != nil {
		return err
	}
	return nil
}