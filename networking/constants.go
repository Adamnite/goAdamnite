package networking

import "fmt"

var (
	err_preexistingConnection   = fmt.Errorf("contact already has active connection")
	err_outboundCapacityReached = fmt.Errorf("currently at capacity for outbound connections")
)
