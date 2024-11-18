package admnode

type NodeIterator interface {
	Next() bool
	Node() *GossipNode
	Close()
}
