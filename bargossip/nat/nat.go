// NAT(Network Address Translation) is a method of mapping
// an IP address space into another by modifying network
// address information in the IP header of packets while
// they are in transit across a traffic routing device(Wifi, Router).
//
// go-adamnite Nat library include nat-pmp and nat-upnp for UDP hole punching.
//
// UDP hole punching is a utilsly used technique employed in NAT applications
// for maintaining UDP(User Datagram Protocol) packet streams that traverse the NAT.
// nat.Interface can map local ports to ports accessible from the Internet.

package nat

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/adamnite/go-adamnite/log15"
	natpmp "github.com/jackpal/go-nat-pmp"
)

type Interface interface {
	AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error
	DeleteMapping(protocol string, extport, intport int) error
	ExternalIP() (net.IP, error)
	String() string
}

// Parse parses a NAT interface description.
func Parse(spec string) (Interface, error) {
	var (
		parts = strings.SplitN(spec, ":", 2)
		mech  = strings.ToLower(parts[0])
		ip    net.IP
	)
	if len(parts) > 1 {
		ip = net.ParseIP(parts[1])
		if ip == nil {
			return nil, errors.New("invalid IP address")
		}
	}
	switch mech {
	case "", "none":
		return nil, nil
	case "any":
		return Any(), nil
	case "extip":
		if ip == nil {
			return nil, errors.New("missing IP address")
		}
		return ExtIP(ip), nil
	case "upnp":
		return UPnP(), nil
	case "pmp":
		return PMP(ip), nil
	default:
		return nil, fmt.Errorf("unknown mechanism %q", parts[0])
	}
}

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

// Any returns a port mapper that tries to discover any supported
// mechanism on the local network.
func Any() Interface {
	return startautodisc("UPnP or NAT-PMP", func() Interface {
		found := make(chan Interface, 2)
		go func() { found <- discoverUPnP() }()
		go func() { found <- discoverPMP() }()
		for i := 0; i < cap(found); i++ {
			if c := <-found; c != nil {
				return c
			}
		}
		return nil
	})
}

// UPnP(Universal Plug and Play) returns a port mapper that uses UPnP. It will attempt to
// discover the address of your router using UDP broadcasts.
func UPnP() Interface {
	return startautodisc("UPnP", discoverUPnP)
}

// PMP(Port Mapping Protocol) returns a port mapper that uses NAT-PMP. The provided gateway
// address should be the IP of your router. If the given gateway address is nil,
// PMP will attempt to auto-discover the router.
func PMP(gateway net.IP) Interface {
	if gateway != nil {
		return &pmp{gateway: gateway, client: natpmp.NewClient(gateway)}
	}
	return startautodisc("NAT-PMP", discoverPMP)
}
