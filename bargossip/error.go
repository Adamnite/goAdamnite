package bargossip

import (
	"errors"
)

var (
	errServerStopped            = errors.New("server stopped")
	errAlreadyListened          = errors.New("already listened")
	errTooManyInboundConnection = errors.New("too many inbound connection")
	errHandshakeWithSelf        = errors.New("handshake with own")
	errAlreadyConnected         = errors.New("already connected")
	errTooLargeMessage          = errors.New("message body size too large")
	errNotMatchChainProtocol    = errors.New("not match with chain protocol")
	errShuttingDown             = errors.New("shutting down")
	errProtocolReturned         = errors.New("protocol returned")
)

// IsTimeout checks whether the given error is a timeout.
func IsTimeout(err error) bool {
	timeoutErr, ok := err.(interface {
		Timeout() bool
	})
	return ok && timeoutErr.Timeout()
}
