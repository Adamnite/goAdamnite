package findnode

import (
	"errors"
	"net"
	"os"
	"syscall"
)

var (
	errClosed      = errors.New("socket closed")
	errInvalid     = errors.New("invalid IP")
	errUnspecified = errors.New("zero address")
	errSpecial     = errors.New("special network")
	errLoopback    = errors.New("loopback address from non-loopback host")
	errLAN         = errors.New("LAN address from WAN host")
	errTimeout     = errors.New("send call timeout")
	errClockWarp   = errors.New("reply deadline too far in the future")
)

// IsTemporaryError checks whether the given error should be considered temporary.
func IsTemporaryError(err error) bool {
	temporaryErr, ok := err.(interface {
		Temporary() bool
	})
	return ok && temporaryErr.Temporary() || IsPacketTooBig(err)
}

func IsPacketTooBig(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if scErr, ok := opErr.Err.(*os.SyscallError); ok {
			return scErr.Err == syscall.Errno(10040)
		}
		return opErr.Err == syscall.Errno(10040)
	}
	return false
}
