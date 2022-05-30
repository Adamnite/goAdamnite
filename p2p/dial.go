package p2p

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/p2p/netutil"
)

const (
	dialStatsLogInterval = 120 * time.Second
	dialStatsPeerLimit   = 2

	initialResolveDelay = 60 * time.Second
	maxResolveDelay     = time.Hour
)

var (
	errSelf             = errors.New("is self")
	errAlreadyDialing   = errors.New("already dialing")
	errAlreadyConnected = errors.New("already connected")
	errAlreadyListened  = errors.New("already listened")
	errRecentlyDialed   = errors.New("recently dialed")
	errNotWhitelisted   = errors.New("not contained in netrestrict whitelist")
	errNoPort           = errors.New("node does not provide TCP port")
)

type AdamniteNodeDialer interface {
	Dial(context.Context, *enode.Node) (net.Conn, error)
}

type tcpDialer struct {
	dialer *net.Dialer
}

func (tcp tcpDialer) Dial(ctx context.Context, dest *enode.Node) (net.Conn, error) {
	return tcp.dialer.DialContext(ctx, "tcp", nodeAddr(dest).String())
}

func nodeAddr(n *enode.Node) net.Addr {
	return &net.TCPAddr{IP: n.IP(), Port: n.TCP()}
}

type NodeResolver interface {
	Resolve(*enode.Node) *enode.Node
}

type dialConfig struct {
	self           enode.ID
	maxDialPeers   int
	maxActiveDials int
	netRestrict    *netutil.Netlist
	resolver       NodeResolver
	dialer         AdamniteNodeDialer
	log            log15.Logger
	clock          mclock.Clock
	rand           *mrand.Rand
}

// A dialTask generated for each node that is dialed.
type dialTask struct {
	staticPoolIndex int
	flags           connFlag
	// These fields are private to the task and should not be
	// accessed by dialScheduler while the task is running.
	dest         *enode.Node
	lastResolved mclock.AbsTime
	resolveDelay time.Duration
}

type dialSetupFunc func(net.Conn, connFlag, *enode.Node) error

type dialScheduler struct {
	dialConfig
	setupFunc dialSetupFunc
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	ctx       context.Context

	nodesIn        chan *enode.Node
	doneCh         chan *dialTask
	addStaticCh    chan *enode.Node
	removeStaticCh chan *enode.Node
	addPeerCh      chan *conn
	removePeerCh   chan *conn

	peers     map[enode.ID]connFlag  // all connected peers
	dialPeers int                    // Current number of dialed peers
	dialing   map[enode.ID]*dialTask // active tasks

	static     map[enode.ID]*dialTask
	staticPool []*dialTask

	lastStatsLog     mclock.AbsTime
	doneSinceLastLog int
}

func (cfg dialConfig) withDefaults() dialConfig {
	if cfg.maxActiveDials == 0 {
		cfg.maxActiveDials = defaultMaxPendingPeers
	}
	if cfg.log == nil {
		cfg.log = log15.Root()
	}
	if cfg.clock == nil {
		cfg.clock = mclock.System{}
	}
	if cfg.rand == nil {
		seed := make([]byte, 8)
		crand.Read(seed)
		seedNum := int64(binary.BigEndian.Uint64(seed))
		cfg.rand = mrand.New(mrand.NewSource(seedNum))
	}

	return cfg
}

func newDialScheduler(config dialConfig, it enode.Iterator, setupFunc dialSetupFunc) *dialScheduler {
	d := &dialScheduler{
		dialConfig:     config.withDefaults(),
		setupFunc:      setupFunc,
		dialing:        make(map[enode.ID]*dialTask),
		static:         make(map[enode.ID]*dialTask),
		peers:          make(map[enode.ID]connFlag),
		doneCh:         make(chan *dialTask),
		nodesIn:        make(chan *enode.Node),
		addStaticCh:    make(chan *enode.Node),
		removeStaticCh: make(chan *enode.Node),
		addPeerCh:      make(chan *conn),
		removePeerCh:   make(chan *conn),
	}
	d.lastStatsLog = d.clock.Now()
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.wg.Add(2)
	go d.readNodes(it)
	go d.loop(it)
	return d
}

func (d *dialScheduler) stop() {
	d.cancel()
	d.wg.Wait()
}

func (d *dialScheduler) addStatic(n *enode.Node) {
	select {
	case d.addStaticCh <- n:
	case <-d.ctx.Done():
	}
}

func (d *dialScheduler) removeStatic(n *enode.Node) {
	select {
	case d.removeStaticCh <- n:
	case <-d.ctx.Done():
	}
}

func (d *dialScheduler) peerAdded(c *conn) {
	select {
	case d.addPeerCh <- c:
	case <-d.ctx.Done():
	}
}

func (d *dialScheduler) peerRemoved(c *conn) {
	select {
	case d.removePeerCh <- c:
	case <-d.ctx.Done():
	}
}

func (d *dialScheduler) readNodes(it enode.Iterator) {
	defer d.wg.Done()

	for it.Next() {
		select {
		case d.nodesIn <- it.Node():
		case <-d.ctx.Done():
		}
	}
}

func (d *dialScheduler) logStats() {
	now := d.clock.Now()
	if d.lastStatsLog.Add(dialStatsLogInterval) > now {
		return
	}

	if d.dialPeers < dialStatsPeerLimit && d.dialPeers < d.maxDialPeers {
		d.log.Info("Looking for peers", "peercount", len(d.peers), "tried", d.doneSinceLastLog, "static", len(d.static))
	}

	d.doneSinceLastLog = 0
	d.lastStatsLog = now
}

func (d *dialScheduler) freeDialSlots() int {
	slots := (d.maxDialPeers - d.dialPeers) * 2

	if slots > d.maxActiveDials {
		slots = d.maxActiveDials
	}

	free := slots - len(d.dialing)
	return free
}

func (d *dialScheduler) startStaticDials(slots int) (started int) {
	for started = 0; started < slots && len(d.staticPool) > 0; started++ {
		idx := d.rand.Intn(len(d.staticPool))
		task := d.staticPool[idx]
		d.startDial(task)
		d.removeFromStaticPool(idx)
	}
	return started
}

func (d *dialScheduler) startDial(task *dialTask) {
	d.log.Trace("Starting p2p dial", "id", task.dest.ID(), "ip", task.dest.IP(), "flag", task.flags)
	// hkey := string(task.dest.ID().Bytes())
	d.dialing[task.dest.ID()] = task
	go func() {
		task.run(d)
		d.doneCh <- task
	}()
}

func (d *dialScheduler) removeFromStaticPool(idx int) {
	task := d.staticPool[idx]
	end := len(d.staticPool) - 1
	d.staticPool[idx] = d.staticPool[end]
	d.staticPool[idx].staticPoolIndex = idx
	d.staticPool[end] = nil
	d.staticPool = d.staticPool[:end]
	task.staticPoolIndex = -1
}

func (d *dialScheduler) checkDial(n *enode.Node) error {
	if n.ID() == d.self {
		return errSelf
	}
	if n.IP() != nil && n.TCP() == 0 {
		return errNoPort
	}
	if _, ok := d.dialing[n.ID()]; ok {
		return errAlreadyDialing
	}
	if _, ok := d.peers[n.ID()]; ok {
		return errAlreadyConnected
	}
	if d.netRestrict != nil && !d.netRestrict.Contains(n.IP()) {
		return errNotWhitelisted
	}
	return nil
}

func (d *dialScheduler) updateStaticPool(id enode.ID) {
	task, ok := d.static[id]
	if ok && task.staticPoolIndex < 0 && d.checkDial(task.dest) == nil {
		d.addToStaticPool(task)
	}
}

func (d *dialScheduler) addToStaticPool(task *dialTask) {
	if task.staticPoolIndex >= 0 {
		panic("attempt to add task to staticPool twice")
	}
	d.staticPool = append(d.staticPool, task)
	task.staticPoolIndex = len(d.staticPool) - 1
}

func (d *dialScheduler) loop(it enode.Iterator) {
	var (
		nodesCh chan *enode.Node
	)

loop:
	for {
		slots := d.freeDialSlots()
		slots -= d.startStaticDials(slots)

		if slots > 0 {
			nodesCh = d.nodesIn
		} else {
			nodesCh = nil
		}

		d.logStats()

		select {
		case node := <-nodesCh:
			if err := d.checkDial(node); err != nil {
				d.log.Trace("Discarding dial candidate", "id", node.ID(), "ip", node.IP(), "reason", err)
			} else {
				d.startDial(newDialTask(node, dynDialedConn))
			}
		case task := <-d.doneCh:
			id := task.dest.ID()
			delete(d.dialing, id)
			d.updateStaticPool(id)
			d.doneSinceLastLog++
		case conn := <-d.addPeerCh:
			if conn.is(dynDialedConn) || conn.is(staticDialedConn) {
				d.dialPeers++
			}
			id := conn.node.ID()
			d.peers[id] = conn.flags
			task := d.static[id]
			if task != nil && task.staticPoolIndex >= 0 {
				d.removeFromStaticPool(task.staticPoolIndex)
			}
		case conn := <-d.removePeerCh:
			if conn.is(dynDialedConn) || conn.is(staticDialedConn) {
				d.dialPeers--
			}
			delete(d.peers, conn.node.ID())
			d.updateStaticPool(conn.node.ID())

		case node := <-d.addStaticCh:
			id := node.ID()
			_, exists := d.static[id]
			d.log.Trace("Adding static node", "id", id, "ip", node.IP(), "added", !exists)
			if exists {
				continue loop
			}
			task := newDialTask(node, staticDialedConn)
			d.static[id] = task
			if d.checkDial(node) == nil {
				d.addToStaticPool(task)
			}

		case node := <-d.removeStaticCh:
			id := node.ID()
			task := d.static[id]
			d.log.Trace("Removing static node", "id", id, "ok", task != nil)
			if task != nil {
				delete(d.static, id)
				if task.staticPoolIndex >= 0 {
					d.removeFromStaticPool(task.staticPoolIndex)
				}
			}

		case <-d.ctx.Done():
			it.Close()
			break loop
		}
	}

	for range d.dialing {
		<-d.doneCh
	}
	d.wg.Done()
}

type dialError struct {
	error
}

func newDialTask(dest *enode.Node, flags connFlag) *dialTask {
	return &dialTask{dest: dest, flags: flags, staticPoolIndex: -1}
}

func (t *dialTask) run(d *dialScheduler) {
	if t.needResolve() && !t.resolve(d) {
		return
	}

	err := t.dial(d, t.dest)
	if err != nil {
		if _, ok := err.(*dialError); ok && t.flags&staticDialedConn != 0 {
			if t.resolve(d) {
				t.dial(d, t.dest)
			}
		}
	}
}

func (t *dialTask) needResolve() bool {
	return t.flags&staticDialedConn != 0 && t.dest.IP() == nil
}

func (t *dialTask) resolve(d *dialScheduler) bool {
	if d.resolver == nil {
		return false
	}
	if t.resolveDelay == 0 {
		t.resolveDelay = initialResolveDelay
	}
	if t.lastResolved > 0 && time.Duration(d.clock.Now()-t.lastResolved) < t.resolveDelay {
		return false
	}
	resolved := d.resolver.Resolve(t.dest)
	t.lastResolved = d.clock.Now()
	if resolved == nil {
		t.resolveDelay *= 2
		if t.resolveDelay > maxResolveDelay {
			t.resolveDelay = maxResolveDelay
		}
		d.log.Debug("Resolving node failed", "id", t.dest.ID(), "newdelay", t.resolveDelay)
		return false
	}

	t.resolveDelay = initialResolveDelay
	t.dest = resolved
	d.log.Debug("Resolved node", "id", t.dest.ID(), "addr", &net.TCPAddr{IP: t.dest.IP(), Port: t.dest.TCP()})
	return true
}

func (t *dialTask) dial(d *dialScheduler, dest *enode.Node) error {
	fd, err := d.dialer.Dial(d.ctx, t.dest)
	if err != nil {
		d.log.Trace("Dial error", "id", t.dest.ID(), "addr", nodeAddr(t.dest), "conn", t.flags, "err", cleanupDialErr(err))
		return &dialError{err}
	}
	mfd := newMeteredConn(fd, false, &net.TCPAddr{IP: dest.IP(), Port: dest.TCP()})
	return d.setupFunc(mfd, t.flags, dest)
}

func (t *dialTask) String() string {
	id := t.dest.ID()
	return fmt.Sprintf("%v %x %v:%d", t.flags, id[:8], t.dest.IP(), t.dest.TCP())
}

func cleanupDialErr(err error) error {
	if netErr, ok := err.(*net.OpError); ok && netErr.Op == "dial" {
		return netErr.Err
	}
	return err
}
