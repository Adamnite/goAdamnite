package dial

import (
	"context"
	"sync"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

type Scheduler struct {
	Config
	wg           sync.WaitGroup
	nodeIterator admnode.NodeIterator

	ctx    context.Context
	cancel context.CancelFunc

	dialedPeers int // current number of dialed peers.
	dialing     map[admnode.NodeID]*Task

	// Channels
	nodeRecvCh chan *admnode.GossipNode
}

func New(config Config, it admnode.NodeIterator) *Scheduler {
	s := &Scheduler{
		Config:       config,
		nodeIterator: it,

		// Channels
		nodeRecvCh: make(chan *admnode.GossipNode),
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

			}
		case <-s.ctx.Done():
			s.nodeIterator.Close()
			break loop
		}
	}
}
