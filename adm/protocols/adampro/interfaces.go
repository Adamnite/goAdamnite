package adampro

import (
	"time"

	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
)

type TxPool interface {
	Get(hash bytes.Hash) *types.Transaction
}

type AdamniteHandlerInterface interface {
	Chain() *blockchain.Blockchain
	TxPool() TxPool
	RunPeer(peer *Peer, handler Handler) error
	Handle(peer *Peer, packet Packet) error
}

type Decoder interface {
	Decode(val interface{}) error
	Time() time.Time
}

type Packet interface {
	Name() string
	Kind() byte
}
