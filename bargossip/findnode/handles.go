package findnode

import (
	"bytes"
	crand "crypto/rand"
	"errors"
	"math"
	"net"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/admpacket"
	"github.com/adamnite/go-adamnite/p2p/netutil"
)

// handleSYNPacket make an askHandshake packet and send it to the node so that establishs handshake channel.
func (n *UDPLayer) handleSYNPacket(packet *admpacket.SYN, fromID admnode.NodeID, fromAddr *net.UDPAddr) {
	askHandshake := &admpacket.AskHandshake{
		Nonce:     packet.Nonce,
		DposRound: 0,
	}
	crand.Read(askHandshake.RandomID[:])

	if node := n.getNode(fromID); node != nil {
		askHandshake.Node = node
	}
	n.sendWithoutHandshake(fromID, fromAddr, askHandshake)
}

// handleAskHandshakePacket make an packet with handshake and send it to the node.
func (n *UDPLayer) handleAskHandshakePacket(packet *admpacket.AskHandshake, fromID admnode.NodeID, fromAddr *net.UDPAddr) {
	mcr := n.activeCallRequireAuthQueue[packet.Nonce]
	if mcr == nil {
		n.log.Debug("Invalid "+packet.Name(), "addr", fromAddr, "err", errors.New("no matching call"))
		return
	}

	if mcr.handshakeCount > 0 {
		n.log.Debug("Invalid "+packet.Name(), "addr", fromAddr, "err", errors.New("too many handshakes"))
		return
	}

	mcr.handshakeCount++
	mcr.handshake = packet
	packet.Node = mcr.node
	n.sendCall(mcr)
}

func (n *UDPLayer) handleFindNode(packet *admpacket.Findnode, fromID admnode.NodeID, fromAddr *net.UDPAddr) {
	var nodes []*admnode.GossipNode
	var processed = make(map[uint]struct{})

seek:
	for _, dist := range packet.Distances {
		_, seen := processed[dist]
		if seen || dist > 256 {
			continue
		}

		// Get the nodes.
		var bn []*admnode.GossipNode
		if dist == 0 {
			bn = []*admnode.GossipNode{n.SelfNode()}
		} else if dist <= 256 {
			n.nodeTable.mu.Lock()
			bn = unWrapFindNodes(n.nodeTable.getBucketAtDistance(int(dist)).whitelist)
			n.nodeTable.mu.Unlock()
		}
		processed[dist] = struct{}{}

		// Apply some pre-checks to avoid sending invalid nodes.
		for _, n := range bn {
			// TODO livenessChecks > 1
			if netutil.CheckRelayIP(fromAddr.IP, n.IP()) != nil {
				continue
			}
			nodes = append(nodes, n)
			if len(nodes) >= findNodeRspNodeLimit {
				break seek
			}
		}
	}

	if len(nodes) == 0 {
		n.sendWithoutHandshake(fromID, fromAddr, &admpacket.RspNodes{ReqID: packet.ReqID, Total: 1})
		return
	}

	total := uint8(math.Ceil(float64(len(nodes)) / 3))
	var resp []*admpacket.RspNodes
	for len(nodes) > 0 {
		p := &admpacket.RspNodes{ReqID: packet.ReqID, Total: total}

		items := findNodeRspNodeLimit
		if items < len(nodes) {
			items = len(nodes)
		}

		for i := 0; i < items; i++ {
			p.Nodes = append(p.Nodes, nodes[i].NodeInfo())
		}
		nodes = nodes[items:]
		resp = append(resp, p)
	}

	for _, respPacket := range resp {
		n.sendWithoutHandshake(fromID, fromAddr, respPacket)
	}
}

func (n *UDPLayer) handleCallResponse(packet admpacket.ADMPacket, fromID admnode.NodeID, fromAddr *net.UDPAddr) bool {
	call := n.activeCallQueue[fromID]
	if call == nil || !bytes.Equal(call.requestID, packet.RequestID()) {
		n.log.Debug("No activecall "+packet.Name(), "id", fromID, "addr", fromAddr)
		return false
	}
	if !fromAddr.IP.Equal(call.node.IP()) || fromAddr.Port != int(call.node.UDP()) {
		n.log.Debug("wrong endpoint "+packet.Name(), "id", fromID, "addr", fromAddr)
		return false
	}
	if packet.MessageType() != call.expectedRspType {
		n.log.Debug("Wrong response type "+packet.Name(), "id", fromID, "addr", fromAddr)
		return false
	}
	n.waitResponseTimeout(call)
	call.respCh <- packet
	return true
}
