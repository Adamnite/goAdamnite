package findnode

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/admpacket"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
)

// UDPLayer is the implementation of the node discovery protocol.
type UDPLayer struct {
	conn      *net.UDPConn
	privKey   *ecdsa.PrivateKey
	localNode *admnode.LocalNode
	nodeTable *NodePool
	log       log15.Logger
	clock     mclock.Clock

	// channels
	callCh           chan *CallRsp
	callDoneCh       chan *CallRsp
	rspCallTimeoutCh chan *callTimeout
	readCh           chan struct{}
	packetInCh       chan ReadPacket

	// state of background thread
	callQueue                  map[admnode.NodeID][]*CallRsp
	activeCallQueue            map[admnode.NodeID]*CallRsp
	activeCallRequireAuthQueue map[admpacket.Nonce]*CallRsp
	sslencoding                *admpacket.SSLCodec

	// terminate items
	closeCtx       context.Context
	cancelCloseCtx context.CancelFunc
	wg             sync.WaitGroup
	closeOnce      sync.Once
}

// CallRsp represents a call response from another node.
type CallRsp struct {
	node            *admnode.GossipNode
	packet          admpacket.ADMPacket
	expectedRspType byte
	requestID       []byte

	//channels
	err    chan error
	respCh chan admpacket.ADMPacket

	nonce          admpacket.Nonce
	handshake      *admpacket.AskHandshake
	handshakeCount int
	timeout        mclock.Timer
}

// ReadPacket is a packet that couldn't be handled.
type ReadPacket struct {
	Data []byte
	Addr *net.UDPAddr
}

// callTimeout is the response timeout event of a call.
type callTimeout struct {
	callRsp *CallRsp
	timer   mclock.Timer
}

// Start runs the findnode module and listen the connection.
func Start(conn *net.UDPConn, localNode *admnode.LocalNode, cfg Config) (*UDPLayer, error) {
	closeCtx, cancelCloseCtx := context.WithCancel(context.Background())
	cfg = cfg.defaults()

	udpLayer := &UDPLayer{
		conn:      conn,
		localNode: localNode,
		log:       cfg.Log,
		clock:     cfg.Clock,
		privKey:   cfg.PrivateKey,

		//channels
		callCh:           make(chan *CallRsp),
		rspCallTimeoutCh: make(chan *callTimeout),
		readCh:           make(chan struct{}, 1),
		packetInCh:       make(chan ReadPacket, 1),

		// state of background thread
		callQueue:                  make(map[admnode.NodeID][]*CallRsp),
		activeCallQueue:            make(map[admnode.NodeID]*CallRsp),
		activeCallRequireAuthQueue: make(map[admpacket.Nonce]*CallRsp),
		sslencoding:                admpacket.NewSSLCodec(localNode, cfg.Clock, cfg.PrivateKey),

		// terminate items
		closeCtx:       closeCtx,
		cancelCloseCtx: cancelCloseCtx,
	}

	table, err := newNodePool(udpLayer, udpLayer.localNode.Database(), cfg.Bootnodes, cfg.Log)
	if err != nil {
		return nil, err
	}

	udpLayer.nodeTable = table

	go udpLayer.nodeTable.backgroundThread()

	udpLayer.wg.Add(2)
	go udpLayer.readThread()
	go udpLayer.backgroundThread()
	return udpLayer, nil
}

func (n *UDPLayer) SelfNode() *admnode.GossipNode {
	return n.localNode.Node()
}

func (n *UDPLayer) findSelfNode() []*admnode.GossipNode {
	return newFind(n.closeCtx, n.nodeTable, *n.SelfNode().ID(), func(nd *node) ([]*node, error) {
		return n.findFunc(nd, *n.SelfNode().ID())
	}).run()
}

func (n *UDPLayer) findRandomNodes() []*admnode.GossipNode {
	var target admnode.NodeID
	crand.Read(target[:])
	return newFind(n.closeCtx, n.nodeTable, target, func(nd *node) ([]*node, error) {
		return n.findFunc(nd, target)
	}).run()
}

func (n *UDPLayer) findFunc(destNode *node, target admnode.NodeID) ([]*node, error) {
	var (
		dists         = n.findDistances(target, *destNode.ID())
		distanceNodes = nodes{targetId: target}
	)

	resNodes, err := n.findNode(&destNode.GossipNode, dists)
	if err == errClosed {
		return nil, err
	}
	for _, wnode := range resNodes {
		if wnode.ID() != n.SelfNode().ID() {
			distanceNodes.push(&node{GossipNode: *wnode}, 16)
		}
	}
	return distanceNodes.nodes, err
}

func (n *UDPLayer) findDistances(target, dest admnode.NodeID) (dists []uint) {
	distance := admnode.LogDist(target, dest)
	dists = append(dists, uint(distance))

	for i := 1; len(dists) < 3; i++ {
		if distance+i < 256 {
			dists = append(dists, uint(distance+i))
		}
		if distance-i > 0 {
			dists = append(dists, uint(distance-i))
		}
	}
	return dists
}

func (n *UDPLayer) findNode(node *admnode.GossipNode, distances []uint) ([]*admnode.GossipNode, error) {
	resp := n.call(node, admpacket.RspFindnodeMsg, &admpacket.Findnode{Distances: distances})
	return n.waitForNodes(resp, distances)
}

func (n *UDPLayer) call(node *admnode.GossipNode, responseType byte, packet admpacket.ADMPacket) *CallRsp {
	callRsp := &CallRsp{
		node:            node,
		packet:          packet,
		expectedRspType: responseType,
		requestID:       make([]byte, 8),
		//channels
		err: make(chan error),
	}

	crand.Read(callRsp.requestID)
	packet.SetRequestID(callRsp.requestID)
	select {
	case n.callCh <- callRsp:
	case <-n.closeCtx.Done():
		callRsp.err <- errClosed
	}
	return callRsp
}

// backgroundThread runs in its own goroutine, handles incoming udp packets and deals with calls.
func (n *UDPLayer) backgroundThread() {
	defer n.wg.Done()

	n.readCh <- struct{}{}

	for {
		select {
		case callCh := <-n.callCh:
			id := callCh.node.ID()
			n.callQueue[*id] = append(n.callQueue[*id], callCh)
			n.sendCallToQueueAndRun(*id)
		case p := <-n.packetInCh:
			n.handlePacket(p)
			n.readCh <- struct{}{}
		case callTimeout := <-n.rspCallTimeoutCh:
			activeCall := n.activeCallQueue[*callTimeout.callRsp.node.ID()]
			if callTimeout.callRsp == activeCall && callTimeout.timer == activeCall.timeout {
				callTimeout.callRsp.err <- errTimeout
			}
		case call := <-n.callDoneCh:
			id := call.node.ID()
			activeCall := n.activeCallQueue[*id]
			if activeCall != call {
				panic("calldone for inactive call")
			}
			call.timeout.Stop()
			delete(n.activeCallQueue, *call.node.ID())
			delete(n.activeCallRequireAuthQueue, call.nonce)
			n.sendCallToQueueAndRun(*id)
		case <-n.closeCtx.Done():

		}
	}
}

func (n *UDPLayer) readThread() {
	defer n.wg.Done()

	buf := make([]byte, maxPacketSize)
	for range n.readCh {
		nbytes, from, err := n.conn.ReadFromUDP(buf)
		if IsTemporaryError(err) {
			n.log.Debug("Temporary UDP read error", "err", err)
			continue
		} else if err != nil {
			if err != io.EOF {
				n.log.Debug("UDP read error", "err", err)
			}
			return
		}
		n.readPacket(from, buf[:nbytes])
	}
}

func (n *UDPLayer) handlePacket(p ReadPacket) error {
	fromNodeID, fromNode, packet, err := n.sslencoding.Decode(p.Data, p.Addr.String())
	if err != nil {
		n.log.Debug("bad packet", "id", fromNodeID, "addr", p.Addr.String(), "err", err)
		return err
	}

	if fromNode != nil {
		n.nodeTable.addSeenNode(&node{GossipNode: *fromNode})
	}

	switch pkt := packet.(type) {
	case *admpacket.AskHandshake:
		n.handleAskHandshakePacket(pkt, fromNodeID, p.Addr)
	case *admpacket.SYN:
		n.handleSYNPacket(pkt, fromNodeID, p.Addr)
	case *admpacket.Findnode:
		n.handleFindNode(pkt, fromNodeID, p.Addr)
	case *admpacket.RspNodes:
		n.handleCallResponse(pkt, fromNodeID, p.Addr)
	}
	return nil
}

func (n *UDPLayer) readPacket(from *net.UDPAddr, data []byte) bool {
	select {
	case n.packetInCh <- ReadPacket{Addr: from, Data: data}:
		return true
	case <-n.closeCtx.Done():
		return false
	}
}

// sendCallToQueueAndRun sends a call to the queue and run the call if there is no active call.
func (n *UDPLayer) sendCallToQueueAndRun(id admnode.NodeID) {
	queue := n.callQueue[id]
	if len(queue) == 0 || n.activeCallQueue[id] != nil {
		return
	}

	n.activeCallQueue[id] = queue[0]
	n.sendCall(n.activeCallQueue[id])
	if len(queue) == 1 {
		delete(n.callQueue, id)
	} else {
		copy(queue, queue[1:])
		n.callQueue[id] = queue[:len(queue)-1]
	}
}

// sendCall sends a call
func (n *UDPLayer) sendCall(c *CallRsp) {
	if c.nonce != (admpacket.Nonce{}) {
		delete(n.activeCallRequireAuthQueue, c.nonce)
	}

	ipAddr := &net.UDPAddr{IP: c.node.IP(), Port: int(c.node.UDP())}
	newNonce, _ := n.send(*c.node.ID(), ipAddr, c.packet, c.handshake)
	c.nonce = newNonce
	n.activeCallRequireAuthQueue[newNonce] = c
	n.waitResponseTimeout(c)
}

func (n *UDPLayer) waitResponseTimeout(c *CallRsp) {
	if c.timeout != nil {
		c.timeout.Stop()
	}

	var done = make(chan struct{})
	var timer mclock.Timer

	timer = n.clock.AfterFunc(udpLayerResponseTimeout, func() {
		<-done
		select {
		case n.rspCallTimeoutCh <- &callTimeout{callRsp: c, timer: timer}:
		case <-n.closeCtx.Done():
		}
	})
	c.timeout = timer
	close(done)
}

func (n *UDPLayer) send(toID admnode.NodeID, toAddr *net.UDPAddr, packet admpacket.ADMPacket, handshake *admpacket.AskHandshake) (admpacket.Nonce, error) {
	addr := toAddr.String()
	enc, nonce, err := n.sslencoding.Encode(toID, addr, packet, handshake)
	if err != nil {
		n.log.Warn("UDP Encoding >> "+packet.Name(), "id", toID, "udpAddr", addr, "err", err)
		return nonce, err
	}
	_, err = n.conn.WriteToUDP(enc, toAddr)
	n.log.Trace("UDP Sent Packet >> "+packet.Name(), "id", toID, "udpAddr", addr)
	return nonce, err
}

func (n *UDPLayer) sendWithoutHandshake(toId admnode.NodeID, toAddr *net.UDPAddr, packet admpacket.ADMPacket) error {
	_, err := n.send(toId, toAddr, packet, nil)
	return err
}

// waitForNodes waits for NODES responses to the given call.
func (n *UDPLayer) waitForNodes(c *CallRsp, distances []uint) ([]*admnode.GossipNode, error) {
	defer n.callDone(c)

	var (
		nodes           []*admnode.GossipNode
		seen            = make(map[admnode.NodeID]struct{})
		received, total = 0, -1
	)
	for {
		select {
		case responseP := <-c.respCh:
			response := responseP.(*admpacket.RspNodes)
			for _, nodeInfo := range response.Nodes {
				node, err := n.checkResponseNode(nodeInfo, distances, c, seen)
				if err != nil {
					n.log.Debug("Invalid nodeinfo in "+response.Name(), "id", c.node.ID(), "err", err)
					continue
				}
				nodes = append(nodes, node)
			}
			if total == -1 {
				total = int(response.Total)
				if total > findNodeRspNodeLimit {
					total = findNodeRspNodeLimit
				}
			}
			if received++; received == total {
				return nodes, nil
			}
		case err := <-c.err:
			return nodes, err
		}
	}
}

// callDone tells dispatch that the active call is done.
func (n *UDPLayer) callDone(c *CallRsp) {
	for {
		select {
		case <-c.respCh:
			// late response, discard.
		case <-c.err:
			// late error, discard.
		case n.callDoneCh <- c:
			return
		case <-n.closeCtx.Done():
			return
		}
	}
}

// getNode returns the node if it was registered on table or database
func (n *UDPLayer) getNode(id admnode.NodeID) *admnode.GossipNode {
	if node := n.nodeTable.getNode(id); node != nil {
		return node
	}

	if node := n.localNode.Database().Node(id); node != nil {
		return node
	}
	return nil
}

func (n *UDPLayer) checkResponseNode(nodeInfo *admnode.NodeInfo, distances []uint, callRsp *CallRsp, seen map[admnode.NodeID]struct{}) (*admnode.GossipNode, error) {
	node, err := admnode.New(nodeInfo)
	if err != nil {
		return nil, err
	}
	if err := utils.CheckRelayIP(callRsp.node.IP(), nodeInfo.GetIP()); err != nil {
		return nil, err
	}

	if distances != nil {
		found := false
		nodeDistance := admnode.LogDist(*callRsp.node.ID(), *nodeInfo.GetNodeID())
		for _, distance := range distances {
			if distance == uint(nodeDistance) {
				found = true
				break
			}
		}
		if found {
			return nil, errors.New("not required node in distance")
		}
	}

	if _, ok := seen[*nodeInfo.GetNodeID()]; ok {
		return nil, fmt.Errorf("dupplicate node")
	}
	seen[*nodeInfo.GetNodeID()] = struct{}{}
	return node, nil
}

func (n *UDPLayer) Close() {
	n.closeOnce.Do(func() {
		n.cancelCloseCtx()
		n.conn.Close()
		n.wg.Wait()
		n.nodeTable.close()
	})
}
