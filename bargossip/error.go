package bargossip

import (
	"errors"
	"net"
	"os"
	"syscall"
)

var (
	errServerStopped   = errors.New("server stopped")
	errAlreadyListened = errors.New("already listened")
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

// IsTimeout checks whether the given error is a timeout.
func IsTimeout(err error) bool {
	timeoutErr, ok := err.(interface {
		Timeout() bool
	})
	return ok && timeoutErr.Timeout()
}
