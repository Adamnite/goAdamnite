package dpos

import (
	"sync"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/log15"
)

type Config struct {
	Log log15.Logger `toml:"-"`
}

type AdamniteDPOS struct {
	config Config

	lock      sync.Mutex
	closeOnce sync.Once
}

func New(config Config) *AdamniteDPOS {
	if config.Log == nil {
		config.Log = log15.Root()
	}

	dpos := &AdamniteDPOS{
		config: config,
	}

	return dpos
}

func (adpos *AdamniteDPOS) Close() error {
	adpos.closeOnce.Do(func() {

	})
	return nil
}

func (adpos *AdamniteDPOS) Witness(header *types.BlockHeader) (common.Address, error) {
	return header.Witness, nil
}

func (adpos *AdamniteDPOS) VerifyHeader(header *types.BlockHeader, chain ChainHeaderReader) error {
	return nil
}

func (adpos *AdamniteDPOS) GetRoundNumber() uint64 {
	return 0
}
