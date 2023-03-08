package utils

// import "github.com/cakturk/go-netstat/netstat"

// func FindUDPPortListeners(port int) (processes map[int]string, err error) {
// 	processes = make(map[int]string)

// 	acceptFn := func(se *netstat.SockTabEntry) bool {
// 		return int(se.LocalAddr.Port) == port
// 	}
// 	socketEntries, err := netstat.UDPSocks(acceptFn)
// 	if err != nil {
// 		return
// 	}

// 	for _, listener := range socketEntries {
// 		processes[listener.Process.Pid] = listener.Process.Name
// 	}

// 	socket6Entries, err := netstat.UDP6Socks(acceptFn)
// 	if err != nil {
// 		return
// 	}

// 	for _, listener := range socket6Entries {
// 		processes[listener.Process.Pid] = listener.Process.Name
// 	}

// 	return
// }
