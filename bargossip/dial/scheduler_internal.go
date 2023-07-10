package dial

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

func (s *Scheduler) freeDialConnections() int {
	freeConncections := s.MaxOutboundConnections - s.dialedPeers
	if freeConncections > s.MaxPendingOutboundConnections {
		freeConncections = s.MaxPendingOutboundConnections
	}
	freeConncections = freeConncections - len(s.dialing)
	return freeConncections
}

func (s *Scheduler) isValidateDialNode(node *admnode.GossipNode) error {
	if *node.ID() == s.SelfID {
		return errDialSelfNode
	}

	if node.IP() != nil && node.TCP() == 0 {
		return errDialNoPort
	}

	if _, task := s.dialing[*node.ID()]; task {
		return errDialAlreadyExsit
	}

	if _, peer := s.peers[*node.ID()]; peer {
		return errAlreadyConnected
	}

	if s.PeerBlackList != nil && s.PeerBlackList.Contains(node.IP()) {
		return errBlackListNode
	}

	if s.PeerWhiteList != nil && !s.PeerWhiteList.Contains(node.IP()) {
		return errNotWhiteListNode
	}

	if s.history.Contains(node.ID().String()) {
		return errDialedRecently
	}

	return nil
}

// dialNode run the tcp dial to connect the node
func (s *Scheduler) dialNode(task *Task) {
	s.Log.Debug("Starting Adamnite gossip", "id", task.destNode.ID(), "ip", task.destNode.IP())
	s.dialing[*task.destNode.ID()] = task
	s.history.Add(task.destNode.ID().String(), s.Clock.Now().Add(defaultDialHisotryExpiration))

	go func() {
		task.Start(s)
		s.dialDoneCh <- task
	}()
}
