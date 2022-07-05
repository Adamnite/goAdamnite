package nat

import (
	"net"
	"time"
)

const (
	mapTimeout = 10 * time.Minute
)

var (
	// LAN IP ranges
	_, lan_10, _  = net.ParseCIDR("10.0.0.0/8")     // 10.0.0.0 ~ 10.255.255.255 | Total Host 16,777,216 | Netmask 255.0.0.0
	_, lan_176, _ = net.ParseCIDR("172.16.0.0/12")  // 172.16.0.0 ~ 172.31.255.255 | Total Host 1,048,576 | Netmask 255.240.0.0
	_, lan_192, _ = net.ParseCIDR("192.168.0.0/16") // 192.168.0.0 ~ 192.168.255.255 | Total Host 65,536 | Netmask 255.255.0.0
)
