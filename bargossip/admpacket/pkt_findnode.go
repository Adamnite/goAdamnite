package admpacket

import "github.com/adamnite/go-adamnite/bargossip/admnode"

type Findnode struct {
	ReqID     []byte
	Distances []uint
}

type RspNodes struct {
	ReqID []byte
	Total uint8
	Nodes []*admnode.NodeInfo
}

func (p *Findnode) Name() string {
	return "ADM-Findnode-req"
}

func (p *Findnode) MessageType() byte {
	return FindnodeMsg
}

func (p *Findnode) RequestID() []byte {
	return p.ReqID
}

func (p *Findnode) SetRequestID(id []byte) {
	p.ReqID = id
}

func (p *RspNodes) Name() string {
	return "ADM-Findnode-rsp"
}

func (p *RspNodes) MessageType() byte {
	return RspFindnodeMsg
}

func (p *RspNodes) RequestID() []byte {
	return p.ReqID
}

func (p *RspNodes) SetRequestID(id []byte) {
	p.ReqID = id
}
