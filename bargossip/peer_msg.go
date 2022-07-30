package bargossip

import (
	"io"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/event"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	pingMsg             = 0x00
	pongMsg             = 0x01
	exchangeProtocolMsg = 0x02
)

// Msg defines the structure of a p2p message
type Msg struct {
	Code       uint64
	Size       uint32
	Payload    []byte
	ReceivedAt time.Time
}

type exchangeProtocol struct {
	ProtocolIDs []uint
}

func (msg Msg) Decode(val interface{}) error {
	if err := msgpack.Unmarshal(msg.Payload, val); err != nil {
		return err
	}
	return nil
}

func Send(w MsgWriter, msgcode uint64, data interface{}) error {
	r, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	return w.WriteMsg(Msg{Code: msgcode, Size: uint32(len(r)), Payload: r})
}

func SendItems(w MsgWriter, msgcode uint64, elems ...interface{}) error {
	return Send(w, msgcode, elems)
}

// msgEventer wraps a MsgReadWriter and sends events whenever a message is sent
// or received
type msgEventer struct {
	MsgReadWriter

	feed          *event.Feed
	peerID        admnode.NodeID
	ProtocolID    uint
	localAddress  string
	remoteAddress string
}

// newMsgEventer returns a msgEventer which sends message events to the given
// feed
func newMsgEventer(rw MsgReadWriter, feed *event.Feed, peerID admnode.NodeID, proto uint, remote, local string) *msgEventer {
	return &msgEventer{
		MsgReadWriter: rw,
		feed:          feed,
		peerID:        peerID,
		ProtocolID:    proto,
		remoteAddress: remote,
		localAddress:  local,
	}
}

// ReadMsg reads a message from the underlying MsgReadWriter and emits a
// "message received" event
func (ev *msgEventer) ReadMsg() (Msg, error) {
	msg, err := ev.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	ev.feed.Send(&PeerEvent{
		Type:          PeerEventTypeMsgRecv,
		Peer:          ev.peerID,
		ProtocolID:    ev.ProtocolID,
		MsgCode:       &msg.Code,
		MsgSize:       &msg.Size,
		LocalAddress:  ev.localAddress,
		RemoteAddress: ev.remoteAddress,
	})
	return msg, nil
}

// WriteMsg writes a message to the underlying MsgReadWriter and emits a
// "message sent" event
func (ev *msgEventer) WriteMsg(msg Msg) error {
	err := ev.MsgReadWriter.WriteMsg(msg)
	if err != nil {
		return err
	}
	ev.feed.Send(&PeerEvent{
		Type:          PeerEventTypeMsgSend,
		Peer:          ev.peerID,
		ProtocolID:    ev.ProtocolID,
		MsgCode:       &msg.Code,
		MsgSize:       &msg.Size,
		LocalAddress:  ev.localAddress,
		RemoteAddress: ev.remoteAddress,
	})
	return nil
}

// Close closes the underlying MsgReadWriter if it implements the io.Closer
// interface
func (ev *msgEventer) Close() error {
	if v, ok := ev.MsgReadWriter.(io.Closer); ok {
		return v.Close()
	}
	return nil
}
