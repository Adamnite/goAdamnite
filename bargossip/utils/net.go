package utils

import (
	"errors"
	"net"
)

var lan4, lan6, special4, special6 IPNetList

// NetAddrToIP gets the IP address of the addr parameter.
func NetAddrToIP(addr net.Addr) net.IP {
	switch a := addr.(type) {
	case *net.IPAddr:
		return a.IP
	case *net.TCPAddr:
		return a.IP
	case *net.UDPAddr:
		return a.IP
	default:
		return nil
	}
}

// IsLAN reports whether an IP is a local network address.
func IsLAN(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		return lan4.Contains(v4)
	}
	return lan6.Contains(ip)
}

// IsSpecialNetwork reports whether an IP is located in a special-use network range
// This includes broadcast, multicast and documentation addresses.
func IsSpecialNetwork(ip net.IP) bool {
	if ip.IsMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		return special4.Contains(v4)
	}
	return special6.Contains(ip)
}

var (
	errInvalid     = errors.New("invalid IP")
	errUnspecified = errors.New("zero address")
	errSpecial     = errors.New("special network")
	errLoopback    = errors.New("loopback address from non-loopback host")
	errLAN         = errors.New("LAN address from WAN host")
)

// CheckRelayIP reports whether an IP relayed from the given sender IP
// is a valid connection target.
//
// There are four rules:
//   - Special network addresses are never valid.
//   - Loopback addresses are OK if relayed by a loopback host.
//   - LAN addresses are OK if relayed by a LAN host.
//   - All other addresses are always acceptable.
func CheckRelayIP(sender, addr net.IP) error {
	if len(addr) != net.IPv4len && len(addr) != net.IPv6len {
		return errInvalid
	}
	if addr.IsUnspecified() {
		return errUnspecified
	}
	if IsSpecialNetwork(addr) {
		return errSpecial
	}
	if addr.IsLoopback() && !sender.IsLoopback() {
		return errLoopback
	}
	if IsLAN(addr) && !IsLAN(sender) {
		return errLAN
	}
	return nil
}
