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
	"time"
	"bytes"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/admpacket"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/crypto"

	"github.com/vmihailenco/msgpack/v5"
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

	addReplyMatcher chan *replyMatcher

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

// replyMatcher represents a pending reply.
//
// Some implementations of the protocol wish to send more than one
// reply packet to findnode. In general, any neighbors packet cannot
// be matched up with a specific findnode packet.
//
// Our implementation handles this by storing a callback function for
// each pending reply. Incoming packets from a node are dispatched
// to all callback functions for that node.
type replyMatcher struct {
	// these fields must match in the reply.
	from  admnode.NodeID
	ip    net.IP
	ptype byte

	// time when the request must complete
	deadline time.Time

	// callback is called when a matching reply arrives. If it returns matched == true, the
	// reply was acceptable. The second return value indicates whether the callback should
	// be removed from the pending reply queue. If it returns false, the reply is considered
	// incomplete and the callback will be invoked again for the next matching reply.
	callback replyMatchFunc

	// errc receives nil when the callback indicates completion or an
	// error if no further reply is received within the timeout.
	errc chan error

	// reply contains the most recent reply. This field is safe for reading after errc has
	// received a value.
	reply admpacket.ADMPacket
}

type replyMatchFunc func(interface{}) (matched bool, requestDone bool)

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
		addReplyMatcher:  make(chan *replyMatcher),
		callDoneCh:		  make(chan *CallRsp),

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

func (n *UDPLayer) FindFunc(destNode *node, target admnode.NodeID) ([]*node, error) {
	return n.findFunc(destNode, target)
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
		respCh:          make(chan admpacket.ADMPacket),
		err: 			 make(chan error),
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
			return
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

	n.log.Trace("Packet << "+packet.Name(), "addr", p.Addr, "id", fromNodeID, "err", err)

	switch pkt := packet.(type) {
	case *admpacket.AskHandshake:
		n.handleAskHandshakePacket(pkt, fromNodeID, p.Addr)
	case *admpacket.SYN:
		n.handleSYNPacket(pkt, fromNodeID, p.Addr)
	case *admpacket.Findnode:
		n.handleFindNode(pkt, fromNodeID, p.Addr)
	case *admpacket.RspNodes:
		n.handleCallResponse(pkt, fromNodeID, p.Addr)
	case *admpacket.Ping:
		n.handlePingPacket(pkt, fromNodeID, p.Addr)
	case *admpacket.Pong:
		n.handlePongPacket(pkt, fromNodeID, p.Addr)
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
	n.log.Trace("Packet >> "+packet.Name(), "id", toID, "udpAddr", addr)
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
			n.log.Warn("Failed to wait find node", "Packet", c.packet.Name(), "err", err)
			return nil, err
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

func (l *UDPLayer) ping(n *admnode.GossipNode) (uint64, error) {
	remoteAddr := &net.UDPAddr{IP: n.IP(), Port: int(n.UDP())}
	reqPing := l.makePingPacket(remoteAddr)

	resp := l.call(n, admpacket.PongMsg, reqPing)
	return l.waitForPong(resp)

	/*
	rm := l.sendPing(*n.ID(), &net.UDPAddr{IP: n.IP(), Port: int(n.UDP())}, nil)
	if err = <-rm.errc; err == nil {
		// rm.reply.(*admpacket.Pong).PingHash
	}
	return nret, err;*/
}

// waitForPong waits for PONG responses to the given call.
func (n *UDPLayer) waitForPong(c *CallRsp) (uint64, error) {
	defer n.callDone(c)

	for {
		select {
		case responseP := <-c.respCh:
			response := responseP.(*admpacket.Pong)
			return uint64(response.Version), nil
		case err := <-c.err:
			n.log.Error("Packet >> " + c.packet.Name(), "err", err)
			return 0, err
		}
	}
}

func (l *UDPLayer) sendPing(remoteID admnode.NodeID, remoteAddr *net.UDPAddr, callback func()) *replyMatcher {
	reqPing := l.makePingPacket(remoteAddr)
	pkt, _, err := l.encode(l.privKey, reqPing)
	if err != nil {
		errc := make(chan error, 1)
		errc <- err
		return &replyMatcher{errc: errc}
	}

	rm := l.pending(remoteID, remoteAddr.IP, admpacket.PongMsg, func(p interface{}) (matched bool, requestDone bool) {
		matched = bytes.Equal(p.(*admpacket.Pong).ReqID, reqPing.ReqID)
		if matched && callback != nil {
			callback()
		}
		return matched, matched
	})
	// Send the packet.
	// l.localNode.UDPContact(remoteAddr)
	l.write(remoteAddr, remoteID, reqPing.Name(), pkt)
	return rm
}

func (l *UDPLayer) write(toaddr *net.UDPAddr, toid admnode.NodeID, what string, packet []byte) error {
	_, err := l.conn.WriteToUDP(packet, toaddr)
	l.log.Trace(">> "+what, "id", toid, "addr", toaddr, "err", err)
	return err
}

func (l *UDPLayer) pending(id admnode.NodeID, ip net.IP, pktType byte, callback replyMatchFunc) *replyMatcher {
	ch := make(chan error, 1)
	p := &replyMatcher{from: id, ip: ip, ptype: pktType, callback: callback, errc: ch}
	select {
	case l.addReplyMatcher <- p:
		// loop will handle it
	case <-l.closeCtx.Done():
		ch <- errClosed
	}
	return p
}

func (l *UDPLayer) encode(privKey *ecdsa.PrivateKey, reqPkt admpacket.ADMPacket) (byPacket, hash []byte, err error) {
	b := new(bytes.Buffer)
	b.Write(headSpace)
	b.WriteByte(reqPkt.MessageType())

	data, err := msgpack.Marshal(reqPkt)
	if err != nil {
		return nil, nil, err
	}

	b.Write(data)
	byPacket = b.Bytes()

	sig, err := crypto.Sign(crypto.Keccak256(byPacket[headSize:]), privKey)
	if err != nil {
		l.log.Error(fmt.Sprintf("Can't sign %s packet", reqPkt.Name()), "err", err)
		return nil, nil, err
	}

	copy(byPacket[hashSize:], sig)
	hash = crypto.Keccak256(byPacket[hashSize:])
	copy(byPacket, hash)
	return byPacket, hash, nil
}

func (l *UDPLayer) makePingPacket(remoteAddr *net.UDPAddr) *admpacket.Ping {
	return &admpacket.Ping {
		Version: pingPacketVersion,
		From: l.localEndpoint(),
		To: makePeerEndpoint(remoteAddr, 0),
		PubKey: l.localNode.GetPubKey(),
	}
}

func (l *UDPLayer) localEndpoint() admpacket.PeerEndpoint {
	ln := l.SelfNode()
	udpAddr := &net.UDPAddr{IP: ln.IP(), Port: int(ln.UDP())}

	return makePeerEndpoint(udpAddr, ln.TCP())
}

func (n *UDPLayer) Close() {
	n.closeOnce.Do(func() {
		n.cancelCloseCtx()
		n.conn.Close()
		n.wg.Wait()
		n.nodeTable.close()
	})
}

func makePeerEndpoint(addr *net.UDPAddr, port uint16) admpacket.PeerEndpoint {
	ip := net.IP{}
	if ip4 := addr.IP.To4(); ip4 != nil {
		ip = ip4
	} else if ip6 := addr.IP.To16(); ip6 != nil {
		ip = ip6
	}
	return admpacket.PeerEndpoint{IP: ip, UDP: uint16(addr.Port), TCP: port}
}