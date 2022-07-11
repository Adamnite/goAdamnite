// ExtIP assumes that the local machine is reachable on the given
// external IP address, and that any required ports were mapped manually.
// Mapping operations will not return an error but won't actually do anything.

package nat

import (
	"fmt"
	"net"
	"time"
)

type ExtIP net.IP

func (n ExtIP) ExternalIP() (net.IP, error) { return net.IP(n), nil }
func (n ExtIP) String() string              { return fmt.Sprintf("ExtIP(%v)", net.IP(n)) }

// These do nothing.

func (ExtIP) AddMapping(string, int, int, string, time.Duration) error { return nil }
func (ExtIP) DeleteMapping(string, int, int) error                     { return nil }
