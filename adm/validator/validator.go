package validator

import (
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/params"
	// "github.com/adamnite/go-adamnite/log15"
)

// Validator creates blocks based on Adamnite DPOS consensus.
type Validator struct {
	witnessAddr common.Address
	adamnite    AdamniteImplInterface
	dposEngine  dpos.DPOS
	dposWorker  *dposWorker
	coinbase    common.Address
	mux         *event.TypeMux

	exitCh  chan struct{}
	startCh chan struct{}
	stopCh  chan struct{}
}

type Config struct {
	WitnessAddress common.Address
	Recommit       time.Duration
}

func New(adamnite AdamniteImplInterface, config *Config, chainConfig *params.ChainConfig, dpos dpos.DPOS, mux *event.TypeMux) *Validator {
	validator := &Validator{
		dposEngine:  dpos,
		adamnite:    adamnite,
		witnessAddr: config.WitnessAddress,

		mux: mux,

		exitCh:  make(chan struct{}),
		startCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}

	validator.dposWorker = newDposWorker(config, chainConfig, dpos, adamnite, mux, true)

	go validator.mainLoop()
	return validator
}

func (v *Validator) mainLoop() {
	for {
		select {
		case <-v.stopCh:
		case <-v.exitCh:
			return
		case <-v.startCh:
			v.dposWorker.start()
		}
	}
}

func (v *Validator) Start() {
	v.startCh <- struct{}{}
}

func (v *Validator) Stop() {
	v.dposWorker.stop()
}

func (v *Validator) Close() {
	v.dposWorker.close()
	close(v.exitCh)
}
func (v *Validator) Pending() (*types.Block, *statedb.StateDB) {
	return v.dposWorker.pending()
}

func (v *Validator) PendingBlock() *types.Block {
	return v.dposWorker.pendingBlock()
}

func (v *Validator) SetCoinbase(addr common.Address) {
	v.coinbase = addr
	v.dposWorker.setCoinbase(addr)
}

func (v *Validator) SetExtra(extra []byte) error {
	if uint64(len(extra)) > 32 {
		return fmt.Errorf("extra exceeds max length. %d > %v", len(extra), 32)
	}
	v.dposWorker.setExtra(extra)
	return nil
}
