package dial

import (
	"context"
	"net"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

type Dialer interface {
	Dial(context.Context, *admnode.GossipNode) (net.Conn, error)
}

type tcpDialer struct {
	dialer *net.Dialer
}

func (t tcpDialer) Dial(ctx context.Context, node *admnode.GossipNode) (net.Conn, error) {
	addr := &net.TCPAddr{IP: node.IP(), Port: int(node.TCP())}
	return t.dialer.DialContext(ctx, "tcp", addr.String())
}
