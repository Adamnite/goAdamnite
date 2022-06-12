package adampro

import "github.com/adamnite/go-adamnite/core/types"

const (
	ADAM_PROTOCOL_DEMO = 22
)

const MaxMsgSize = 10 * 1024 * 1024 // 10 MB

var ProtocolVersions = []uint{
	ADAM_PROTOCOL_DEMO,
}

var ProtocolName = "adamnite"

var ProtocolLength = map[uint]uint64{
	ADAM_PROTOCOL_DEMO: 1,
}

const (
	NewBlockMsg       = 0x00
	TransactionsMsg   = 0x01
	GetAdmNodeDataMsg = 0x02
	AdmNodeDataMsg    = 0x03
)

type NewBlockPacket struct {
	Block *types.Block
}

func (*NewBlockPacket) Name() string { return "NewBlock" }
func (*NewBlockPacket) Kind() byte   { return NewBlockMsg }
