// A BAR Resilient P2P network server for distributed ledger system

package bargossip

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/bargossip/findnode"
	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/log15"
)

// Server manages the peer connections.
type Server struct {
	Config

	isRunning bool

	listener net.Listener
	log      log15.Logger

	localnode *admnode.LocalNode

	findNodeUdpLayer *findnode.UDPLayer

	lock   sync.Mutex
	loopWG sync.WaitGroup

	nodedb *admnode.NodeDB

	inboundConnHistory utils.InboundConnHeap

	// Channels
	quit chan struct{}
}

// Start starts the server.
func (srv *Server) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if srv.isRunning {
		return errors.New("adamnite p2p server is running")
	}

	srv.isRunning = true
	if err = srv.initialize(); err != nil {
		return err
	}

	srv.loopWG.Add(1)
	go srv.run()
	return nil
}

// Stop terminates the server and all active peer connections.
func (srv *Server) Stop() {
	srv.lock.Lock()
	if !srv.isRunning {
		srv.lock.Unlock()
		return
	}
	srv.isRunning = false
	if srv.listener != nil {
		srv.listener.Close()
	}
	close(srv.quit)
	srv.lock.Unlock()
	srv.loopWG.Wait()
}

// initialize initializes the gossip p2p server.
func (srv *Server) initialize() (err error) {
	srv.log = srv.Config.Logger
	if srv.log == nil {
		srv.log = log15.Root()
	}

	if srv.clock == nil {
		srv.clock = mclock.System{}
	}

	if srv.ListenAddr == "" {
		srv.log.Warn("adamnite p2p server listening address is not set")
	}

	if srv.ServerPrvKey == nil {
		return errors.New("adaminte p2p server private key must be set to a non-nil key")
	}

	srv.quit = make(chan struct{})

	if err := srv.initializeLocalNode(); err != nil {
		return err
	}

	return nil
}

// initializeLocalNode
func (srv *Server) initializeLocalNode() error {
	// Create the local node DB.
	db, err := admnode.OpenDB(srv.Config.NodeDatabase)
	if err != nil {
		return err
	}

	srv.nodedb = db
	srv.localnode = admnode.NewLocalNode(db, srv.ServerPrvKey)

	switch srv.NAT.(type) {
	case nil:
	case nat.ExtIP:
		ip, _ := srv.NAT.ExternalIP()
		srv.localnode.SetIP(ip)
	default:
		srv.loopWG.Add(1)
		go func() {
			defer srv.loopWG.Done()
			if ip, err := srv.NAT.ExternalIP(); err == nil {
				srv.localnode.SetIP(ip)
			}
		}()
	}

	// Launch the TCP listener.
	listener, err := net.Listen("tcp", srv.ListenAddr)
	if err != nil {
		return err
	}

	srv.listener = listener
	srv.ListenAddr = listener.Addr().String()

	if tcp, ok := listener.Addr().(*net.TCPAddr); ok {
		srv.localnode.SetTCP(uint16(tcp.Port))
		srv.log.Info("TCP listener", "addr", tcp)

		if !tcp.IP.IsLoopback() && srv.NAT != nil {
			srv.loopWG.Add(1)
			go func() {
				nat.Map(srv.NAT, srv.quit, "tcp", tcp.Port, tcp.Port, "adamnite p2p")
				srv.loopWG.Done()
			}()
		}
	}

	srv.loopWG.Add(1)
	go srv.listenThread()

	// Launch the UDP listener
	addr, err := net.ResolveUDPAddr("udp", srv.ListenAddr)
	if err != nil {
		return err
	}

	listeners, err := utils.FindUDPPortListeners(addr.Port)
	if err != nil {
		return err
	}
	if len(listeners) > 0 {
		err = errAlreadyListened
		srv.log.Error("UDP Port", "addr", addr, "err", err, "listeners", listeners)
		return err
	}

	udpListener, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	udpAddr := udpListener.LocalAddr().(*net.UDPAddr)
	srv.log.Info("UDP listener", "addr", udpAddr)

	if srv.NAT != nil && !udpAddr.IP.IsLoopback() {
		srv.loopWG.Add(1)
		go func() {
			nat.Map(srv.NAT, srv.quit, "udp", udpAddr.Port, udpAddr.Port, "admanite udp listener")
		}()
	}
	srv.localnode.SetUDP(uint16(udpAddr.Port))

	err = srv.initializeFindPeerModule(udpListener)
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) initializeFindPeerModule(listener *net.UDPConn) (err error) {
	findnodeCfg := &findnode.Config{
		PrivateKey:    srv.ServerPrvKey,
		PeerBlackList: srv.PeerBlackList,
		PeerWhiteList: srv.PeerWhiteList,
		Bootnodes:     srv.BootstrapNodes,
		Log:           srv.log,
		Clock:         srv.clock,
	}

	srv.findNodeUdpLayer, err = findnode.Start(listener, srv.localnode, *findnodeCfg)
	return err
}

// ***************************************************************************************************** //
// **************************** ADAMNITE P2P Server Common functions *********************************** //
// ***************************************************************************************************** //

// checkInboundConections check the ip address to accept connection.
func (srv *Server) checkInboundConnections(ip net.IP) error {
	if ip == nil {
		return nil
	}

	// Reject connections that do not match with witelist
	if srv.PeerWhiteList != nil && !srv.PeerWhiteList.Contains(ip) {
		return fmt.Errorf("not whitelisted peer")
	}

	// Reject connections that do match with blacklist
	if srv.PeerBlackList != nil && srv.PeerBlackList.Contains(ip) {
		return fmt.Errorf("blacklist peer")
	}

	// Reject peers that try too often
	now := srv.clock.Now()
	srv.inboundConnHistory.Expire(now, nil)
	if srv.inboundConnHistory.Contains(ip.String()) {
		return fmt.Errorf("too many attempts")
	}
	srv.inboundConnHistory.Add(ip.String(), now.Add(inboundAtemptDuration))
	return nil
}

// getNodeFromConn return the GossipNode from connection.
func getNodeFromConn(pubKey *ecdsa.PublicKey, conn net.Conn) *admnode.GossipNode {
	var ip net.IP
	var port uint16
	if tcp, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ip = tcp.IP
		port = uint16(tcp.Port)
	}
	return admnode.NewWithParams(pubKey, ip, port, port)
}

// ***************************************************************************************************** //
// ******************************** ADAMNITE P2P Server Threads **************************************** //
// ***************************************************************************************************** //

// run is a background thread
func (srv *Server) run() {
	srv.log.Info("Adamnite BAR-GOSSIP server started", "localnode", srv.localnode.NodeInfo().ToURL())
	defer srv.loopWG.Done()
	defer srv.nodedb.Close()

	srv.log.Debug("Adamnite BAR-GOSSIP is stopping now ")
	if srv.findNodeUdpLayer != nil {
		srv.findNodeUdpLayer.Close()
	}
}

// listenThread runs in its own goroutine and accepts inbound connections. (TCP listen thread)
func (srv *Server) listenThread() {
	defer srv.loopWG.Done()

	pendingConnections := defaultMaxPendingConnections
	if srv.MaxPendingConnections > 0 {
		pendingConnections = srv.MaxPendingConnections
	}

	pendingInboundConnSlots := make(chan struct{}, pendingConnections)
	for i := 0; i < pendingConnections; i++ {
		pendingInboundConnSlots <- struct{}{}
	}

	defer func() {
		for i := 0; i < cap(pendingInboundConnSlots); i++ {
			<-pendingInboundConnSlots
		}
	}()

	srv.log.Info("TCP listener started", "addr", srv.listener.Addr(), "inboundSlots", pendingConnections)

	for {
		<-pendingInboundConnSlots

		var peerConn net.Conn
		var err error

		for {
			peerConn, err = srv.listener.Accept()
			if findnode.IsTemporaryError(err) {
				srv.log.Debug("Peer packet temporary read error", "err", err)
				time.Sleep(time.Millisecond * 100)
				continue
			} else if err != nil {
				srv.log.Debug("Peer packet read error", "err", err)
				pendingInboundConnSlots <- struct{}{}
				return
			}
			break
		}

		remotePeerIP := utils.NetAddrToIP(peerConn.RemoteAddr())
		if err := srv.checkInboundConnections(remotePeerIP); err != nil {
			srv.log.Debug("Rejected inbound connection", "addr", peerConn.RemoteAddr(), "err", err)
			peerConn.Close()
			pendingInboundConnSlots <- struct{}{}
			continue
		}

		if remotePeerIP != nil {
			srv.log.Debug("Accepted connection", "addr", peerConn.RemoteAddr())
		}

		go func() {
			srv.AddConnection(peerConn, inboundConnection, nil)
			pendingInboundConnSlots <- struct{}{}
		}()
	}
}

func (srv *Server) AddConnection(peerConn net.Conn, connFlag connectionFlag, remotePeerNode *admnode.GossipNode) error {
	wrapPeerConn := wrapPeerConnection{conn: peerConn, connFlags: connFlag, chError: make(chan error)}

	if remotePeerNode != nil {
		wrapPeerConn.peerTransport = NewPeerTransport(peerConn, remotePeerNode.Pubkey())
	} else {
		wrapPeerConn.peerTransport = NewPeerTransport(peerConn, nil)
	}

	srv.lock.Lock()
	isRunning := srv.isRunning
	srv.lock.Unlock()
	if !isRunning {
		wrapPeerConn.peerTransport.close(errServerStopped)
		return errServerStopped
	}

	// Start handshake
	remotePubKey, err := wrapPeerConn.doHandshake(srv.ServerPrvKey)
	if err != nil {
		srv.log.Debug("Failed handshake", "addr", wrapPeerConn.conn.RemoteAddr(), "conn", wrapPeerConn.connFlags, "err", err)
		wrapPeerConn.peerTransport.close(errServerStopped)
		return err
	}

	if remotePeerNode != nil {
		wrapPeerConn.node = remotePeerNode
	} else {
		wrapPeerConn.node = getNodeFromConn(remotePubKey, peerConn)
	}

	return nil
}
