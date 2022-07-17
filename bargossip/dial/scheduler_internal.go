package dial

func (s *Scheduler) freeDialConnections() int {
	freeConncections := s.MaxOutboundConnections - s.dialedPeers
	if freeConncections > s.MaxPendingOutboundConnections {
		freeConncections = s.MaxPendingOutboundConnections
	}
	freeConncections = freeConncections - len(s.dialing)
	return freeConncections
}
