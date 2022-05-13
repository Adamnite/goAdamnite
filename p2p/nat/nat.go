package nat

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/log15"
	natpmp "github.com/jackpal/go-nat-pmp"
)

type Interface interface {
	AddMapping(protocol string, extport int, intport int, name string, lifetime time.Duration) error
	DeleteMapping(protocol string, extport int, intport int) error

	ExternalIP() (net.IP, error)
	String() string
}

func Parse(spec string) (Interface, error) {
	var (
		parts     = strings.SplitN(spec, ":", 2)
		mechanism = strings.ToLower(parts[0])
		ip        net.IP
	)

	if len(parts) > 1 {
		ip = net.ParseIP(parts[1])
		if ip == nil {
			return nil, errors.New("invalid IP address")
		}
	}

	switch mechanism {
	case "", "none", "off":
		return nil, nil
	case "any", "auto", "on":
		return Any(), nil
	case "extip", "ip":
		if ip == nil {
			return nil, errors.New("missing IP address")
		}
		return ExtIP(ip), nil
	case "upnp":
		return UPnP(), nil
	case "pmp", "natpmp", "nat-pmp":
		return PMP(ip), nil
	default:
		return nil, fmt.Errorf("unknown mechanism %q", parts[0])
	}
}

type ExtIP net.IP

func (n ExtIP) ExternalIP() (net.IP, error) { return net.IP(n), nil }
func (n ExtIP) String() string              { return fmt.Sprintf("ExtIP(%v)", net.IP(n)) }

func (ExtIP) AddMapping(string, int, int, string, time.Duration) error { return nil }
func (ExtIP) DeleteMapping(string, int, int) error                     { return nil }

func Any() Interface {
	return startautodiscovery("UPnP or NAT-PMP", func() Interface {
		found := make(chan Interface, 2)
		go func() {
			found <- discoverUPnP()
		}()
		go func() {
			found <- discoverPMP()
		}()

		for i := 0; i < cap(found); i++ {
			if c := <-found; c != nil {
				return c
			}
		}
		return nil
	})
}

func UPnP() Interface {
	return startautodiscovery("UPnP", discoverUPnP)
}

func PMP(gateway net.IP) Interface {
	if gateway != nil {
		return &pmp{gateway: gateway, client: natpmp.NewClient(gateway)}
	}
	return startautodiscovery("NAT-PMP", discoverPMP)
}

type autodiscovery struct {
	what string
	once sync.Once
	doit func() Interface

	mu    sync.Mutex
	found Interface
}

func startautodiscovery(what string, doit func() Interface) Interface {
	return &autodiscovery{what: what, doit: doit}
}

func (ad *autodiscovery) AddMapping(protocol string, extport int, intport int, name string, lifetime time.Duration) error {
	if err := ad.wait(); err != nil {
		return err
	}
	return ad.found.AddMapping(protocol, extport, intport, name, lifetime)
}

func (ad *autodiscovery) DeleteMapping(protocol string, extport, intport int) error {
	if err := ad.wait(); err != nil {
		return err
	}
	return ad.found.DeleteMapping(protocol, extport, intport)
}

func (ad *autodiscovery) ExternalIP() (net.IP, error) {
	if err := ad.wait(); err != nil {
		return nil, err
	}
	return ad.found.ExternalIP()
}

func (ad *autodiscovery) String() string {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	if ad.found == nil {
		return ad.what
	}
	return ad.found.String()
}

func (ad *autodiscovery) wait() error {
	ad.once.Do(func() {
		ad.mu.Lock()
		ad.found = ad.doit()
		ad.mu.Unlock()
	})
	if ad.found == nil {
		return fmt.Errorf("no %s router discovered", ad.what)
	}
	return nil
}

const (
	mapTimeout = 10 * time.Minute
)

// Map adds a port mapping on m and keeps it alive until c is closed.
// This function is typically invoked in its own goroutine.
func Map(m Interface, c <-chan struct{}, protocol string, extport, intport int, name string) {
	log := log15.New("proto", protocol, "extport", extport, "intport", intport, "interface", m)
	refresh := time.NewTimer(mapTimeout)
	defer func() {
		refresh.Stop()
		log.Debug("Deleting port mapping")
		m.DeleteMapping(protocol, extport, intport)
	}()
	if err := m.AddMapping(protocol, extport, intport, name, mapTimeout); err != nil {
		log.Debug("Couldn't add port mapping", "err", err)
	} else {
		log.Info("Mapped network port")
	}
	for {
		select {
		case _, ok := <-c:
			if !ok {
				return
			}
		case <-refresh.C:
			log.Trace("Refreshing port mapping")
			if err := m.AddMapping(protocol, extport, intport, name, mapTimeout); err != nil {
				log.Debug("Couldn't add port mapping", "err", err)
			}
			refresh.Reset(mapTimeout)
		}
	}
}
