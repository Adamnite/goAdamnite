package bargossip

import (
	"io"
	"time"
)

// Msg defines the structure of a p2p message
type Msg struct {
	Code       uint64
	Size       uint32
	Payload    io.Reader
	ReceivedAt time.Time
}
