package p2p

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/metrics"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/p2p/enr"
	"github.com/adamnite/go-adamnite/rlp"
)

var (
	ErrShuttingDown = errors.New("shutting down")
)

const (
	baseProtocolVersion    = 1
	baseProtocolLength     = uint64(16)
	baseProtocolMaxMsgSize = 2 * 1024

	snappyProtocolVersion = 5

	pingInterval = 15 * time.Second
)

const (
	pingMsg      = 0x00
	pongMsg      = 0x01
	handshakeMsg = 0x02
	discMsg      = 0x03
)

type protoHandshake struct {
	Version    uint64
	Name       string
	ListenPort uint64
	ID         []byte
	Caps       []PeerCapability
}

type PeerEventType string

const (
	PeerEventTypeAdd     PeerEventType = "PeerAdd"
	PeerEventTypeDrop    PeerEventType = "PeerDrop"
	PeerEventTypeMsgSend PeerEventType = "MsgSend"
	PeerEventTypeMsgRecv PeerEventType = "MsgRecv"
)

type PeerEvent struct {
	Type          PeerEventType `json:"type"`
	Peer          enode.ID      `json:"peer"`
	Error         string        `json:"error,omitempty"`
	Protocol      string        `json:"protocol,omitempty"`
	MsgCode       *uint64       `json:"msgcode,omitempty"`
	MsgSize       *uint32       `json:"msgsize,omitempty"`
	LocalAddress  string        `json:"localaddr,omitempty"`
	RemoteAddress string        `json:"remoteaddr,omitempty"`
}

type Peer struct {
	rw      *conn
	log     log15.Logger
	running map[string]*protoRW
	created mclock.AbsTime

	wg       sync.WaitGroup
	protoErr chan error
	closed   chan struct{}
	disc     chan DiscReason

	events *event.Feed
}

type protoRW struct {
	Protocol
	in     chan Msg        // receives read messages
	closed <-chan struct{} // receives when peer is shutting down
	wstart <-chan struct{} // receives when write may start
	werr   chan<- error    // for write results
	offset uint64
	w      MsgWriter
}

func NewPeer(id enode.ID, name string, caps []PeerCapability) *Peer {
	pipe, _ := net.Pipe()
	node := enode.SignNull(new(enr.Record), id)
	conn := &conn{fd: pipe, transport: nil, node: node, caps: caps, name: name}
	peer := newPeer(log15.Root(), conn, nil)
	close(peer.closed) // ensures Disconnect doesn't block
	return peer
}

func newPeer(log log15.Logger, conn *conn, protocols []Protocol) *Peer {
	protomap := matchProtocols(protocols, conn.caps, conn)
	p := &Peer{
		rw:       conn,
		running:  protomap,
		created:  mclock.Now(),
		disc:     make(chan DiscReason),
		protoErr: make(chan error, len(protomap)+1), // protocols + pingLoop
		closed:   make(chan struct{}),
		log:      log.New("id", conn.node.ID(), "conn", conn.flags),
	}
	return p
}

func matchProtocols(protocols []Protocol, caps []PeerCapability, rw MsgReadWriter) map[string]*protoRW {
	sort.Sort(capsByNameAndVersion(caps))
	offset := baseProtocolLength
	result := make(map[string]*protoRW)

outer:
	for _, cap := range caps {
		for _, proto := range protocols {
			if proto.Name == cap.Name && proto.Version == cap.Version {
				// If an old protocol version matched, revert it
				if old := result[cap.Name]; old != nil {
					offset -= old.Length
				}
				// Assign the new match
				result[cap.Name] = &protoRW{Protocol: proto, offset: offset, in: make(chan Msg), w: rw}
				offset += proto.Length

				continue outer
			}
		}
	}
	return result
}

func (p *Peer) ID() enode.ID {
	return p.rw.node.ID()
}

func (p *Peer) Node() *enode.Node {
	return p.rw.node
}

func (p *Peer) Name() string {
	s := p.rw.name
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}

func (p *Peer) Fullname() string {
	return p.rw.name
}

func (p *Peer) Caps() []PeerCapability {
	return p.rw.caps
}

func (p *Peer) RunningCap(protocol string, versions []uint) bool {
	if proto, ok := p.running[protocol]; ok {
		for _, ver := range versions {
			if proto.Version == ver {
				return true
			}
		}
	}
	return false
}

func (p *Peer) RemoteAddr() net.Addr {
	return p.rw.fd.RemoteAddr()
}

func (p *Peer) LocalAddr() net.Addr {
	return p.rw.fd.LocalAddr()
}

func (p *Peer) Disconnect(reason DiscReason) {
	select {
	case p.disc <- reason:
	case <-p.closed:
	}
}

func (p *Peer) String() string {
	id := p.ID()
	return fmt.Sprintf("Peer %x %v", id[:8], p.RemoteAddr())
}

func (p *Peer) Inbound() bool {
	return p.rw.is(inboundConn)
}

func (p *Peer) Log() log15.Logger {
	return p.log
}

func (p *Peer) run() (remoteRequested bool, err error) {
	var (
		writeStart = make(chan struct{}, 1)
		writeErr   = make(chan error, 1)
		readErr    = make(chan error, 1)
		reason     DiscReason // sent to the peer
	)
	p.wg.Add(2)
	go p.readLoop(readErr)
	go p.pingLoop()

	// Start all protocol handlers.
	writeStart <- struct{}{}
	p.startProtocols(writeStart, writeErr)

	// Wait for an error or disconnect.
loop:
	for {
		select {
		case err = <-writeErr:
			// A write finished. Allow the next write to start if
			// there was no error.
			if err != nil {
				reason = DiscNetworkError
				break loop
			}
			writeStart <- struct{}{}
		case err = <-readErr:
			if r, ok := err.(DiscReason); ok {
				remoteRequested = true
				reason = r
			} else {
				reason = DiscNetworkError
			}
			break loop
		case err = <-p.protoErr:
			reason = discReasonForError(err)
			break loop
		case err = <-p.disc:
			reason = discReasonForError(err)
			break loop
		}
	}

	close(p.closed)
	p.rw.close(reason)
	p.wg.Wait()
	return remoteRequested, err
}

func (p *Peer) pingLoop() {
	ping := time.NewTimer(pingInterval)
	defer p.wg.Done()
	defer ping.Stop()
	for {
		select {
		case <-ping.C:
			if err := SendItems(p.rw, pingMsg); err != nil {
				p.protoErr <- err
				return
			}
			ping.Reset(pingInterval)
		case <-p.closed:
			return
		}
	}
}

func (p *Peer) readLoop(errc chan<- error) {
	defer p.wg.Done()
	for {
		msg, err := p.rw.ReadMsg()
		if err != nil {
			errc <- err
			return
		}
		msg.ReceivedAt = time.Now()
		if err = p.handle(msg); err != nil {
			errc <- err
			return
		}
	}
}

func (p *Peer) handle(msg Msg) error {
	switch {
	case msg.Code == pingMsg:
		msg.Discard()
		go SendItems(p.rw, pongMsg)
	case msg.Code == discMsg:
		var reason [1]DiscReason
		// This is the last message. We don't need to discard or
		// check errors because, the connection will be closed after it.
		rlp.Decode(msg.Payload, &reason)
		return reason[0]
	case msg.Code < baseProtocolLength:
		// ignore other base protocol messages
		return msg.Discard()
	default:
		// it's a subprotocol message
		proto, err := p.getProto(msg.Code)
		if err != nil {
			return fmt.Errorf("msg code out of range: %v", msg.Code)
		}
		if metrics.Enabled {
			m := fmt.Sprintf("%s/%s/%d/%#02x", ingressMeterName, proto.Name, proto.Version, msg.Code-proto.offset)
			metrics.GetOrRegisterMeter(m, nil).Mark(int64(msg.meterSize))
			metrics.GetOrRegisterMeter(m+"/packets", nil).Mark(1)
		}
		select {
		case proto.in <- msg:
			return nil
		case <-p.closed:
			return io.EOF
		}
	}
	return nil
}

func countMatchingProtocols(protocols []Protocol, caps []PeerCapability) int {
	n := 0
	for _, cap := range caps {
		for _, proto := range protocols {
			if proto.Name == cap.Name && proto.Version == cap.Version {
				n++
			}
		}
	}
	return n
}

func (p *Peer) startProtocols(writeStart <-chan struct{}, writeErr chan<- error) {
	p.wg.Add(len(p.running))
	for _, proto := range p.running {
		proto := proto
		proto.closed = p.closed
		proto.wstart = writeStart
		proto.werr = writeErr
		var rw MsgReadWriter = proto
		if p.events != nil {
			rw = newMsgEventer(rw, p.events, p.ID(), proto.Name, p.Info().Network.RemoteAddress, p.Info().Network.LocalAddress)
		}
		p.log.Trace(fmt.Sprintf("Starting protocol %s/%d", proto.Name, proto.Version))
		go func() {
			defer p.wg.Done()
			err := proto.Run(p, rw)
			if err == nil {
				p.log.Trace(fmt.Sprintf("Protocol %s/%d returned", proto.Name, proto.Version))
				err = errProtocolReturned
			} else if err != io.EOF {
				p.log.Trace(fmt.Sprintf("Protocol %s/%d failed", proto.Name, proto.Version), "err", err)
			}
			p.protoErr <- err
		}()
	}
}

// getProto finds the protocol responsible for handling
// the given message code.
func (p *Peer) getProto(code uint64) (*protoRW, error) {
	for _, proto := range p.running {
		if code >= proto.offset && code < proto.offset+proto.Length {
			return proto, nil
		}
	}
	return nil, newPeerError(errInvalidMsgCode, "%d", code)
}

func (rw *protoRW) WriteMsg(msg Msg) (err error) {
	if msg.Code >= rw.Length {
		return newPeerError(errInvalidMsgCode, "not handled")
	}
	msg.meterCap = rw.cap()
	msg.meterCode = msg.Code

	msg.Code += rw.offset

	select {
	case <-rw.wstart:
		err = rw.w.WriteMsg(msg)
		// Report write status back to Peer.run. It will initiate
		// shutdown if the error is non-nil and unblock the next write
		// otherwise. The calling protocol code should exit for errors
		// as well but we don't want to rely on that.
		rw.werr <- err
	case <-rw.closed:
		err = ErrShuttingDown
	}
	return err
}

func (rw *protoRW) ReadMsg() (Msg, error) {
	select {
	case msg := <-rw.in:
		msg.Code -= rw.offset
		return msg, nil
	case <-rw.closed:
		return Msg{}, io.EOF
	}
}

type PeerInfo struct {
	ENR     string   `json:"enr,omitempty"` // Adamnite Node Record
	Enode   string   `json:"enode"`         // Node URL
	ID      string   `json:"id"`            // Unique node identifier
	Name    string   `json:"name"`          // Name of the node, including client type, version, OS, custom data
	Caps    []string `json:"caps"`          // Protocols advertised by this peer
	Network struct {
		LocalAddress  string `json:"localAddress"`  // Local endpoint of the TCP data connection
		RemoteAddress string `json:"remoteAddress"` // Remote endpoint of the TCP data connection
		Inbound       bool   `json:"inbound"`
		Trusted       bool   `json:"trusted"`
		Static        bool   `json:"static"`
	} `json:"network"`
	Protocols map[string]interface{} `json:"protocols"` // Sub-protocol specific metadata fields
}

// Info gathers and returns a collection of metadata known about a peer.
func (p *Peer) Info() *PeerInfo {
	// Gather the protocol capabilities
	var caps []string
	for _, cap := range p.Caps() {
		caps = append(caps, cap.String())
	}
	// Assemble the generic peer metadata
	info := &PeerInfo{
		Enode:     p.Node().URLv4(),
		ID:        p.ID().String(),
		Name:      p.Fullname(),
		Caps:      caps,
		Protocols: make(map[string]interface{}),
	}
	if p.Node().Seq() > 0 {
		info.ENR = p.Node().String()
	}
	info.Network.LocalAddress = p.LocalAddr().String()
	info.Network.RemoteAddress = p.RemoteAddr().String()
	info.Network.Inbound = p.rw.is(inboundConn)
	info.Network.Trusted = p.rw.is(trustedConn)
	info.Network.Static = p.rw.is(staticDialedConn)

	// Gather all the running protocol infos
	for _, proto := range p.running {
		protoInfo := interface{}("unknown")
		if query := proto.Protocol.PeerInfo; query != nil {
			if metadata := query(p.ID()); metadata != nil {
				protoInfo = metadata
			} else {
				protoInfo = "handshake"
			}
		}
		info.Protocols[proto.Name] = protoInfo
	}
	return info
}
