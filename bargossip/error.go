package bargossip

import (
	"errors"
)

var (
	errServerStopped   = errors.New("server stopped")
	errAlreadyListened = errors.New("already listened")
)

// IsTimeout checks whether the given error is a timeout.
func IsTimeout(err error) bool {
	timeoutErr, ok := err.(interface {
		Timeout() bool
	})
	return ok && timeoutErr.Timeout()
}
