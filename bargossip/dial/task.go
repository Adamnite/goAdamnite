package dial

import (
	"github.com/adamnite/go-adamnite/bargossip/admnode"

	log "github.com/sirupsen/logrus"
)

type Task struct {
	destNode       *admnode.GossipNode
	connectionFlag ConnectionFlag
}

func newTask(destNode *admnode.GossipNode, flag ConnectionFlag) *Task {
	return &Task{
		destNode:       destNode,
		connectionFlag: flag,
	}
}

func (t *Task) Start(s *Scheduler) {
	conn, err := s.dialer.Dial(s.ctx, t.destNode)
	if err != nil {
		log.Debug("Adamnite gossip error", "id", t.destNode.ID(), "ip", t.destNode.IP(), "err", err)
		return
	}
	s.addConnectionFunc(conn, OutboundConnection, t.destNode)
}
