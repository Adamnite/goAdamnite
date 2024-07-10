package node

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/log"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"

	"github.com/adamnite/go-adamnite/internal/bargossip"
	p2p "github.com/adamnite/go-adamnite/internal/bargossip/pb"
	ggio "github.com/gogo/protobuf/io"
)

// Node is a container on which services can be registered.
type Node struct {
	// config describes the configuration of the Adamnite node.
	config Config

	server *host.Host

	prvKey crypto.PrivKey

	// Channels
	stop     chan struct{} // Channel to wait for termination notifications
	onceStop sync.Once

	// Protocols
	*PingProtocol
}

func New(config *Config) (*Node, error) {
	var r io.Reader

	r = rand.Reader

	// check if the prvkey exists on the datadir

	prvKey := checkNodeKey(config)

	if prvKey == nil {
		var err error

		prvKey, _, err = crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
		if err != nil {
			log.Error("Key generation failed")
			return nil, err
		}

		saveNodeKey(config, prvKey)
	}

	// save the prvkey

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2PPort))

	server, err := libp2p.New(libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(prvKey))

	if err != nil {
		log.Error("Bargossip server hosting error!")
	}

	return &Node{
		config: *config,
		prvKey: prvKey,
		server: &server,
		stop:   make(chan struct{}),
	}, nil
}

func (n *Node) Wait() {
	<-n.stop
}

func (n *Node) Start() {
	log.Info("Adamnite DataDir path", "PATH", n.config.DataDir)
	log.Info("Adamnite external address", "Addr", (*n.server).Addrs())
	log.Info("Adamnite host ID", "ID", (*n.server).ID())

}

func (n *Node) Close() {
	closeOnce := func() {
		close(n.stop)
	}

	n.onceStop.Do(closeOnce)
}

// Authenticate incoming p2p message
// message: a protobuf go data object
// data: common p2p message data
func (n *Node) authenticateMessage(message proto.Message, data *p2p.MessageData) bool {
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
func (n *Node) signProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return n.signData(data)
}

// sign binary data using the local node's private key
func (n *Node) signData(data []byte) ([]byte, error) {
	key := (*n.server).Peerstore().PrivKey((*n.server).ID())
	res, err := key.Sign(data)
	return res, err
}

// Verify incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// peerId: author peer id from the message payload
// pubKeyData: author public key from the message payload
func (n *Node) verifyData(data []byte, signature []byte, peerId peer.ID, pubKeyData []byte) bool {
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
func (n *Node) NewMessageData(messageId string, gossip bool) *p2p.MessageData {
	// Add protobuf bin data for message author public key
	// this is useful for authenticating  messages forwarded by a node authored by another node
	nodePubKey, err := crypto.MarshalPublicKey((*n.server).Peerstore().PubKey((*n.server).ID()))

	if err != nil {
		panic("Failed to get public key for sender from local peer store.")
	}

	return &p2p.MessageData{ClientVersion: bargossip.ClientVersion,
		NodeId:     (*n.server).ID().String(),
		NodePubKey: nodePubKey,
		Timestamp:  time.Now().Unix(),
		Id:         messageId,
		Gossip:     gossip}
}

// helper method - writes a protobuf go data object to a network stream
// data: reference of protobuf go data object to send (not the object itself)
// s: network stream to write the data to
func (n *Node) sendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	s, err := (*n.server).NewStream(context.Background(), id, p)
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
