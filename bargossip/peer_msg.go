package bargossip

import (
	"time"

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
