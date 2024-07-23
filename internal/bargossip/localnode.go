package bargossip

import (
	"bufio"
	"context"
	"time"

	p2p "github.com/adamnite/go-adamnite/internal/bargossip/pb"
	"github.com/adamnite/go-adamnite/log"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	maddr "github.com/multiformats/go-multiaddr"
)

type Config struct {
	ProtocolID     string
	BootstrapNodes []maddr.Multiaddr
}

type LocalNode struct {
	Config

	Server host.Host

	bootstrapNodes []host.Host

	*PingProtocol
}

func (n *LocalNode) StartDHTService() {
	log.Info("Adamnite bootstrap service starts")

	kademliaDHT, err := dht.New(context.Background(), n.Server, dht.Mode(dht.ModeAutoServer))
	if err != nil {
		log.Error("Failed to start DHT service", "err", err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	log.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(context.Background()); err != nil {
		log.Error("dht table creating occurs error", "err", err)
	}

	// Wait a bit to let bootstrapping finish (really bootstrap should block until it's ready, but that isn't the case yet.)
	time.Sleep(1 * time.Second)

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(context.Background(), routingDiscovery, "rendezvous")
	log.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	log.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(context.Background(), "rendezvous")
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == (n.Server).ID() {
			continue
		}
		log.Debug("Found peer:", peer)

		log.Debug("Connecting to:", peer)
		stream, err := n.Server.NewStream(context.Background(), peer.ID, protocol.ID(n.Config.ProtocolID))

		if err != nil {
			log.Warn("Connection failed:", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			log.Trace("RW", rw)
			// go writeData(rw)
			// go readData(rw)
		}

		log.Info("Connected to:", peer)
	}
}

func (n *LocalNode) StartPeerDiscovery() {
	log.Info("Adamnite bar-gossip peer discovery starts")

	bootstrapPeers := make([]peer.AddrInfo, len(n.Config.BootstrapNodes))
	n.bootstrapNodes = make([]host.Host, len(n.Config.BootstrapNodes))

	for i, addr := range n.Config.BootstrapNodes {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
		bootstrapPeers[i] = *peerinfo
	}

	kademliaDHT, err := dht.New(context.Background(), n.Server, dht.BootstrapPeers(bootstrapPeers...))
	if err != nil {
		log.Error("dht table creating occurs error", "err", err)
	}

	n.ConnectToBootstrap(bootstrapPeers)

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	log.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(context.Background()); err != nil {
		log.Error("dht table creating occurs error", "err", err)
	}

	// Wait a bit to let bootstrapping finish (really bootstrap should block until it's ready, but that isn't the case yet.)
	time.Sleep(1 * time.Second)

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(context.Background(), routingDiscovery, "rendezvous")
	log.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	log.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(context.Background(), "rendezvous")
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == n.Server.ID() {
			continue
		}
		log.Debug("Found peer:", peer)

		log.Debug("Connecting to:", peer)
		stream, err := n.Server.NewStream(context.Background(), peer.ID, protocol.ID(n.Config.ProtocolID))

		if err != nil {
			log.Warn("Connection failed:", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			log.Trace("RW", rw)
			// go writeData(rw)
			// go readData(rw)
		}

		log.Info("Connected to:", peer)
	}
}

func (n *LocalNode) ConnectToBootstrap(bootstrapPeers []peer.AddrInfo) {
	// Connecting to boostrap peers
	log.Info("Connecting to bootsrap nodes")

	for i := range bootstrapPeers {
		n.Server.Peerstore().AddAddrs(bootstrapPeers[i].ID, bootstrapPeers[i].Addrs, peerstore.PermanentAddrTTL)

		// Connect to bootsrap node
		if err := n.Server.Connect(context.Background(), bootstrapPeers[i]); err != nil {
			log.Error("Failed to connect bootstrap node", "bootstrap", bootstrapPeers[i].ID)
			continue
		}

		log.Debug("Connected to bootstrap node:", "ID", bootstrapPeers[i].ID)
		// n.Server.Ping(bootstrapPeers[i].ID)
	}
}

// Authenticate incoming p2p message
// message: a protobuf go data object
// data: common p2p message data
func (n *LocalNode) authenticateMessage(message proto.Message, data *p2p.MessageData) bool {
	// store a temp ref to signature and remove it from message data
	// sign is a string to allow easy reset to zero-value (empty string)
	sign := data.Sign
	data.Sign = nil

	// marshall data without the signature to protobufs3 binary format
	bin, err := proto.Marshal(message)
	if err != nil {
		log.Error("failed to marshal pb message", "err", err)
		return false
	}

	// restore sig in message data (for possible future use)
	data.Sign = sign

	// restore peer id binary format from base58 encoded node id data
	peerId, err := peer.Decode(data.NodeId)
	if err != nil {
		log.Error("Failed to decode node id from base58", "err", err)
		return false
	}

	// verify the data was authored by the signing peer identified by the public key
	// and signature included in the message
	return n.verifyData(bin, []byte(sign), peerId, data.NodePubKey)
}

// sign an outgoing p2p message payload
func (n *LocalNode) signProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return n.signData(data)
}

// sign binary data using the local node's private key
func (n *LocalNode) signData(data []byte) ([]byte, error) {
	key := n.Server.Peerstore().PrivKey(n.Server.ID())
	res, err := key.Sign(data)
	return res, err
}

// Verify incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// peerId: author peer id from the message payload
// pubKeyData: author public key from the message payload
func (n *LocalNode) verifyData(data []byte, signature []byte, peerId peer.ID, pubKeyData []byte) bool {
	key, err := crypto.UnmarshalPublicKey(pubKeyData)
	if err != nil {
		log.Error("Failed to extract key from message key data", "err", err)
		return false
	}

	// extract node id from the provided public key
	idFromKey, err := peer.IDFromPublicKey(key)

	if err != nil {
		log.Error("Failed to extract peer id from public key", "err", err)
		return false
	}

	// verify that message author node id matches the provided node public key
	if idFromKey != peerId {
		log.Error("Node id and provided public key mismatch", "err", err)
		return false
	}

	res, err := key.Verify(data, signature)
	if err != nil {
		log.Error("Error authenticating data", "err", err)
		return false
	}

	return res
}

// helper method - generate message data shared between all node's p2p protocols
// messageId: unique for requests, copied from request for responses
func (n *LocalNode) NewMessageData(messageId string, gossip bool) *p2p.MessageData {
	// Add protobuf bin data for message author public key
	// this is useful for authenticating  messages forwarded by a node authored by another node
	nodePubKey, err := crypto.MarshalPublicKey(n.Server.Peerstore().PubKey(n.Server.ID()))

	if err != nil {
		panic("Failed to get public key for sender from local peer store.")
	}

	return &p2p.MessageData{ClientVersion: ClientVersion,
		NodeId:     n.Server.ID().String(),
		NodePubKey: nodePubKey,
		Timestamp:  time.Now().Unix(),
		Id:         messageId,
		Gossip:     gossip}
}

// helper method - writes a protobuf go data object to a network stream
// data: reference of protobuf go data object to send (not the object itself)
// s: network stream to write the data to
func (n *LocalNode) sendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	s, err := n.Server.NewStream(context.Background(), id, p)
	if err != nil {
		log.Error("sendProtoMessage error", "err", err)
		return false
	}
	defer s.Close()

	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Error("sendProtoMessage write error", "err", err)
		s.Reset()
		return false
	}
	return true
}
