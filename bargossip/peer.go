// Adamnite remote peer

package bargossip

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/event"

	log "github.com/sirupsen/logrus"
)

type Peer struct {
	peerConn *wrapPeerConnection
	created  mclock.AbsTime
	running  map[uint]*protoRW

	// channels
	closed      chan struct{}
	errChan     chan error
	readErrChan chan error

	wg sync.WaitGroup

	// events receives message send / receive events if set
	events *event.Feed
}

func newPeer(connection *wrapPeerConnection, protocols []SubProtocol) *Peer {
	protomap := matchProtocols(protocols, connection.protocol.ProtocolIDs, connection)
	peer := &Peer{
		peerConn:    connection,
		created:     mclock.Now(),
		running:     protomap,
		closed:      make(chan struct{}),
		errChan:     make(chan error),
		readErrChan: make(chan error),
	}

	return peer
}
func (p *Peer) ID() admnode.NodeID {
	return *p.peerConn.node.NodeInfo().GetNodeID()
}

// matchProtocols creates structures for matching named subprotocols.
func matchProtocols(protocols []SubProtocol, protoIDs []uint, rw MsgReadWriter) map[uint]*protoRW {
	result := make(map[uint]*protoRW)

outer:
	for _, protoID := range protoIDs {
		for _, proto := range protocols {
			if proto.ProtocolID == protoID {
				result[protoID] = &protoRW{SubProtocol: proto, in: make(chan Msg), w: rw}
				continue outer
			}
		}
	}
	return result
}

func (p *Peer) start() {
	var err error
	p.wg.Add(2)
	go p.pingThread()
	go p.readThread(p.readErrChan)

	// start all the sub protocols
	var writeStart = make(chan struct{}, 1)
	var writeErr = make(chan error, 1)
	writeStart <- struct{}{}
	p.startSubProtocols(writeStart, writeErr)

	close(p.closed)
	p.peerConn.close(err)
	p.wg.Wait()
}

func (p *Peer) startSubProtocols(writeStart <-chan struct{}, writeErr chan<- error) {
	p.wg.Add(len(p.running))
	for _, proto := range p.running {
		proto := proto
		proto.closed = p.closed
		proto.wstart = writeStart
		proto.werr = writeErr
		var rw MsgReadWriter = proto
		if p.events != nil {
			rw = newMsgEventer(rw, p.events, *p.peerConn.node.ID(), proto.ProtocolID, p.peerConn.conn.RemoteAddr().String(), p.peerConn.conn.LocalAddr().String())
		}
		log.Trace(fmt.Sprintf("Starting protocol %d", proto.ProtocolID))
		go func() {
			defer p.wg.Done()
			err := proto.Run(p, rw)
			if err == nil {
				log.Trace(fmt.Sprintf("Protocol %d returned", proto.ProtocolID))
				err = errProtocolReturned
			} else if err != io.EOF {
				log.Trace(fmt.Sprintf("Protocol %d failed", proto.ProtocolID), "err", err)
			}
			p.errChan <- err
		}()
	}
}

func (p *Peer) pingThread() {
	ping := time.NewTimer(pingInterval)
	defer p.wg.Done()
	defer ping.Stop()
	for {
		select {
		case <-ping.C:
			if err := SendItems(p.peerConn, pingMsg); err != nil {
				p.errChan <- err
				return
			}
			ping.Reset(pingInterval)
		case <-p.closed:
			return
		}
	}
}

func (p *Peer) readThread(errc chan<- error) {
	defer p.wg.Done()
	for {
		msg, err := p.peerConn.ReadMsg()
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
		go SendItems(p.peerConn, pongMsg)
	default:
		// it's a subprotocol message
		proto, err := p.getProto(msg.Code)
		if err != nil {
			return fmt.Errorf("msg code out of range: %v", msg.Code)
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

// getProto finds the protocol responsible for handling
// the given message code.
func (p *Peer) getProto(code uint64) (*protoRW, error) {
	for _, proto := range p.running {
		if code >= uint64(proto.ProtocolCodeOffset) && code < uint64(proto.ProtocolCodeOffset)+uint64(proto.ProtocolLength) {
			return proto, nil
		}
	}
	return nil, errors.New("invalid sub protocol message code")
}
