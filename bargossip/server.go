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
	"github.com/adamnite/go-adamnite/bargossip/dial"
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
	exchProto *exchangeProtocol

	subProtocol []SubProtocol

	findNodeUdpLayer *findnode.UDPLayer

	lock   sync.Mutex
	loopWG sync.WaitGroup

	dialScheduler *dial.Scheduler

	nodedb *admnode.NodeDB

	inboundConnHistory utils.InboundConnHeap

	// Channels
	quit                       chan struct{}
	handshakeValidateCh        chan *wrapPeerConnection
	exchangeProtocolValidateCh chan *wrapPeerConnection
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

	if srv.dialScheduler != nil {
		srv.dialScheduler.Stop()
	}
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
	srv.handshakeValidateCh = make(chan *wrapPeerConnection)
	srv.exchangeProtocolValidateCh = make(chan *wrapPeerConnection)

	if err := srv.initializeLocalNode(); err != nil {
		return err
	}

	srv.initializeDialScheduler()

	return nil
}

// initializeLocalNode
func (srv *Server) initializeLocalNode() error {
	// Create exchange protocol
	srv.exchProto = &exchangeProtocol{}

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

// initializeDialScheduler initialize the dial scheduler so that connect to nodes on TCP channel
func (srv *Server) initializeDialScheduler() {
	cfg := dial.Config{
		SelfID:        *srv.localnode.Node().ID(),
		PeerBlackList: srv.PeerBlackList,
		PeerWhiteList: srv.PeerWhiteList,
	}
	srv.dialScheduler = dial.New(cfg, nil, srv.AddConnection)
	srv.dialScheduler.Start()
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

	var peers = make(map[admnode.NodeID]*Peer)
	inboundConnCount := 0

running:
	for {
		select {
		case <-srv.quit:
			break running
		case wc := <-srv.handshakeValidateCh:
			wc.chError <- srv.handshakeValidate(wc, peers, inboundConnCount)
		case wc := <-srv.exchangeProtocolValidateCh:
			err := srv.addPeerValidate(wc, peers, inboundConnCount)
			if err != nil {
				wc.chError <- err
				continue
			}

			p := srv.startPeer(wc)
			peers[*p.peerConn.node.ID()] = p

			srv.log.Debug("Added peer", "id", p.peerConn.node.ID(), "addr", p.peerConn.node.IP(), "TCP", p.peerConn.node.TCP())

			// ToDo: add peer to dial scheduler
		}
	}

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
			srv.AddConnection(peerConn, dial.InboundConnection, nil)
			pendingInboundConnSlots <- struct{}{}
		}()
	}
}

func (srv *Server) AddConnection(peerConn net.Conn, connFlag dial.ConnectionFlag, remotePeerNode *admnode.GossipNode) error {
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

	if err = srv.getValidate(&wrapPeerConn, srv.handshakeValidateCh); err != nil {
		srv.log.Debug("Reject peer", "id", wrapPeerConn.node.ID(), "addr", wrapPeerConn.node.IP(), "err", err)
		return err
	}
	remoteExchProto, err := wrapPeerConn.doExchangeProtocol(srv.exchProto)
	if err != nil {
		srv.log.Debug("Reject peer", "id", wrapPeerConn.node.ID(), "addr", wrapPeerConn.node.IP(), "err", err)
		return err
	}

	wrapPeerConn.protocol = remoteExchProto
	if err = srv.getValidate(&wrapPeerConn, srv.exchangeProtocolValidateCh); err != nil {
		srv.log.Debug("Reject peer", "id", wrapPeerConn.node.ID(), "addr", wrapPeerConn.node.IP(), "err", err)
		return err
	}

	return nil
}

// handshakeValidate validates the connection
func (srv *Server) handshakeValidate(connection *wrapPeerConnection, peers map[admnode.NodeID]*Peer, inboundCount int) error {
	switch {
	case connection.connFlags&dial.InboundConnection != 0 && inboundCount > srv.MaxInboundConnections:
		return errTooManyInboundConnection
	case connection.node.ID() == srv.localnode.Node().ID():
		return errHandshakeWithSelf
	case peers[*connection.node.ID()] != nil:
		return errAlreadyConnected
	default:
		return nil
	}
}

func (srv *Server) addPeerValidate(connection *wrapPeerConnection, peers map[admnode.NodeID]*Peer, inboundCount int) error {
	if len(srv.ChainProtocol) == 0 || !srv.isMatchChainProtocols(connection.protocol) {
		return errNotMatchChainProtocol
	}

	return srv.handshakeValidate(connection, peers, inboundCount)
}

// getValidate send connection to the channel and returns the result
func (srv *Server) getValidate(connection *wrapPeerConnection, channel chan<- *wrapPeerConnection) error {
	select {
	case <-srv.quit:
		return errServerStopped
	case channel <- connection:
		return <-connection.chError
	}
}

// isMatchChainProtocols checks the protocol that matches with remote.
func (srv *Server) isMatchChainProtocols(remoteExchProto *exchangeProtocol) bool {
	matchCount := 0

	for _, remoteProtoID := range remoteExchProto.ProtocolIDs {
		for _, ownProtocol := range srv.ChainProtocol {
			if remoteProtoID == ownProtocol.ProtocolID {
				matchCount++
			}
		}
	}
	return matchCount > 0
}

// startPeer starts the peer module to communicate with remote peer.
func (srv *Server) startPeer(conn *wrapPeerConnection) *Peer {
	peer := newPeer(conn, srv.log, srv.subProtocol)
	go peer.start()
	return peer
}
