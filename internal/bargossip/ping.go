package bargossip

import (
	"fmt"
	"io"
	"sync"

	p2p "github.com/adamnite/go-adamnite/internal/bargossip/pb"
	"github.com/adamnite/go-adamnite/log"
	proto "github.com/gogo/protobuf/proto"
	uuid "github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// pattern: /protocol-name/request-or-response-message/version
const pingRequest = "/ping/pingreq/0.0.1"
const pingResponse = "/ping/pingresp/0.0.1"

// PingProtocol type
type PingProtocol struct {
	node     *LocalNode // local host
	mu       sync.Mutex
	requests map[string]*p2p.PingRequest // used to access request data from response handlers. Protected by mu
	done     chan bool                   // only for demo purposes to stop main from terminating
}

func NewPingProtocol(node *LocalNode, done chan bool) *PingProtocol {
	log.Info("Starting PingProtocol", "ProtocolID", pingRequest)
	p := &PingProtocol{node: node, requests: make(map[string]*p2p.PingRequest), done: done}
	node.server.SetStreamHandler(pingRequest, p.onPingRequest)
	node.server.SetStreamHandler(pingResponse, p.onPingResponse)
	return p
}

// remote peer requests handler
func (p *PingProtocol) onPingRequest(s network.Stream) {

	// get request data
	data := &p2p.PingRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error("OnPingRequest occures error", "err", err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("OnPingRequest occures error", "err", err)
		return
	}

	log.Trace("Received ping message", "local", s.Conn().LocalPeer(), "from", s.Conn().RemotePeer(), "msg", data.Message)

	valid := p.node.authenticateMessage(data, data.MessageData)

	if !valid {
		log.Error("Failed to authenticate message")
		return
	}

	// generate response message
	log.Trace("Sending ping response", "local", s.Conn().LocalPeer(), "to", s.Conn().RemotePeer(), data.MessageData.Id)

	resp := &p2p.PingResponse{MessageData: p.node.NewMessageData(data.MessageData.Id, false),
		Message: fmt.Sprintf("Ping response from %s", p.node.server.ID())}

	// sign the data
	signature, err := p.node.signProtoMessage(resp)
	if err != nil {
		log.Error("Failed to sign response")
		return
	}

	// add the signature to the message
	resp.MessageData.Sign = signature

	// send the response
	ok := p.node.sendProtoMessage(s.Conn().RemotePeer(), pingResponse, resp)

	if ok {
		log.Trace("Ping response sent", "local", s.Conn().LocalPeer().String(), "to", s.Conn().RemotePeer().String())
	}
	p.done <- true
}

// remote ping response handler
func (p *PingProtocol) onPingResponse(s network.Stream) {
	data := &p2p.PingResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error("OnPingResponse occures error", "err", err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("OnPingResponse occures error", "err", err)
		return
	}

	valid := p.node.authenticateMessage(data, data.MessageData)

	if !valid {
		log.Error("Failed to authenticate message")
		return
	}

	// locate request data and remove it if found
	p.mu.Lock()
	_, ok := p.requests[data.MessageData.Id]
	if ok {
		// remove request from map as we have processed it here
		delete(p.requests, data.MessageData.Id)
	} else {
		log.Error("Failed to locate request data boject for response")
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	log.Debug("Received ping response", "local", s.Conn().LocalPeer(), "from", s.Conn().RemotePeer(), "msgid", data.MessageData.Id, "msg", data.Message)
	p.done <- true
}

func (p *PingProtocol) Ping(peerID peer.ID) bool {
	log.Debug("Sending ping", "From", p.node.server.ID(), "To", peerID)
	// create message data
	req := &p2p.PingRequest{MessageData: p.node.NewMessageData(uuid.New().String(), false),
		Message: fmt.Sprintf("Ping from %s", p.node.server.ID())}

	// sign the data
	signature, err := p.node.signProtoMessage(req)
	if err != nil {
		log.Error("failed to sign pb data", "err", err)
		return false
	}

	// add the signature to the message
	req.MessageData.Sign = signature

	// store ref request so response handler has access to it
	p.mu.Lock()
	p.requests[req.MessageData.Id] = req
	p.mu.Unlock()

	ok := p.node.sendProtoMessage(peerID, pingRequest, req)
	if !ok {
		return false
	}

	log.Debug("Ping message was sent", "To", peerID, "MsgID", req.MessageData.Id, "Msg", req.Message)
	return true
}
