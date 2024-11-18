package dial

import (
	"context"
	"net"
	"sync"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/utils"
)

type AddConnection func(net.Conn, ConnectionFlag, *admnode.GossipNode) error

type Scheduler struct {
	Config
	wg                sync.WaitGroup
	nodeIterator      admnode.NodeIterator
	dialer            Dialer
	addConnectionFunc AddConnection

	ctx    context.Context
	cancel context.CancelFunc

	dialedPeers int // current number of dialed peers.
	dialing     map[admnode.NodeID]*Task
	peers       map[admnode.NodeID]ConnectionFlag

	// Channels
	nodeRecvCh chan *admnode.GossipNode
	dialDoneCh chan *Task

	// History
	history utils.InboundConnHeap
}

func New(config Config, it admnode.NodeIterator, addConnectionFunc AddConnection) *Scheduler {
	s := &Scheduler{
		Config:            config,
		nodeIterator:      it,
		dialer:            tcpDialer{dialer: &net.Dialer{Timeout: defaultDialTimeout}},
		addConnectionFunc: addConnectionFunc,

		// Channels
		nodeRecvCh: make(chan *admnode.GossipNode),
		dialDoneCh: make(chan *Task),
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Scheduler) Start() {
	s.wg.Add(2)
	s.getNodeToConnect()
	s.run()
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

// getNodeToConnect runs in its go routine and send the node to the channel.
func (s *Scheduler) getNodeToConnect() {
	defer s.wg.Done()

loop:
	for s.nodeIterator.Next() {
		select {
		case s.nodeRecvCh <- s.nodeIterator.Node():
		case <-s.ctx.Done():
			break loop
		}
	}
}

// This is the background go routine to process the dial schedule.
func (s *Scheduler) run() {
	defer s.wg.Done()

loop:
	for {
		freeConnections := s.freeDialConnections()

		select {
		case node := <-s.nodeRecvCh:
			if freeConnections > 0 {
				if err := s.isValidateDialNode(node); err != nil {
					s.Log.Debug("dial validate error", "id", node.ID(), "ip", node.IP(), "err", err)
				} else {
					s.dialNode(newTask(node, OutboundConnection))
				}
			}
		case task := <-s.dialDoneCh:
			delete(s.dialing, *task.destNode.ID())
			s.peers[*task.destNode.ID()] = OutboundConnection
		case <-s.ctx.Done():
			s.nodeIterator.Close()
			break loop
		}
	}
}
