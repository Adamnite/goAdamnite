package validator

import (
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/event"
	"github.com/adamnite/go-adamnite/params"
)

// Validator creates blocks based on Adamnite DPOS consensus.
type Validator struct {
	witnessAddr common.Address
	adamnite    AdamniteImplInterface
	dposEngine  dpos.DPOS
	dposWorker  *dposWorker

	mux *event.TypeMux

	exitCh  chan struct{}
	startCh chan struct{}
	stopCh  chan struct{}
}

type Config struct {
	WitnessAddress common.Address
}

var DefaultDemoConfig = Config{}

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

	validator.dposWorker = newDposWorker(config, chainConfig, dpos, adamnite, mux)

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
