// NAT-PMP(Port Mapping Protocol) is supported by Apple brand routers
// and open source routers like Tomato and DD-WRT.

package nat

import (
	"fmt"
	"net"
	"strings"
	"time"

	natpmp "github.com/jackpal/go-nat-pmp"
)

type pmp struct {
	gateway net.IP
	client  *natpmp.Client
}

func (n *pmp) String() string {
	return fmt.Sprintf("NAT-PMP(%v)", n.gateway)
}

func (n *pmp) ExternalIP() (net.IP, error) {
	response, err := n.client.GetExternalAddress()
	if err != nil {
		return nil, err
	}
	return response.ExternalIPAddress[:], nil
}

func (n *pmp) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
	if lifetime <= 0 {
		return fmt.Errorf("lifetime must not be <= 0")
	}

	_, err := n.client.AddPortMapping(strings.ToLower(protocol), intport, extport, int(lifetime/time.Second))
	return err
}

func (n *pmp) DeleteMapping(protocol string, extport, intport int) (err error) {
	_, err = n.client.AddPortMapping(strings.ToLower(protocol), intport, 0, 0)
	return err
}

func discoverPMP() Interface {
	gws := potentialGateways()
	found := make(chan *pmp, len(gws))
	for i := range gws {
		gw := gws[i]
		go func() {
			c := natpmp.NewClient(gw)
			if _, err := c.GetExternalAddress(); err != nil {
				found <- nil
			} else {
				found <- &pmp{gw, c}
			}
		}()
	}

	timeout := time.NewTimer(1 * time.Second)
	defer timeout.Stop()
	for range gws {
		select {
		case c := <-found:
			if c != nil {
				return c
			}
		case <-timeout.C:
			return nil
		}
	}
	return nil
}

func potentialGateways() (gws []net.IP) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		ifaddrs, err := iface.Addrs()
		if err != nil {
			return gws
		}
		for _, addr := range ifaddrs {
			if x, ok := addr.(*net.IPNet); ok {
				if lan_10.Contains(x.IP) || lan_176.Contains(x.IP) || lan_192.Contains(x.IP) {
					ip := x.IP.Mask(x.Mask).To4()
					if ip != nil {
						ip[3] = ip[3] | 0x01
						gws = append(gws, ip)
					}
				}
			}
		}
	}
	return gws
}
